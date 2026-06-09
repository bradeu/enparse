package identity

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/nacl/box"
)

type Identity struct {
	EthAddress  string `json:"eth_address"`
	EthPrivKey  string `json:"eth_privkey"`
	NaclPubKey  string `json:"nacl_pubkey"`
	NaclPrivKey string `json:"nacl_privkey"`
}

func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".enparse", "identity.json")
}

func Generate() (*Identity, error) {
	ethKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("generate eth key: %w", err)
	}

	naclPub, naclPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate nacl key: %w", err)
	}

	return &Identity{
		EthAddress:  crypto.PubkeyToAddress(ethKey.PublicKey).Hex(),
		EthPrivKey:  hex.EncodeToString(crypto.FromECDSA(ethKey)),
		NaclPubKey:  hex.EncodeToString(naclPub[:]),
		NaclPrivKey: hex.EncodeToString(naclPriv[:]),
	}, nil
}

func Load() (*Identity, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		return nil, err
	}
	var id Identity
	return &id, json.Unmarshal(data, &id)
}

func (id *Identity) Save() error {
	dir := filepath.Dir(Path())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(id, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0600)
}

func (id *Identity) EthPrivKeyECDSA() (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(id.EthPrivKey)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(b)
}

func (id *Identity) NaclPubKeyBytes() (*[32]byte, error) {
	b, err := hex.DecodeString(id.NaclPubKey)
	if err != nil {
		return nil, err
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("nacl pubkey must be 32 bytes")
	}
	var key [32]byte
	copy(key[:], b)
	return &key, nil
}

func (id *Identity) NaclPrivKeyBytes() (*[32]byte, error) {
	b, err := hex.DecodeString(id.NaclPrivKey)
	if err != nil {
		return nil, err
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("nacl privkey must be 32 bytes")
	}
	var key [32]byte
	copy(key[:], b)
	return &key, nil
}
