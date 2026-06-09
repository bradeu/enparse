package cmd

import (
	"fmt"

	"github.com/enparse/cli/internal/config"
	"github.com/enparse/cli/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage enparse configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		cfg, err := config.Load()
		if err != nil {
			cfg = &config.Config{}
		}

		switch key {
		case "rpc_url":
			cfg.RPCURL = value
		case "identity_registry_addr":
			cfg.IdentityRegistryAddr = value
		case "project_vault_addr":
			cfg.ProjectVaultAddr = value
		default:
			return fmt.Errorf("unknown key %q — valid keys: rpc_url, identity_registry_addr, project_vault_addr", key)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		output.Success("Set %s = %s", key, value)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			output.Info("No config found.")
			output.Hint("run 'enparse config set rpc_url <url>'")
			return nil
		}
		output.Field("rpc_url", or(cfg.RPCURL, "(not set)"))
		output.Field("identity_registry_addr", or(cfg.IdentityRegistryAddr, "(not set)"))
		output.Field("project_vault_addr", or(cfg.ProjectVaultAddr, "(not set)"))
		output.Field("config file", config.Path())
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}
