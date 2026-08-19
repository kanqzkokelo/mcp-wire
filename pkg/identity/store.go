package identity

import (
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/mcp-wire/mcp-wire/pkg/config"
)

// SaveIdentityKey writes the private key to file with 0600 permissions.
func SaveIdentityKey(path string, key crypto.PrivKey) error {
	data, err := MarshalPrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Write key to disk with strict user-only read/write permissions (0600)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write identity key to %s: %w", path, err)
	}

	// Enforce 0600 permissions
	return os.Chmod(path, 0600)
}

// LoadIdentityKey reads a private key from file and validates 0600 permissions.
func LoadIdentityKey(path string) (crypto.PrivKey, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Warn/enforce permission check on POSIX systems
	if info.Mode().Perm()&0077 != 0 {
		// Fix permissions automatically if too loose
		_ = os.Chmod(path, 0600)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read identity key at %s: %w", path, err)
	}

	return UnmarshalPrivateKey(data)
}

// LoadOrGenerateIdentity loads the identity key from default state path, generating if absent.
func LoadOrGenerateIdentity() (crypto.PrivKey, error) {
	keyPath, err := config.GetIdentityPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		priv, _, err := GenerateIdentity()
		if err != nil {
			return nil, fmt.Errorf("failed to generate new identity: %w", err)
		}
		if err := SaveIdentityKey(keyPath, priv); err != nil {
			return nil, err
		}
		return priv, nil
	}

	return LoadIdentityKey(keyPath)
}
