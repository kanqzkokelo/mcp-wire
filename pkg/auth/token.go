package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
)

// GenerateSecureToken generates a cryptographically secure random hex string of 2*byteLen characters.
func GenerateSecureToken(byteLen int) (string, error) {
	if byteLen <= 0 {
		byteLen = 16 // Default 16 bytes = 32 hex chars
	}
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidateToken performs constant-time string comparison between provided and expected tokens
func ValidateToken(provided string, expected string) bool {
	if expected == "" {
		return true // Token check disabled if no secret token expected
	}
	if provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
