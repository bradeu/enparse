package chain

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// DeployResult holds addresses of deployed contracts.
type DeployResult struct {
	IdentityRegistryAddr string
	ProjectVaultAddr     string
	ChainID              *big.Int
}

// DeployContracts deploys IdentityRegistry then ProjectVault and returns their addresses.
// Nonce is managed automatically via PendingNonceAt — no manual reset needed between txs.
func DeployContracts(ctx context.Context, rpcURL, privKeyHex string) (*DeployResult, error) {
	eth, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial RPC: %w", err)
	}
	defer eth.Close()

	chainID, err := eth.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get chain ID: %w", err)
	}

	auth, err := newTransactor(privKeyHex, chainID)
	if err != nil {
		return nil, err
	}

	regAddr, err := deployOne(ctx, auth, eth, IdentityRegistryABI, IdentityRegistryBytecode)
	if err != nil {
		return nil, fmt.Errorf("deploy IdentityRegistry: %w", err)
	}

	vaultAddr, err := deployOne(ctx, auth, eth, ProjectVaultABI, ProjectVaultBytecode, regAddr)
	if err != nil {
		return nil, fmt.Errorf("deploy ProjectVault: %w", err)
	}

	return &DeployResult{
		IdentityRegistryAddr: regAddr.Hex(),
		ProjectVaultAddr:     vaultAddr.Hex(),
		ChainID:              chainID,
	}, nil
}

// deployBackend combines the two interfaces required for deploy + wait.
type deployBackend interface {
	bind.ContractBackend
	bind.DeployBackend
}

func deployOne(ctx context.Context, auth *bind.TransactOpts, backend deployBackend, abiStr, bytecodeHex string, constructorArgs ...interface{}) (common.Address, error) {
	parsed, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return common.Address{}, fmt.Errorf("parse ABI: %w", err)
	}
	bytecode := common.FromHex(bytecodeHex)
	addr, tx, _, err := bind.DeployContract(auth, parsed, bytecode, backend, constructorArgs...)
	if err != nil {
		return common.Address{}, err
	}
	if _, err := bind.WaitDeployed(ctx, backend, tx); err != nil {
		return common.Address{}, fmt.Errorf("wait deployment: %w", err)
	}
	return addr, nil
}

func newTransactor(privKeyHex string, chainID *big.Int) (*bind.TransactOpts, error) {
	b, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	key, err := crypto.ToECDSA(b)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return bind.NewKeyedTransactorWithChainID(key, chainID)
}
