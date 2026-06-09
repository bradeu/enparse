package cmd

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/enparse/cli/internal/output"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project (you become the owner)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		id, _, client, err := loadAll()
		if err != nil {
			return err
		}
		output.Step("Creating project %q", name)
		if err := client.CreateProject(id.EthPrivKey, name); err != nil {
			return fmt.Errorf("create project: %w", err)
		}
		output.Success("Project %q created — you are the owner", name)
		output.Hint("share your address with teammates: %s", id.EthAddress)
		return nil
	},
}

var projectAddCmd = &cobra.Command{
	Use:   "add <project> <address>",
	Short: "Add a registered member to a project (owner only)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName, addrHex := args[0], args[1]
		if !common.IsHexAddress(addrHex) {
			return fmt.Errorf("invalid Ethereum address: %s", addrHex)
		}

		id, _, client, err := loadAll()
		if err != nil {
			return err
		}

		member := common.HexToAddress(addrHex)
		output.Step("Adding %s to %q", member.Hex(), projectName)
		if err := client.AddMember(id.EthPrivKey, projectName, member); err != nil {
			return fmt.Errorf("add member: %w", err)
		}
		output.Success("Member added")

		// Re-encrypt all existing secrets for the updated member list.
		existing, err := pullSecrets(id, client, projectName)
		if err != nil {
			return fmt.Errorf("pull secrets for re-encryption: %w", err)
		}
		if len(existing) > 0 {
			members, err := client.GetMembers(projectName)
			if err != nil {
				return fmt.Errorf("get members: %w", err)
			}
			output.Step("Re-encrypting %d secret(s) for new member", len(existing))
			for k, v := range existing {
				if err := storeSecret(id, client, projectName, k, v, members); err != nil {
					return fmt.Errorf("re-encrypt %q: %w", k, err)
				}
			}
			output.Success("All secrets re-encrypted for %d member(s)", len(members))
		}
		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects you belong to",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _, client, err := loadAll()
		if err != nil {
			return err
		}

		addr := common.HexToAddress(id.EthAddress)
		projectIDs, err := client.GetMemberProjects(addr)
		if err != nil {
			return fmt.Errorf("get projects: %w", err)
		}

		if len(projectIDs) == 0 {
			output.Info("No projects found.")
			return nil
		}

		fmt.Printf("%-24s  %-42s  %s\n", "NAME", "OWNER", "ID (prefix)")
		fmt.Printf("%-24s  %-42s  %s\n", "----", "-----", "----------")
		for _, pid := range projectIDs {
			name, _ := client.GetProjectName(pid)
			owner, _ := client.GetOwnerByID(pid)
			if name == "" {
				name = "(unknown)"
			}
			fmt.Printf("%-24s  %-42s  0x%x...\n", name, owner.Hex(), pid[:4])
		}
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectAddCmd)
	projectCmd.AddCommand(projectListCmd)
	rootCmd.AddCommand(projectCmd)
}
