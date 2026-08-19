package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mcp-wire/mcp-wire/pkg/config"
)

type ACLConfig struct {
	mu           sync.RWMutex
	AllowedPeers []string `json:"allowed_peers"`
}

func NewACLConfig() *ACLConfig {
	return &ACLConfig{
		AllowedPeers: make([]string, 0),
	}
}

func LoadACL(path string) (*ACLConfig, error) {
	if path == "" {
		var err error
		path, err = config.GetACLPath()
		if err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		acl := NewACLConfig()
		_ = SaveACL(path, acl)
		return acl, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ACL file at %s: %w", path, err)
	}

	var acl ACLConfig
	if err := json.Unmarshal(data, &acl); err != nil {
		return nil, fmt.Errorf("invalid ACL JSON structure: %w", err)
	}

	return &acl, nil
}

func SaveACL(path string, acl *ACLConfig) error {
	if path == "" {
		var err error
		path, err = config.GetACLPath()
		if err != nil {
			return err
		}
	}

	acl.mu.RLock()
	data, err := json.MarshalIndent(acl, "", "  ")
	acl.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal ACL JSON: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

func (a *ACLConfig) IsPeerAllowed(pid peer.ID) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// If ACL is empty, default open unless explicitly managed
	if len(a.AllowedPeers) == 0 {
		return true
	}

	pStr := pid.String()
	for _, allowed := range a.AllowedPeers {
		if allowed == pStr || allowed == "*" {
			return true
		}
	}
	return false
}

func (a *ACLConfig) AddAllowedPeer(peerIDStr string) error {
	if peerIDStr != "*" {
		if _, err := peer.Decode(peerIDStr); err != nil {
			return fmt.Errorf("invalid peer ID '%s': %w", peerIDStr, err)
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	for _, p := range a.AllowedPeers {
		if p == peerIDStr {
			return nil // Already added
		}
	}

	a.AllowedPeers = append(a.AllowedPeers, peerIDStr)
	return nil
}

func (a *ACLConfig) RemoveAllowedPeer(peerIDStr string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, p := range a.AllowedPeers {
		if p == peerIDStr {
			a.AllowedPeers = append(a.AllowedPeers[:i], a.AllowedPeers[i+1:]...)
			return true
		}
	}
	return false
}

func (a *ACLConfig) ListAllowedPeers() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	res := make([]string, len(a.AllowedPeers))
	copy(res, a.AllowedPeers)
	return res
}
