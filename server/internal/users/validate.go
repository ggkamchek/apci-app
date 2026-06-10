package users

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("username must be 3-32 chars: letters, digits, underscore")
	}
	return nil
}

func validateDeviceHash(deviceHash string) error {
	deviceHash = strings.TrimSpace(deviceHash)
	if deviceHash == "" {
		return fmt.Errorf("device_hash is required")
	}
	if utf8.RuneCountInString(deviceHash) > 128 {
		return fmt.Errorf("device_hash is too long")
	}
	return nil
}

func validatePublicKey(publicKey []byte) error {
	if len(publicKey) != 32 {
		return fmt.Errorf("ed25519_public_key must be exactly 32 bytes")
	}
	return nil
}

func validateSignature(signature []byte) error {
	if len(signature) != 64 {
		return fmt.Errorf("signature must be exactly 64 bytes")
	}
	return nil
}
