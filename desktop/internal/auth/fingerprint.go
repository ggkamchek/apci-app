package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func DeviceHash(dataDir string) (string, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", fmt.Errorf("create data dir: %w", err)
	}

	path := filepath.Join(dataDir, "installation_id")
	id, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("read installation id: %w", err)
		}
		id = []byte(uuid.NewString())
		if err := os.WriteFile(path, id, 0o600); err != nil {
			return "", fmt.Errorf("write installation id: %w", err)
		}
	}

	sum := sha256.Sum256(id)
	return hex.EncodeToString(sum[:16]), nil
}
