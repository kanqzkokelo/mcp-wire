package identity

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// MarshalPrivateKey encodes a libp2p private key to bytes.
func MarshalPrivateKey(priv crypto.PrivKey) ([]byte, error) {
	if priv == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	return crypto.MarshalPrivateKey(priv)
}

// UnmarshalPrivateKey decodes raw bytes into a libp2p private key.
func UnmarshalPrivateKey(data []byte) (crypto.PrivKey, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot unmarshal empty key data")
	}
	return crypto.UnmarshalPrivateKey(data)
}
