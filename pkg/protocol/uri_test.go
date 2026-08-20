package protocol

import (
	"testing"

	"github.com/mcp-wire/mcp-wire/pkg/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURIValid(t *testing.T) {
	priv, err := identity.NewEphemeralIdentity()
	require.NoError(t, err)

	pid, err := identity.GetPeerID(priv)
	require.NoError(t, err)

	raw := "mcp://" + pid.String() + "/gpu-whisper?token=secret123"
	mcpURI, err := ParseURI(raw)
	require.NoError(t, err)

	assert.Equal(t, pid, mcpURI.PeerID)
	assert.Equal(t, "gpu-whisper", mcpURI.ServiceName)
	assert.Equal(t, "secret123", mcpURI.Token)

	// Test Protocol ID conversion
	protoID := ToProtocolID(mcpURI.ServiceName)
	assert.Equal(t, "/mcp-wire/v1/gpu-whisper", string(protoID))

	resolvedName, err := FromProtocolID(protoID)
	require.NoError(t, err)
	assert.Equal(t, "gpu-whisper", resolvedName)

	// Test Build URI
	rebuilt := BuildURI(pid, "gpu-whisper", "secret123")
	assert.Equal(t, raw, rebuilt)
}

func TestParseURIInvalid(t *testing.T) {
	_, err := ParseURI("http://invalid/service")
	assert.Error(t, err)

	_, err = ParseURI("mcp://invalid-peer-id/service")
	assert.Error(t, err)

	priv, _ := identity.NewEphemeralIdentity()
	pid, _ := identity.GetPeerID(priv)

	_, err = ParseURI("mcp://" + pid.String() + "/invalid@service!")
	assert.Error(t, err)
}

func TestRedactURI(t *testing.T) {
	raw := "mcp://12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq395JTV5V15fs5i4FiE/gpu-whisper?token=secret123"
	redacted := RedactURI(raw)
	assert.Contains(t, redacted, "token=%2A%2A%2A")
	assert.NotContains(t, redacted, "secret123")

	noToken := "mcp://12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq395JTV5V15fs5i4FiE/gpu-whisper"
	assert.Equal(t, noToken, RedactURI(noToken))
}
