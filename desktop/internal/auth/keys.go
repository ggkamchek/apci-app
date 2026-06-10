package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

type KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

func GenerateKeyPair() (KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyPair{}, fmt.Errorf("generate keypair: %w", err)
	}
	return KeyPair{PublicKey: publicKey, PrivateKey: privateKey}, nil
}

func SaveKeyPair(keysDir, username string, pair KeyPair) error {
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		return fmt.Errorf("create keys dir: %w", err)
	}

	seed := pair.PrivateKey.Seed()
	path := keyPath(keysDir, username)
	encoded := base64.StdEncoding.EncodeToString(seed)
	if err := os.WriteFile(path, []byte(encoded), 0o600); err != nil {
		return fmt.Errorf("write key file: %w", err)
	}
	return nil
}

func LoadKeyPair(keysDir, username string) (KeyPair, error) {
	path := keyPath(keysDir, username)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return KeyPair{}, fmt.Errorf("ключи для пользователя %q не найдены на этом устройстве — зарегистрируйтесь", username)
		}
		return KeyPair{}, fmt.Errorf("read key file: %w", err)
	}

	seed, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		return KeyPair{}, fmt.Errorf("decode key: %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return KeyPair{}, fmt.Errorf("invalid key size")
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	return KeyPair{
		PublicKey:  privateKey.Public().(ed25519.PublicKey),
		PrivateKey: privateKey,
	}, nil
}

func HasKeyPair(keysDir, username string) bool {
	_, err := os.Stat(keyPath(keysDir, username))
	return err == nil
}

func keyPath(keysDir, username string) string {
	return filepath.Join(keysDir, username+".key")
}
