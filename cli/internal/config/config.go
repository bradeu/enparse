package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Env vars override the config file. Useful for CI/CD and container deployments.
//
//	ENPARSE_RPC_URL        overrides rpc_url
//	ENPARSE_REGISTRY_ADDR  overrides identity_registry_addr
//	ENPARSE_VAULT_ADDR     overrides project_vault_addr

type Config struct {
	RPCURL               string `json:"rpc_url"`
	IdentityRegistryAddr string `json:"identity_registry_addr"`
	ProjectVaultAddr     string `json:"project_vault_addr"`
}

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".enparse")
}

func Path() string {
	return filepath.Join(Dir(), "config.json")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			cfg := &Config{RPCURL: "http://127.0.0.1:8545"}
			applyEnvOverrides(cfg)
			return cfg, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("ENPARSE_RPC_URL"); v != "" {
		cfg.RPCURL = v
	}
	if v := os.Getenv("ENPARSE_REGISTRY_ADDR"); v != "" {
		cfg.IdentityRegistryAddr = v
	}
	if v := os.Getenv("ENPARSE_VAULT_ADDR"); v != "" {
		cfg.ProjectVaultAddr = v
	}
}

func (c *Config) Save() error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0600)
}

func (c *Config) Validate() error {
	if c.RPCURL == "" {
		return fmt.Errorf("rpc_url not configured (run: enparse config set rpc_url <url>)")
	}
	if c.IdentityRegistryAddr == "" {
		return fmt.Errorf("identity_registry_addr not configured (run deploy script first)")
	}
	if c.ProjectVaultAddr == "" {
		return fmt.Errorf("project_vault_addr not configured (run deploy script first)")
	}
	return nil
}
