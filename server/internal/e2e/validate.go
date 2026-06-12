package e2e

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	x25519KeySize       = 32
	maxCiphertextSize   = 64 * 1024
	maxRatchetHeader    = 4 * 1024
	maxOneTimePrekeys   = 100
	maxRatchetSessionID = 128
	maxChatID           = 128
	defaultFetchLimit   = 50
	maxFetchLimit       = 100
)

func validateX25519PublicKey(key []byte, field string) error {
	if len(key) != x25519KeySize {
		return fmt.Errorf("%s must be exactly 32 bytes", field)
	}
	return nil
}

func validateNonEmptyBytes(data []byte, field string) error {
	if len(data) == 0 {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

func validateCiphertext(data []byte) error {
	if err := validateNonEmptyBytes(data, "ciphertext"); err != nil {
		return err
	}
	if len(data) > maxCiphertextSize {
		return fmt.Errorf("ciphertext is too large")
	}
	return nil
}

func validateRatchetHeader(data []byte) error {
	if err := validateNonEmptyBytes(data, "ratchet_header"); err != nil {
		return err
	}
	if len(data) > maxRatchetHeader {
		return fmt.Errorf("ratchet_header is too large")
	}
	return nil
}

func validateUserID(userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func validateChatID(chatID string) error {
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return fmt.Errorf("chat_id is required")
	}
	if utf8.RuneCountInString(chatID) > maxChatID {
		return fmt.Errorf("chat_id is too long")
	}
	return nil
}

func validateRatchetSessionID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("ratchet_session_id is required")
	}
	if utf8.RuneCountInString(id) > maxRatchetSessionID {
		return fmt.Errorf("ratchet_session_id is too long")
	}
	return nil
}

func normalizeFetchLimit(limit uint32) int {
	if limit == 0 {
		return defaultFetchLimit
	}
	if limit > maxFetchLimit {
		return maxFetchLimit
	}
	return int(limit)
}
