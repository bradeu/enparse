package cmd

import (
	"fmt"

	"github.com/enparse/cli/internal/output"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register your NaCl public key on-chain",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _, client, err := loadAll()
		if err != nil {
			return err
		}

		registered, err := client.IsRegistered(common.HexToAddress(id.EthAddress))
		if err == nil && registered {
			output.Success("Already registered on-chain")
			return nil
		}

		pubKey, err := id.NaclPubKeyBytes()
		if err != nil {
			return err
		}

		output.Step("Registering %s", id.EthAddress)
		if err := client.RegisterIdentity(id.EthPrivKey, *pubKey); err != nil {
			return fmt.Errorf("register: %w", err)
		}
		output.Success("Registered on-chain")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
