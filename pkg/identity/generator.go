package identity

import (
	"crypto/rand"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// GenerateIdentity creates a new Ed25519 keypair for libp2p.
func GenerateIdentity() (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateEd25519Key(rand.Reader)
}

// NewEphemeralIdentity creates an in-memory throwaway keypair for testing.
func NewEphemeralIdentity() (crypto.PrivKey, error) {
	priv, _, err := GenerateIdentity()
	return priv, err
}
