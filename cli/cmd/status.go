package cmd

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/enparse/cli/internal/chain"
	"github.com/enparse/cli/internal/config"
	"github.com/enparse/cli/internal/identity"
	"github.com/enparse/cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show identity and chain connection status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := identity.Load()
		if err != nil {
			output.Info("No identity found.")
			output.Hint("run 'enparse init'")
			return nil
		}

		output.Header("Identity")
		output.Address("address", id.EthAddress)
		output.Field("public key", id.NaclPubKey[:16]+"...")
		output.Field("file", identity.Path())

		cfg, err := config.Load()
		if err != nil {
			output.Info("")
			output.Info("No config found.")
			output.Hint("run 'enparse config set rpc_url <url>'")
			return nil
		}

		output.Info("")
		output.Header("Config")
		output.Field("rpc_url", or(cfg.RPCURL, "(not set)"))
		output.Field("identity_registry", or(cfg.IdentityRegistryAddr, "(not set)"))
		output.Field("project_vault", or(cfg.ProjectVaultAddr, "(not set)"))
		output.Field("config file", config.Path())
		output.Info("")

		if err := cfg.Validate(); err != nil {
			output.Fail("Config incomplete: %v", err)
			return nil
		}

		client, err := chain.New(cfg.RPCURL, cfg.IdentityRegistryAddr, cfg.ProjectVaultAddr)
		if err != nil {
			output.Fail("Cannot connect to chain: %v", err)
			return nil
		}

		output.Header("Chain")
		registered, err := client.IsRegistered(common.HexToAddress(id.EthAddress))
		if err != nil {
			output.Fail("Registration check failed: %v", err)
			return nil
		}

		if registered {
			output.Success("registered")
		} else {
			output.Info("  not registered")
			output.Hint("run 'enparse register'")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
