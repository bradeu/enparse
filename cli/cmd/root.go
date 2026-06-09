package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "enparse",
	Version: version,
	Short:   "Decentralized secrets manager for small teams",
	Long: `enparse — share secrets securely using blockchain identity.

Each developer gets a unique CLI-generated identity. Secrets are encrypted
per-member using NaCl box and stored on-chain. No central server needed.

Quick start:
  enparse init                          # generate your identity
  enparse register                      # register on-chain (you pay gas)
  enparse project create myapp          # create a project (owner only)
  enparse project add myapp 0x<addr>    # add a teammate (owner only)
  enparse set myapp DB_URL=postgres://  # store a secret (owner only)
  enparse run myapp -- go run .         # inject secrets and run`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
