package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/hkdf"
)

// MVP crypto:
// - Identity key: X25519
// - Message encryption: X25519(shared) -> HKDF-SHA256 -> AES-256-GCM
// - ratchet_header carries (sender_identity_pub, nonce)
//
// This is a stepping stone to replace with Double Ratchet without touching gRPC API.

type Store struct {
	path string
}

type Identity struct {
	PrivB64 string `json:"priv"`
	PubB64  string `json:"pub"`
}

type State struct {
	Identity Identity `json:"identity"`
	// Prekeys are kept for UploadPreKeyBundle; for MVP we don't consume them.
	SignedPreKeyPrivB64 string    `json:"signedPreKeyPriv,omitempty"`
	SignedPreKeyPubB64  string    `json:"signedPreKeyPub,omitempty"`
	SignedPreKeyID      uint32    `json:"signedPreKeyId,omitempty"`
	SignedPreKeySigB64  string    `json:"signedPreKeySig,omitempty"`
	SignedPreKeyExpUnix int64     `json:"signedPreKeyExpUnix,omitempty"`
	OneTimePreKeys      []OneTime `json:"oneTimePreKeys,omitempty"`
}

type OneTime struct {
	ID      uint32 `json:"id"`
	PrivB64 string `json:"priv"`
	PubB64  string `json:"pub"`
}

type Header struct {
	SenderIdentityPubB64 string `json:"sip"`
	NonceB64             string `json:"n"`
}

func NewStore(dataDir string) *Store {
	return &Store{path: filepath.Join(dataDir, "crypto", "e2e_state.json")}
}

func (s *Store) LoadOrCreate() (State, error) {
	st, err := s.load()
	if err == nil {
		return st, nil
	}
	if !os.IsNotExist(err) {
		return State{}, err
	}

	st, err = newState()
	if err != nil {
		return State{}, err
	}
	if err := s.save(st); err != nil {
		return State{}, err
	}
	return st, nil
}

func (s *Store) Save(st State) error {
	return s.save(st)
}

func (s *Store) load() (State, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return State{}, err
	}
	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return State{}, fmt.Errorf("parse crypto state: %w", err)
	}
	return st, nil
}

func (s *Store) save(st State) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create crypto dir: %w", err)
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, b, 0o600); err != nil {
		return err
	}
	return nil
}

func newState() (State, error) {
	curve := ecdh.X25519()

	idPriv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return State{}, err
	}

	// Signed prekey (MVP: signature is random bytes, server doesn't verify).
	spkPriv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return State{}, err
	}
	sig := make([]byte, 64)
	if _, err := rand.Read(sig); err != nil {
		return State{}, err
	}

	// A few one-time prekeys.
	ot := make([]OneTime, 0, 10)
	for i := uint32(1); i <= 10; i++ {
		k, err := curve.GenerateKey(rand.Reader)
		if err != nil {
			return State{}, err
		}
		ot = append(ot, OneTime{
			ID:      i,
			PrivB64: b64(k.Bytes()),
			PubB64:  b64(k.PublicKey().Bytes()),
		})
	}

	exp := time.Now().Add(30 * 24 * time.Hour).Unix()

	return State{
		Identity: Identity{
			PrivB64: b64(idPriv.Bytes()),
			PubB64:  b64(idPriv.PublicKey().Bytes()),
		},
		SignedPreKeyPrivB64: b64(spkPriv.Bytes()),
		SignedPreKeyPubB64:  b64(spkPriv.PublicKey().Bytes()),
		SignedPreKeyID:      1,
		SignedPreKeySigB64:  b64(sig),
		SignedPreKeyExpUnix: exp,
		OneTimePreKeys:      ot,
	}, nil
}

