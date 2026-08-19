package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultDirName     = ".mcp-wire"
	IdentityFileName   = "identity.key"
	AllowedPeersFile   = "allowed_peers.json"
	SecurityLogFile    = "security.log"
)

var CustomConfigDir string

// GetStateDir returns the absolute path to ~/.mcp-wire (or CustomConfigDir) and creates it with 0700 permissions if needed.
func GetStateDir() (string, error) {
	if CustomConfigDir != "" {
		if err := os.MkdirAll(CustomConfigDir, 0700); err != nil {
			return "", err
		}
		return CustomConfigDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, DefaultDirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// GetIdentityPath returns the path to identity.key
func GetIdentityPath() (string, error) {
	dir, err := GetStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, IdentityFileName), nil
}

// GetACLPath returns the path to allowed_peers.json
func GetACLPath() (string, error) {
	dir, err := GetStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, AllowedPeersFile), nil
}

// GetSecurityLogPath returns the path to security.log
func GetSecurityLogPath() (string, error) {
	dir, err := GetStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SecurityLogFile), nil
}
