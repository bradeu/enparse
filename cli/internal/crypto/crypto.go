package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/nacl/box"
)

// Encrypt encrypts plaintext for recipient using sender's private key.
// Returns nonce (24 bytes) || ciphertext.
func Encrypt(plaintext []byte, recipientPub, senderPriv *[32]byte) ([]byte, error) {
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}
	sealed := box.Seal(nonce[:], plaintext, &nonce, recipientPub, senderPriv)
	return sealed, nil
}

// Decrypt decrypts a blob (nonce || ciphertext).
// senderPub is the public key of whoever encrypted it.
// recipientPriv is the caller's private key.
func Decrypt(blob []byte, senderPub, recipientPriv *[32]byte) ([]byte, error) {
	if len(blob) < 24+box.Overhead {
		return nil, fmt.Errorf("blob too short (%d bytes)", len(blob))
	}
	var nonce [24]byte
	copy(nonce[:], blob[:24])
	plain, ok := box.Open(nil, blob[24:], &nonce, senderPub, recipientPriv)
	if !ok {
		return nil, fmt.Errorf("decryption failed: wrong key or corrupted data")
	}
	return plain, nil
}
