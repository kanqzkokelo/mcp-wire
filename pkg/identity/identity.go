package identity

import (
	"bytes"
	"encoding/pem"
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// GetPeerID derives the libp2p peer.ID from a private key.
func GetPeerID(priv crypto.PrivKey) (peer.ID, error) {
	if priv == nil {
		return "", fmt.Errorf("private key is nil")
	}
	return peer.IDFromPrivateKey(priv)
}

// FormatPeerID returns the multibase base58 string representation of a peer.ID.
func FormatPeerID(id peer.ID) string {
	return id.String()
}

// ValidateIdentity signs sample payload and verifies signature against public key.
func ValidateIdentity(priv crypto.PrivKey) error {
	msg := []byte("mcp-wire-identity-sanity-check")
	sig, err := priv.Sign(msg)
	if err != nil {
		return fmt.Errorf("failed to sign sanity check msg: %w", err)
	}

	pub := priv.GetPublic()
	valid, err := pub.Verify(msg, sig)
	if err != nil {
		return fmt.Errorf("verification error: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid signature produced by keypair")
	}
	return nil
}

// ExportPublicKeyPEM formats the public key as a PEM encoded block.
func ExportPublicKeyPEM(priv crypto.PrivKey) (string, error) {
	pub := priv.GetPublic()
	pubBytes, err := crypto.MarshalPublicKey(pub)
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	var buf bytes.Buffer
	if err := pem.Encode(&buf, block); err != nil {
		return "", err
	}
	return buf.String(), nil
}
