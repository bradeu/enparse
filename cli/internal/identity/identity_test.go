package identity

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	gocrypto "github.com/ethereum/go-ethereum/crypto"
)

// ─── Generate ────────────────────────────────────────────────────────────────

func TestGenerateReturnsNonNil(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if id == nil {
		t.Fatal("Generate returned nil identity")
	}
}

func TestGenerateEthAddressIsValid(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !common.IsHexAddress(id.EthAddress) {
		t.Errorf("EthAddress %q is not a valid hex address", id.EthAddress)
	}
}

func TestGenerateEthPrivKeyDecodesAndMatchesAddress(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	privKey, err := id.EthPrivKeyECDSA()
	if err != nil {
		t.Fatalf("EthPrivKeyECDSA: %v", err)
	}

	derived := gocrypto.PubkeyToAddress(privKey.PublicKey).Hex()
	if !strings.EqualFold(derived, id.EthAddress) {
		t.Errorf("derived address %q != stored address %q", derived, id.EthAddress)
	}
}

func TestGenerateNaclPubKeyIs32Bytes(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	b, err := hex.DecodeString(id.NaclPubKey)
	if err != nil {
		t.Fatalf("decode NaclPubKey hex: %v", err)
	}
	if len(b) != 32 {
		t.Errorf("NaclPubKey length = %d, want 32", len(b))
	}
}

func TestGenerateNaclPrivKeyIs32Bytes(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	b, err := hex.DecodeString(id.NaclPrivKey)
	if err != nil {
		t.Fatalf("decode NaclPrivKey hex: %v", err)
	}
	if len(b) != 32 {
		t.Errorf("NaclPrivKey length = %d, want 32", len(b))
	}
}

func TestGenerateProducesUniqueIdentities(t *testing.T) {
	id1, err := Generate()
	if err != nil {
		t.Fatalf("Generate (1): %v", err)
	}
	id2, err := Generate()
	if err != nil {
		t.Fatalf("Generate (2): %v", err)
	}

	if strings.EqualFold(id1.EthAddress, id2.EthAddress) {
		t.Error("two Generate calls produced the same Ethereum address")
	}
	if id1.NaclPubKey == id2.NaclPubKey {
		t.Error("two Generate calls produced the same NaCl public key")
	}
}

// ─── NaclPubKeyBytes / NaclPrivKeyBytes ─────────────────────────────────────

func TestNaclPubKeyBytesRoundTrip(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	keyPtr, err := id.NaclPubKeyBytes()
	if err != nil {
		t.Fatalf("NaclPubKeyBytes: %v", err)
	}

	if hex.EncodeToString(keyPtr[:]) != id.NaclPubKey {
		t.Error("NaclPubKeyBytes does not round-trip with NaclPubKey hex field")
	}
}

func TestNaclPrivKeyBytesRoundTrip(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	keyPtr, err := id.NaclPrivKeyBytes()
	if err != nil {
		t.Fatalf("NaclPrivKeyBytes: %v", err)
	}

	if hex.EncodeToString(keyPtr[:]) != id.NaclPrivKey {
		t.Error("NaclPrivKeyBytes does not round-trip with NaclPrivKey hex field")
	}
}

func TestNaclPubKeyBytesRejectsShortHex(t *testing.T) {
	id := &Identity{NaclPubKey: hex.EncodeToString(make([]byte, 16))} // 16 bytes, not 32
	_, err := id.NaclPubKeyBytes()
	if err == nil {
		t.Fatal("expected error for 16-byte NaclPubKey, got nil")
	}
}

func TestNaclPrivKeyBytesRejectsShortHex(t *testing.T) {
	id := &Identity{NaclPrivKey: hex.EncodeToString(make([]byte, 16))}
	_, err := id.NaclPrivKeyBytes()
	if err == nil {
		t.Fatal("expected error for 16-byte NaclPrivKey, got nil")
	}
}

func TestNaclPubKeyBytesRejectsInvalidHex(t *testing.T) {
	id := &Identity{NaclPubKey: "not-hex"}
	_, err := id.NaclPubKeyBytes()
	if err == nil {
		t.Fatal("expected error for invalid hex, got nil")
	}
}

func TestEthPrivKeyECDSARejectsInvalidHex(t *testing.T) {
	id := &Identity{EthPrivKey: "not-hex"}
	_, err := id.EthPrivKeyECDSA()
	if err == nil {
		t.Fatal("expected error for invalid hex private key, got nil")
	}
}

// ─── Save / Load (JSON round-trip) ───────────────────────────────────────────

// saveLoadFixture writes id to a temp file, rewrites Path() to point there,
// then calls Load() and returns the result.
func saveAndLoad(t *testing.T, id *Identity) *Identity {
	t.Helper()

	// Write JSON directly to a temp file.
	dir := t.TempDir()
	p := filepath.Join(dir, "identity.json")

	data, err := json.MarshalIndent(id, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent: %v", err)
	}
	if err := os.WriteFile(p, data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Load it back by reading the file directly (avoids patching global Path()).
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var loaded Identity
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	return &loaded
}

func TestSaveLoadRoundTrip(t *testing.T) {
	original, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	loaded := saveAndLoad(t, original)

	if loaded.EthAddress != original.EthAddress {
		t.Errorf("EthAddress mismatch: got %q, want %q", loaded.EthAddress, original.EthAddress)
	}
	if loaded.EthPrivKey != original.EthPrivKey {
		t.Errorf("EthPrivKey mismatch")
	}
	if loaded.NaclPubKey != original.NaclPubKey {
		t.Errorf("NaclPubKey mismatch")
	}
	if loaded.NaclPrivKey != original.NaclPrivKey {
		t.Errorf("NaclPrivKey mismatch")
	}
}

// TestSaveCreatesDirectoryAndFile tests the Save() method end-to-end by
// pointing Path() at a temp location via the HOME env var override.
func TestSaveCreatesDirectoryAndFile(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Redirect HOME so Path() points inside the temp dir.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := id.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	expectedPath := filepath.Join(tmpHome, ".enparse", "identity.json")
	info, err := os.Stat(expectedPath)
	if err != nil {
		t.Fatalf("identity.json not found after Save: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("file permissions = %o, want 0600", info.Mode().Perm())
	}
}

// TestLoadFromSavedFile verifies that Load() correctly deserialises what
// Save() wrote.
func TestLoadFromSavedFile(t *testing.T) {
	original, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := original.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.EthAddress != original.EthAddress {
		t.Errorf("EthAddress mismatch after Save/Load")
	}
	if loaded.NaclPubKey != original.NaclPubKey {
		t.Errorf("NaclPubKey mismatch after Save/Load")
	}
}

// TestLoadReturnsErrorWhenFileMissing verifies Load returns a non-nil error
// when no identity file exists.
func TestLoadReturnsErrorWhenFileMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // empty directory — no identity.json

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when identity file does not exist, got nil")
	}
}
