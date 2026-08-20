package protocol

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mcp-wire/mcp-wire/pkg/config"
)

var (
	// ServiceNameRegex enforces alphanumeric, dashes, and underscores (1-64 chars)
	serviceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
	protocolPrefix   = "/mcp-wire/v1/"
)

type MCPURI struct {
	PeerID      peer.ID
	ServiceName string
	Token       string
	Raw         string
}

// ValidateServiceName checks service name compliance.
func ValidateServiceName(name string) error {
	if !serviceNameRegex.MatchString(name) {
		return fmt.Errorf("invalid service name '%s': must be 1-64 alphanumeric characters, dashes, or underscores", name)
	}
	return nil
}

// ParseURI parses mcp://<PeerID>/<ServiceName>?token=<secret>
func ParseURI(rawURI string) (*MCPURI, error) {
	if !strings.HasPrefix(rawURI, "mcp://") {
		return nil, config.ErrInvalidURI
	}

	// Parse as standard URL
	u, err := url.Parse(rawURI)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", config.ErrInvalidURI, err)
	}

	peerIDStr := u.Host
	if peerIDStr == "" {
		return nil, fmt.Errorf("%w: missing peer ID host", config.ErrInvalidURI)
	}

	pid, err := peer.Decode(peerIDStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid peer ID '%s': %v", config.ErrInvalidPeerID, peerIDStr, err)
	}

	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return nil, fmt.Errorf("%w: missing service name path", config.ErrInvalidURI)
	}

	if err := ValidateServiceName(path); err != nil {
		return nil, err
	}

	token := u.Query().Get("token")

	return &MCPURI{
		PeerID:      pid,
		ServiceName: path,
		Token:       token,
		Raw:         rawURI,
	}, nil
}

// ToProtocolID converts service name to libp2p protocol ID /mcp-wire/v1/<serviceName>
func ToProtocolID(serviceName string) protocol.ID {
	return protocol.ID(protocolPrefix + serviceName)
}

// FromProtocolID resolves service name from /mcp-wire/v1/<serviceName>
func FromProtocolID(protoID protocol.ID) (string, error) {
	str := string(protoID)
	if !strings.HasPrefix(str, protocolPrefix) {
		return "", fmt.Errorf("invalid mcp-wire protocol ID: %s", str)
	}
	name := strings.TrimPrefix(str, protocolPrefix)
	if err := ValidateServiceName(name); err != nil {
		return "", err
	}
	return name, nil
}

// BuildURI constructs canonical mcp://<PeerID>/<ServiceName>
func BuildURI(peerID peer.ID, serviceName string, token string) string {
	base := fmt.Sprintf("mcp://%s/%s", peerID.String(), serviceName)
	if token != "" {
		base += "?token=" + url.QueryEscape(token)
	}
	return base
}

// RedactURI replaces secret tokens in connection URIs with '***' for logging/diagnostics.
func RedactURI(rawURI string) string {
	if !strings.Contains(rawURI, "token=") {
		return rawURI
	}
	u, err := url.Parse(rawURI)
	if err != nil {
		return rawURI
	}
	q := u.Query()
	if q.Get("token") != "" {
		q.Set("token", "***")
		u.RawQuery = q.Encode()
	}
	return u.String()
}
