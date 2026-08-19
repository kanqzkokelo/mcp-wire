package identity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentityGenerationAndSerialization(t *testing.T) {
	priv, pub, err := GenerateIdentity()
	require.NoError(t, err)
	require.NotNil(t, priv)
	require.NotNil(t, pub)

	// Test Peer ID derivation
	pid, err := GetPeerID(priv)
	require.NoError(t, err)
	assert.NotEmpty(t, FormatPeerID(pid))

	// Test Sanity check
	err = ValidateIdentity(priv)
	assert.NoError(t, err)

	// Test Serialization
	data, err := MarshalPrivateKey(priv)
	require.NoError(t, err)

	unmarshaled, err := UnmarshalPrivateKey(data)
	require.NoError(t, err)

	pid2, err := GetPeerID(unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, pid, pid2)
}

func TestIdentityStoreAndPermissions(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "test_identity.key")

	priv, err := NewEphemeralIdentity()
	require.NoError(t, err)

	// Save
	err = SaveIdentityKey(keyPath, priv)
	require.NoError(t, err)

	// Check permission 0600
	info, err := os.Stat(keyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Load
	loaded, err := LoadIdentityKey(keyPath)
	require.NoError(t, err)

	pid1, _ := GetPeerID(priv)
	pid2, _ := GetPeerID(loaded)
	assert.Equal(t, pid1, pid2)
}
