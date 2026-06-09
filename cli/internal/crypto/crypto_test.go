package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/nacl/box"
)

// generateKeypair is a test helper that produces a fresh NaCl keypair.
func generateKeypair(t *testing.T) (pub, priv *[32]byte) {
	t.Helper()
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("box.GenerateKey: %v", err)
	}
	return pub, priv
}

// TestRoundTrip is the primary correctness test: whatever Encrypt produces,
// Decrypt must recover exactly.
func TestRoundTrip(t *testing.T) {
	senderPub, senderPriv := generateKeypair(t)
	recipientPub, recipientPriv := generateKeypair(t)

	plaintext := []byte("the quick brown fox jumps over the lazy dog")

	ciphertext, err := Encrypt(plaintext, recipientPub, senderPriv)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	got, err := Decrypt(ciphertext, senderPub, recipientPriv)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Errorf("round-trip mismatch: got %q, want %q", got, plaintext)
	}
}

// TestRoundTripEmptyPlaintext verifies that zero-length plaintext is handled
// correctly — the ciphertext carries overhead even when the message is empty.
func TestRoundTripEmptyPlaintext(t *testing.T) {
	senderPub, senderPriv := generateKeypair(t)
	recipientPub, recipientPriv := generateKeypair(t)

	ciphertext, err := Encrypt([]byte{}, recipientPub, senderPriv)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	got, err := Decrypt(ciphertext, senderPub, recipientPriv)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected empty plaintext, got %d bytes", len(got))
	}
}

// TestRoundTripBinaryData verifies that arbitrary binary (non-UTF-8) bytes
// survive the round-trip intact.
func TestRoundTripBinaryData(t *testing.T) {
	senderPub, senderPriv := generateKeypair(t)
	recipientPub, recipientPriv := generateKeypair(t)

	plaintext := make([]byte, 256)
	for i := range plaintext {
		plaintext[i] = byte(i)
	}

	ciphertext, err := Encrypt(plaintext, recipientPub, senderPriv)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	got, err := Decrypt(ciphertext, senderPub, recipientPriv)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Error("binary round-trip mismatch")
	}
}

// TestDecryptWrongRecipientKey ensures Decrypt returns an error when the
// recipient private key does not correspond to the intended recipient.
func TestDecryptWrongRecipientKey(t *testing.T) {
	senderPub, senderPriv := generateKeypair(t)
	recipientPub, _ := generateKeypair(t)
	_, wrongPriv := generateKeypair(t) // unrelated keypair

	ciphertext, err := Encrypt([]byte("secret"), recipientPub, senderPriv)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Decrypt(ciphertext, senderPub, wrongPriv)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key, got nil")
	}
}

// TestDecryptWrongSenderKey ensures Decrypt fails when given the wrong sender
// public key even if the recipient private key is correct.
func TestDecryptWrongSenderKey(t *testing.T) {
	senderPub, senderPriv := generateKeypair(t)
	recipientPub, recipientPriv := generateKeypair(t)
	wrongPub, _ := generateKeypair(t)

	_ = senderPub // correct sender pub is intentionally not used in Decrypt below

	ciphertext, err := Encrypt([]byte("secret"), recipientPub, senderPriv)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Decrypt(ciphertext, wrongPub, recipientPriv)
	if err == nil {
		t.Fatal("expected error with wrong sender pubkey, got nil")
	}
}

// TestDecryptBlobTooShort verifies that Decrypt rejects a blob that is shorter
// than the minimum possible ciphertext (nonce + box.Overhead bytes).
func TestDecryptBlobTooShort(t *testing.T) {
	_, recipientPriv := generateKeypair(t)
	_, senderPub := generateKeypair(t)

	// 23 bytes < 24 (nonce) + box.Overhead (16)
	short := make([]byte, 23)

	_, err := Decrypt(short, senderPub, recipientPriv)
	if err == nil {
		t.Fatal("expected error for too-short blob, got nil")
	}
}

// TestDecryptCorruptedCiphertext verifies that a bit-flip in the ciphertext
// is detected and returns an error.
func TestDecryptCorruptedCiphertext(t *testing.T) {
	senderPub, senderPriv := generateKeypair(t)
	recipientPub, recipientPriv := generateKeypair(t)

	ciphertext, err := Encrypt([]byte("important secret"), recipientPub, senderPriv)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Flip a byte in the ciphertext portion (after the 24-byte nonce).
	corrupt := make([]byte, len(ciphertext))
	copy(corrupt, ciphertext)
	corrupt[24] ^= 0xFF

	_, err = Decrypt(corrupt, senderPub, recipientPriv)
	if err == nil {
		t.Fatal("expected error for corrupted ciphertext, got nil")
	}
}

// TestEncryptProducesUniqueNonces verifies that two Encrypt calls with the
// same inputs produce different ciphertexts (probabilistic nonce uniqueness).
func TestEncryptProducesUniqueNonces(t *testing.T) {
	senderPub, senderPriv := generateKeypair(t)
	recipientPub, _ := generateKeypair(t)
	_ = senderPub

	plaintext := []byte("same message")

	c1, err := Encrypt(plaintext, recipientPub, senderPriv)
	if err != nil {
		t.Fatalf("Encrypt (1): %v", err)
	}
	c2, err := Encrypt(plaintext, recipientPub, senderPriv)
	if err != nil {
		t.Fatalf("Encrypt (2): %v", err)
	}

	if bytes.Equal(c1, c2) {
		t.Error("two Encrypt calls produced identical output — nonce is not random")
	}
}

// TestEncryptOutputLength verifies the output is exactly nonce(24) + overhead(16)
// + len(plaintext).
func TestEncryptOutputLength(t *testing.T) {
	senderPub, senderPriv := generateKeypair(t)
	recipientPub, _ := generateKeypair(t)
	_ = senderPub

	plaintext := []byte("hello")
	want := 24 + box.Overhead + len(plaintext)

	ciphertext, err := Encrypt(plaintext, recipientPub, senderPriv)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if len(ciphertext) != want {
		t.Errorf("ciphertext length = %d, want %d", len(ciphertext), want)
	}
}
