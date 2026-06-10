package cmd

import (
	"context"
	"fmt"

	"github.com/enparse/cli/internal/chain"
	"github.com/enparse/cli/internal/config"
	"github.com/enparse/cli/internal/identity"
	"github.com/enparse/cli/internal/output"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy IdentityRegistry and ProjectVault contracts on-chain",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := identity.Load()
		if err != nil {
			return fmt.Errorf("no identity found — run 'enparse init' first")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg.RPCURL == "" {
			return fmt.Errorf("rpc_url not configured — run: enparse config set rpc_url <url>")
		}

		output.Step("Deployer: %s", id.EthAddress)
		output.Step("RPC: %s", cfg.RPCURL)
		output.Info("Deploying IdentityRegistry and ProjectVault...")

		ctx := context.Background()
		result, err := chain.DeployContracts(ctx, cfg.RPCURL, id.EthPrivKey)
		if err != nil {
			return fmt.Errorf("deploy: %w", err)
		}

		output.Success("IdentityRegistry: %s", result.IdentityRegistryAddr)
		output.Success("ProjectVault:     %s", result.ProjectVaultAddr)
		output.Field("chain ID", result.ChainID.String())

		cfg.IdentityRegistryAddr = result.IdentityRegistryAddr
		cfg.ProjectVaultAddr = result.ProjectVaultAddr
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		output.Success("Saved to ~/.enparse/config.json")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
