package auth

import (
	"crypto/subtle"
)

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