func BuildUploadBundle(st State) (identityPub []byte, signedPreKeyPub []byte, signedPreKeySig []byte, signedPreKeyID uint32, signedPreKeyExpUnix int64, oneTimes []struct {
	ID  uint32
	Pub []byte
}, err error) {
	identityPub, err = b64d(st.Identity.PubB64)
	if err != nil {
		return nil, nil, nil, 0, 0, nil, err
	}
	signedPreKeyPub, err = b64d(st.SignedPreKeyPubB64)
	if err != nil {
		return nil, nil, nil, 0, 0, nil, err
	}
	signedPreKeySig, err = b64d(st.SignedPreKeySigB64)
	if err != nil {
		return nil, nil, nil, 0, 0, nil, err
	}

	oneTimes = make([]struct {
		ID  uint32
		Pub []byte
	}, 0, len(st.OneTimePreKeys))
	for _, k := range st.OneTimePreKeys {
		pub, e := b64d(k.PubB64)
		if e != nil {
			return nil, nil, nil, 0, 0, nil, e
		}
		oneTimes = append(oneTimes, struct {
			ID  uint32
			Pub []byte
		}{ID: k.ID, Pub: pub})
	}

	return identityPub, signedPreKeyPub, signedPreKeySig, st.SignedPreKeyID, st.SignedPreKeyExpUnix, oneTimes, nil
}

func EncryptMVP(st State, recipientIdentityPub []byte, plaintext []byte) (ciphertext []byte, headerBytes []byte, err error) {
	if len(plaintext) == 0 {
		return nil, nil, fmt.Errorf("plaintext is required")
	}

	curve := ecdh.X25519()
	idPrivBytes, err := b64d(st.Identity.PrivB64)
	if err != nil {
		return nil, nil, err
	}
	idPriv, err := curve.NewPrivateKey(idPrivBytes)
	if err != nil {
		return nil, nil, err
	}
	recipientPub, err := curve.NewPublicKey(recipientIdentityPub)
	if err != nil {
		return nil, nil, err
	}

	shared, err := idPriv.ECDH(recipientPub)
	if err != nil {
		return nil, nil, err
	}

	key, err := deriveAESKey(shared)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	ct := gcm.Seal(nil, nonce, plaintext, nil)

	h := Header{
		SenderIdentityPubB64: st.Identity.PubB64,
		NonceB64:             b64(nonce),
	}
	hb, err := json.Marshal(h)
	if err != nil {
		return nil, nil, err
	}
	return ct, hb, nil
}

func DecryptMVP(st State, senderIdentityPub []byte, headerBytes []byte, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("ciphertext is required")
	}

	var h Header
	if err := json.Unmarshal(headerBytes, &h); err != nil {
		return nil, fmt.Errorf("invalid header: %w", err)
	}
	nonce, err := b64d(h.NonceB64)
	if err != nil {
		return nil, fmt.Errorf("invalid nonce: %w", err)
	}

	curve := ecdh.X25519()
	idPrivBytes, err := b64d(st.Identity.PrivB64)
	if err != nil {
		return nil, err
	}
	idPriv, err := curve.NewPrivateKey(idPrivBytes)
	if err != nil {
		return nil, err
	}

	senderPub, err := curve.NewPublicKey(senderIdentityPub)
	if err != nil {
		return nil, err
	}

	shared, err := idPriv.ECDH(senderPub)
	if err != nil {
		return nil, err
	}
	key, err := deriveAESKey(shared)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	pt, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %w", err)
	}
	return pt, nil
}

func ParseSenderIdentityPubFromHeader(headerBytes []byte) ([]byte, error) {
	var h Header
	if err := json.Unmarshal(headerBytes, &h); err != nil {
		return nil, err
	}
	return b64d(h.SenderIdentityPubB64)
}

func deriveAESKey(shared []byte) ([]byte, error) {
	r := hkdf.New(sha256.New, shared, nil, []byte("apci-e2e-mvp-x25519-hkdf-aesgcm"))
	key := make([]byte, 32)
	if _, err := r.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func b64d(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

