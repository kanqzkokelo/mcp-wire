package auth

import (
	"testing"

	"github.com/mcp-wire/mcp-wire/pkg/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenValidation(t *testing.T) {
	assert.True(t, ValidateToken("secret123", "secret123"))
	assert.False(t, ValidateToken("wrong", "secret123"))
	assert.False(t, ValidateToken("", "secret123"))
	assert.True(t, ValidateToken("anything", "")) // disabled check
}

func TestGenerateSecureToken(t *testing.T) {
	token1, err := GenerateSecureToken(16)
	require.NoError(t, err)
	assert.Len(t, token1, 32)

	token2, err := GenerateSecureToken(16)
	require.NoError(t, err)
	assert.Len(t, token2, 32)
	assert.NotEqual(t, token1, token2)
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(2.0, 2)
	assert.True(t, limiter.Allow())
	assert.True(t, limiter.Allow())
	assert.False(t, limiter.Allow())
}

func TestACLConfig(t *testing.T) {
	priv1, _ := identity.NewEphemeralIdentity()
	pid1, _ := identity.GetPeerID(priv1)

	priv2, _ := identity.NewEphemeralIdentity()
	pid2, _ := identity.GetPeerID(priv2)

	acl := NewACLConfig()

	// Empty ACL allows all by default
	assert.True(t, acl.IsPeerAllowed(pid1))

	// Add pid1 to ACL
	err := acl.AddAllowedPeer(pid1.String())
	require.NoError(t, err)

	assert.True(t, acl.IsPeerAllowed(pid1))
	assert.False(t, acl.IsPeerAllowed(pid2))

	// Remove pid1
	removed := acl.RemoveAllowedPeer(pid1.String())
	assert.True(t, removed)
	assert.True(t, acl.IsPeerAllowed(pid2)) // back to empty = allow all
}
