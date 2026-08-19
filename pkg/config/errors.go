package config

import "errors"

var (
	ErrInvalidURI        = errors.New("invalid mcp-wire URI format, expected mcp://<PeerID>/<ServiceName>")
	ErrInvalidPeerID     = errors.New("invalid peer identity ID")
	ErrUnauthorizedPeer  = errors.New("peer is not authorized by access control list")
	ErrInvalidToken      = errors.New("invalid authentication token")
	ErrStreamClosed      = errors.New("p2p stream closed unexpectedly")
	ErrTimeout           = errors.New("operation timed out")
	ErrServiceNotFound   = errors.New("requested mcp service not found")
	ErrReadOnlyViolation = errors.New("method not allowed in read-only mode")
)
