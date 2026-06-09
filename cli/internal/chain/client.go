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

// Client wraps ethclient with typed methods for IdentityRegistry and ProjectVault.
type Client struct {
	eth     *ethclient.Client
	chainID *big.Int
	reg     *bind.BoundContract
	vault   *bind.BoundContract
}

func New(rpcURL, registryAddr, vaultAddr string) (*Client, error) {
	eth, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial RPC %s: %w", rpcURL, err)
	}

	chainID, err := eth.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get chain ID: %w", err)
	}

	regABI, err := abi.JSON(strings.NewReader(IdentityRegistryABI))
	if err != nil {
		return nil, fmt.Errorf("parse registry ABI: %w", err)
	}
	vaultABI, err := abi.JSON(strings.NewReader(ProjectVaultABI))
	if err != nil {
		return nil, fmt.Errorf("parse vault ABI: %w", err)
	}

	regAddr := common.HexToAddress(registryAddr)
	vAddr := common.HexToAddress(vaultAddr)

	return &Client{
		eth:     eth,
		chainID: chainID,
		reg:     bind.NewBoundContract(regAddr, regABI, eth, eth, eth),
		vault:   bind.NewBoundContract(vAddr, vaultABI, eth, eth, eth),
	}, nil
}

func (c *Client) ChainID() *big.Int { return c.chainID }

func (c *Client) auth(privKeyHex string) (*bind.TransactOpts, error) {
	b, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	key, err := crypto.ToECDSA(b)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return bind.NewKeyedTransactorWithChainID(key, c.chainID)
}

func (c *Client) RegisterIdentity(privKeyHex string, naclPubKey [32]byte) error {
	auth, err := c.auth(privKeyHex)
	if err != nil {
		return err
	}
	tx, err := c.reg.Transact(auth, "register", naclPubKey)
	if err != nil {
		return fmt.Errorf("register tx: %w", err)
	}
	_, err = bind.WaitMined(context.Background(), c.eth, tx)
	return err
}

func (c *Client) GetPublicKey(addr common.Address) ([32]byte, error) {
	var out []interface{}
	if err := c.reg.Call(nil, &out, "getPublicKey", addr); err != nil {
		return [32]byte{}, err
	}
	return out[0].([32]byte), nil
}

func (c *Client) IsRegistered(addr common.Address) (bool, error) {
	var out []interface{}
	if err := c.reg.Call(nil, &out, "isRegistered", addr); err != nil {
		return false, err
	}
	return out[0].(bool), nil
}

func (c *Client) CreateProject(privKeyHex, name string) error {
	auth, err := c.auth(privKeyHex)
	if err != nil {
		return err
	}
	tx, err := c.vault.Transact(auth, "createProject", name)
	if err != nil {
		return fmt.Errorf("createProject tx: %w", err)
	}
	_, err = bind.WaitMined(context.Background(), c.eth, tx)
	return err
}

func (c *Client) AddMember(privKeyHex, projectName string, member common.Address) error {
	auth, err := c.auth(privKeyHex)
	if err != nil {
		return err
	}
	id := ProjectID(projectName)
	tx, err := c.vault.Transact(auth, "addMember", id, member)
	if err != nil {
		return fmt.Errorf("addMember tx: %w", err)
	}
	_, err = bind.WaitMined(context.Background(), c.eth, tx)
	return err
}

func (c *Client) SetSecret(
	privKeyHex, projectName, secretName string,
	members []common.Address,
	encryptedValues [][]byte,
) error {
	auth, err := c.auth(privKeyHex)
	if err != nil {
		return err
	}
	id := ProjectID(projectName)
	tx, err := c.vault.Transact(auth, "setSecret", id, secretName, members, encryptedValues)
	if err != nil {
		return fmt.Errorf("setSecret tx: %w", err)
	}
	_, err = bind.WaitMined(context.Background(), c.eth, tx)
	return err
}

func (c *Client) GetSecret(privKeyHex, projectName, secretName string) ([]byte, error) {
	auth, err := c.auth(privKeyHex)
	if err != nil {
		return nil, err
	}
	id := ProjectID(projectName)
	opts := &bind.CallOpts{From: auth.From}
	var out []interface{}
	if err := c.vault.Call(opts, &out, "getSecret", id, secretName); err != nil {
		return nil, fmt.Errorf("getSecret: %w", err)
	}
	return out[0].([]byte), nil
}

func (c *Client) GetMembers(projectName string) ([]common.Address, error) {
	id := ProjectID(projectName)
	var out []interface{}
	if err := c.vault.Call(nil, &out, "getMembers", id); err != nil {
		return nil, err
	}
	return out[0].([]common.Address), nil
}

func (c *Client) GetSecretNames(privKeyHex, projectName string) ([]string, error) {
	auth, err := c.auth(privKeyHex)
	if err != nil {
		return nil, err
	}
	id := ProjectID(projectName)
	opts := &bind.CallOpts{From: auth.From}
	var out []interface{}
	if err := c.vault.Call(opts, &out, "getSecretNames", id); err != nil {
		return nil, err
	}
	return out[0].([]string), nil
}

func (c *Client) GetOwner(projectName string) (common.Address, error) {
	return c.GetOwnerByID(ProjectID(projectName))
}

func (c *Client) GetOwnerByID(id [32]byte) (common.Address, error) {
	var out []interface{}
	if err := c.vault.Call(nil, &out, "getOwner", id); err != nil {
		return common.Address{}, err
	}
	return out[0].(common.Address), nil
}

func (c *Client) GetMemberProjects(addr common.Address) ([][32]byte, error) {
	var out []interface{}
	if err := c.vault.Call(nil, &out, "getMemberProjects", addr); err != nil {
		return nil, err
	}
	return out[0].([][32]byte), nil
}

func (c *Client) GetProjectName(id [32]byte) (string, error) {
	var out []interface{}
	if err := c.vault.Call(nil, &out, "projectNames", id); err != nil {
		return "", err
	}
	return out[0].(string), nil
}

// ProjectID computes the on-chain project ID from a name (keccak256).
func ProjectID(name string) [32]byte {
	return crypto.Keccak256Hash([]byte(name))
}
