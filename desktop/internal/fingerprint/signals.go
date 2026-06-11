package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Signals matches browser/device characteristics used for fingerprint_hash.
type Signals struct {
	UserAgent        string `json:"userAgent"`
	Platform         string `json:"platform"`
	Language         string `json:"language"`
	ScreenWidth      int    `json:"screenWidth"`
	ScreenHeight     int    `json:"screenHeight"`
	ColorDepth       int    `json:"colorDepth"`
	TimeZone         string `json:"timeZone"`
	TimezoneOffset   int    `json:"timezoneOffset"`
	CanvasHash       string `json:"canvasHash"`
	WebGLVendor      string `json:"webGLVendor"`
	WebGLRenderer    string `json:"webGLRenderer"`
	AudioContextHash string `json:"audioContextHash"`
	FontsHash        string `json:"fontsHash"`
}

// ComputeHash — SHA-256 fingerprint (same algorithm as fingerprint-service).
func ComputeHash(s Signals) string {
	data := fmt.Sprintf("%s|%s|%s|%d|%d|%d|%s|%d|%s|%s|%s|%s|%s",
		s.UserAgent,
		s.Platform,
		s.Language,
		s.ScreenWidth,
		s.ScreenHeight,
		s.ColorDepth,
		s.TimeZone,
		s.TimezoneOffset,
		s.CanvasHash,
		s.WebGLVendor,
		s.WebGLRenderer,
		s.AudioContextHash,
		s.FontsHash,
	)
	sum := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", sum)
}

// DeviceHash combines fingerprint hash with per-installation salt.
func DeviceHash(dataDir string, signals Signals) (string, error) {
	installID, err := installationID(dataDir)
	if err != nil {
		return "", err
	}

	fpHash := ComputeHash(signals)
	payload := installID + "|" + fpHash
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:16]), nil
}

func installationID(dataDir string) (string, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", fmt.Errorf("create data dir: %w", err)
	}

	path := filepath.Join(dataDir, "installation_id")
	raw, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("read installation id: %w", err)
		}
		id := uuid.NewString()
		if err := os.WriteFile(path, []byte(id), 0o600); err != nil {
			return "", fmt.Errorf("write installation id: %w", err)
		}
		return id, nil
	}

	return string(raw), nil
}
