package cmd

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/enparse/cli/internal/chain"
	"github.com/enparse/cli/internal/config"
	"github.com/enparse/cli/internal/identity"
)

// loadAll loads identity, config, and chain client. Shared by all commands.
func loadAll() (*identity.Identity, *config.Config, *chain.Client, error) {
	id, err := identity.Load()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("no identity found — run 'enparse init' first")
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, nil, err
	}
	client, err := chain.New(cfg.RPCURL, cfg.IdentityRegistryAddr, cfg.ProjectVaultAddr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("connect to chain: %w", err)
	}
	return id, cfg, client, nil
}

func ethAddr(hex string) common.Address {
	return common.HexToAddress(hex)
}

func or(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
