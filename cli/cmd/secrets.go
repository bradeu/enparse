package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	enpcrypto "github.com/enparse/cli/internal/crypto"
	"github.com/enparse/cli/internal/chain"
	"github.com/enparse/cli/internal/config"
	"github.com/enparse/cli/internal/identity"
	"github.com/enparse/cli/internal/output"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <project> KEY=value",
	Short: "Encrypt and store a secret for all project members (owner only)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName, kv := args[0], args[1]

		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format — expected KEY=value, got %q", kv)
		}
		key, value := parts[0], parts[1]

		id, _, client, err := loadAll()
		if err != nil {
			return err
		}

		members, err := client.GetMembers(projectName)
		if err != nil {
			return fmt.Errorf("get members for %q: %w", projectName, err)
		}
		if len(members) == 0 {
			return fmt.Errorf("project %q not found or has no members", projectName)
		}

		output.Step("Storing %q for %d member(s)", key, len(members))
		if err := storeSecret(id, client, projectName, key, value, members); err != nil {
			return fmt.Errorf("set secret: %w", err)
		}
		output.Success("Secret %q stored", key)
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <project> KEY",
	Short: "Fetch and decrypt a single secret",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName, key := args[0], args[1]

		id, _, client, err := loadAll()
		if err != nil {
			return err
		}

		value, err := fetchAndDecrypt(id, client, projectName, key)
		if err != nil {
			return err
		}

		fmt.Println(value)
		return nil
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull <project>",
	Short: "Fetch all secrets and cache them locally",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		id, _, client, err := loadAll()
		if err != nil {
			return err
		}

		output.Step("Pulling secrets for %q", projectName)
		secrets, err := pullSecrets(id, client, projectName)
		if err != nil {
			return err
		}

		if err := saveCache(projectName, secrets); err != nil {
			return fmt.Errorf("save cache: %w", err)
		}

		output.Success("Pulled %d secret(s) → %s", len(secrets), cachePath(projectName))
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run <project> -- <command> [args...]",
	Short: "Run a command with secrets injected as environment variables",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		cmdArgs := args[1:]

		id, _, client, err := loadAll()
		if err != nil {
			return err
		}

		secrets, err := pullSecrets(id, client, projectName)
		if err != nil {
			return fmt.Errorf("pull secrets: %w", err)
		}

		env := os.Environ()
		for k, v := range secrets {
			env = append(env, k+"="+v)
		}

		binary, err := exec.LookPath(cmdArgs[0])
		if err != nil {
			return fmt.Errorf("command not found: %s", cmdArgs[0])
		}

		return syscall.Exec(binary, cmdArgs, env)
	},
}

func storeSecret(
	id *identity.Identity,
	client *chain.Client,
	projectName, key, value string,
	members []common.Address,
) error {
	ownerPriv, err := id.NaclPrivKeyBytes()
	if err != nil {
		return err
	}
	memberAddrs := make([]common.Address, 0, len(members))
	encryptedValues := make([][]byte, 0, len(members))
	for _, member := range members {
		pubKey, err := client.GetPublicKey(member)
		if err != nil {
			return fmt.Errorf("get public key for %s: %w", member.Hex(), err)
		}
		encrypted, err := enpcrypto.Encrypt([]byte(value), &pubKey, ownerPriv)
		if err != nil {
			return fmt.Errorf("encrypt for %s: %w", member.Hex(), err)
		}
		memberAddrs = append(memberAddrs, member)
		encryptedValues = append(encryptedValues, encrypted)
	}
	return client.SetSecret(id.EthPrivKey, projectName, key, memberAddrs, encryptedValues)
}

func fetchAndDecrypt(
	id *identity.Identity,
	client *chain.Client,
	projectName, secretName string,
) (string, error) {
	blob, err := client.GetSecret(id.EthPrivKey, projectName, secretName)
	if err != nil {
		return "", fmt.Errorf("get secret %q: %w", secretName, err)
	}
	if len(blob) == 0 {
		return "", fmt.Errorf("secret %q not found (or not set for your address)", secretName)
	}

	ownerAddr, err := client.GetOwner(projectName)
	if err != nil {
		return "", fmt.Errorf("get project owner: %w", err)
	}

	ownerPubKey, err := client.GetPublicKey(ownerAddr)
	if err != nil {
		return "", fmt.Errorf("get owner public key: %w", err)
	}

	myPrivKey, err := id.NaclPrivKeyBytes()
	if err != nil {
		return "", err
	}

	plain, err := enpcrypto.Decrypt(blob, &ownerPubKey, myPrivKey)
	if err != nil {
		return "", fmt.Errorf("decrypt %q: %w", secretName, err)
	}
	return string(plain), nil
}

func pullSecrets(
	id *identity.Identity,
	client *chain.Client,
	projectName string,
) (map[string]string, error) {
	names, err := client.GetSecretNames(id.EthPrivKey, projectName)
	if err != nil {
		return nil, fmt.Errorf("get secret names: %w", err)
	}

	secrets := make(map[string]string, len(names))
	for _, name := range names {
		value, err := fetchAndDecrypt(id, client, projectName, name)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", name, err)
		}
		secrets[name] = value
	}
	return secrets, nil
}

func saveCache(projectName string, secrets map[string]string) error {
	dir := filepath.Join(config.Dir(), "cache")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	var sb strings.Builder
	for k, v := range secrets {
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
		sb.WriteByte('\n')
	}
	return os.WriteFile(cachePath(projectName), []byte(sb.String()), 0600)
}

func cachePath(projectName string) string {
	return filepath.Join(config.Dir(), "cache", projectName+".env")
}

func init() {
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(runCmd)
}
