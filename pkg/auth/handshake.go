package auth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/mcp-wire/mcp-wire/pkg/config"
)

const HandshakeTimeout = 5 * time.Second

type AuthHandshakeFrame struct {
	Token         string `json:"token,omitempty"`
	ClientVersion string `json:"client_version,omitempty"`
	Service       string `json:"service,omitempty"`
}

type AuthResponseFrame struct {
	Status  string `json:"status"` // "ok" or "unauthorized"
	Message string `json:"message,omitempty"`
}

// PerformClientHandshake sends the initial JSON auth frame to the remote host
func PerformClientHandshake(stream network.Stream, token string, serviceName string) error {
	_ = stream.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer func() { _ = stream.SetDeadline(time.Time{}) }()

	frame := AuthHandshakeFrame{
		Token:         token,
		ClientVersion: "1.0.0",
		Service:       serviceName,
	}

	data, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to send handshake frame: %w", err)
	}

	reader := bufio.NewReader(stream)
	respLine, err := ReadLineBounded(reader, MaxFrameBytes)
	if err != nil {
		return fmt.Errorf("failed to read handshake response: %w", err)
	}

	var resp AuthResponseFrame
	if err := json.Unmarshal(respLine, &resp); err != nil {
		return fmt.Errorf("invalid handshake response JSON: %w", err)
	}

	if resp.Status != "ok" {
		return fmt.Errorf("%w: %s", config.ErrUnauthorizedPeer, resp.Message)
	}

	return nil
}

// PerformHostHandshake validates the incoming client auth frame against token & ACL
func PerformHostHandshake(stream network.Stream, expectedToken string, acl *ACLConfig) (*AuthHandshakeFrame, error) {
	_ = stream.SetDeadline(time.Now().Add(HandshakeTimeout))
	defer func() { _ = stream.SetDeadline(time.Time{}) }()

	// 1. Verify remote peer ID against ACL
	remotePeer := stream.Conn().RemotePeer()
	if acl != nil && !acl.IsPeerAllowed(remotePeer) {
		sendAuthResponse(stream, "unauthorized", "peer ID not whitelisted in ACL")
		return nil, fmt.Errorf("%w: peer %s not allowed", config.ErrUnauthorizedPeer, remotePeer.String())
	}

	// 2. Read handshake frame
	reader := bufio.NewReader(stream)
	reqLine, err := ReadLineBounded(reader, MaxFrameBytes)
	if err != nil {
		sendAuthResponse(stream, "unauthorized", "failed to read handshake frame")
		return nil, fmt.Errorf("handshake read failed: %w", err)
	}

	var frame AuthHandshakeFrame
	if err := json.Unmarshal(reqLine, &frame); err != nil {
		sendAuthResponse(stream, "unauthorized", "invalid handshake frame format")
		return nil, fmt.Errorf("handshake unmarshal failed: %w", err)
	}

	// 3. Verify token
	if expectedToken != "" && !ValidateToken(frame.Token, expectedToken) {
		sendAuthResponse(stream, "unauthorized", "invalid pre-shared authentication token")
		return nil, config.ErrInvalidToken
	}

	// 4. Send success response
	if err := sendAuthResponse(stream, "ok", "authenticated successfully"); err != nil {
		return nil, err
	}

	return &frame, nil
}

func sendAuthResponse(stream network.Stream, status string, message string) error {
	resp := AuthResponseFrame{
		Status:  status,
		Message: message,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = stream.Write(data)
	return err
}
