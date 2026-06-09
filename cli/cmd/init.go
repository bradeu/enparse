package cmd

import (
	"fmt"

	"github.com/enparse/cli/internal/identity"
	"github.com/enparse/cli/internal/output"
	"github.com/spf13/cobra"
)

var forceInit bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a new identity (NaCl + Ethereum keypair)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := identity.Load(); err == nil && !forceInit {
			output.Info("Identity already exists at %s", identity.Path())
			output.Hint("use --force to overwrite")
			return nil
		}

		id, err := identity.Generate()
		if err != nil {
			return fmt.Errorf("generate identity: %w", err)
		}
		if err := id.Save(); err != nil {
			return fmt.Errorf("save identity: %w", err)
		}

		output.Success("Identity created")
		output.Address("address", id.EthAddress)
		output.Field("public key", id.NaclPubKey)
		output.Field("saved to", identity.Path())
		output.Info("")
		output.Hint("run 'enparse register' to register on-chain")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Overwrite existing identity")
	rootCmd.AddCommand(initCmd)
}
