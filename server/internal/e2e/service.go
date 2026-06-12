package e2e

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/black/apci-app/server/internal/auth"
	"github.com/black/apci-app/server/internal/users"
)

type UserLookup interface {
	GetByID(ctx context.Context, userID string) (users.User, error)
}

type Service struct {
	repo     *Repository
	users    UserLookup
	now      func() time.Time
}

func NewService(repo *Repository, users UserLookup) *Service {
	return &Service{
		repo:  repo,
		users: users,
		now:   time.Now,
	}
}

type UploadBundleInput struct {
	IdentityKey    []byte
	SignedPreKey   SignedPreKeyRecord
	OneTimePrekeys []OneTimePreKeyRecord
}

func (s *Service) UploadPreKeyBundle(ctx context.Context, input UploadBundleInput) (uint32, error) {
	session, ok := auth.UserSessionFromContext(ctx)
	if !ok {
		return 0, fmt.Errorf("session is required")
	}

	if err := validateX25519PublicKey(input.IdentityKey, "identity_key"); err != nil {
		return 0, err
	}
	if input.SignedPreKey.KeyID == 0 {
		return 0, fmt.Errorf("signed_prekey.key_id is required")
	}
	if err := validateX25519PublicKey(input.SignedPreKey.PublicKey, "signed_prekey.public_key"); err != nil {
		return 0, err
	}
	if err := validateNonEmptyBytes(input.SignedPreKey.Signature, "signed_prekey.signature"); err != nil {
		return 0, err
	}
	if input.SignedPreKey.ExpiresAt.IsZero() {
		return 0, fmt.Errorf("signed_prekey.expires_at is required")
	}
	if len(input.OneTimePrekeys) > maxOneTimePrekeys {
		return 0, fmt.Errorf("too many one_time_prekeys")
	}

	seen := make(map[uint32]struct{}, len(input.OneTimePrekeys))
	for i, key := range input.OneTimePrekeys {
		if key.KeyID == 0 {
			return 0, fmt.Errorf("one_time_prekeys[%d].key_id is required", i)
		}
		if _, dup := seen[key.KeyID]; dup {
			return 0, fmt.Errorf("duplicate one_time_prekey id %d", key.KeyID)
		}
		seen[key.KeyID] = struct{}{}
		if err := validateX25519PublicKey(key.PublicKey, fmt.Sprintf("one_time_prekeys[%d].public_key", i)); err != nil {
			return 0, err
		}
	}

	if err := s.repo.UpsertBundle(ctx, session.UserID, input.IdentityKey, input.SignedPreKey, input.OneTimePrekeys); err != nil {
		return 0, err
	}

	return uint32(len(input.OneTimePrekeys)), nil
}

func (s *Service) GetPreKeyBundle(ctx context.Context, userID string) (PreKeyBundle, error) {
	if _, ok := auth.UserSessionFromContext(ctx); !ok {
		return PreKeyBundle{}, fmt.Errorf("session is required")
	}
	if err := validateUserID(userID); err != nil {
		return PreKeyBundle{}, err
	}

	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return PreKeyBundle{}, err
	}

	return s.repo.GetPreKeyBundle(ctx, userID)
}

type SendMessageInput struct {
	RecipientID      string
	ChatID           string
	RatchetSessionID string
	Ciphertext       []byte
	RatchetHeader    []byte
}

func (s *Service) SendMessage(ctx context.Context, input SendMessageInput) (MessageRecord, error) {
	session, ok := auth.UserSessionFromContext(ctx)
	if !ok {
		return MessageRecord{}, fmt.Errorf("session is required")
	}

	if err := validateUserID(input.RecipientID); err != nil {
		return MessageRecord{}, err
	}
	if err := validateChatID(input.ChatID); err != nil {
		return MessageRecord{}, err
	}
	if err := validateRatchetSessionID(input.RatchetSessionID); err != nil {
		return MessageRecord{}, err
	}
	if err := validateCiphertext(input.Ciphertext); err != nil {
		return MessageRecord{}, err
	}
	if err := validateRatchetHeader(input.RatchetHeader); err != nil {
		return MessageRecord{}, err
	}

	recipientID := strings.TrimSpace(input.RecipientID)
	if recipientID == session.UserID {
		return MessageRecord{}, fmt.Errorf("cannot send message to yourself")
	}

	if _, err := s.users.GetByID(ctx, recipientID); err != nil {
		return MessageRecord{}, err
	}

	return s.repo.InsertMessage(ctx, MessageRecord{
		ChatID:           strings.TrimSpace(input.ChatID),
		RatchetSessionID: strings.TrimSpace(input.RatchetSessionID),
		SenderID:         session.UserID,
		RecipientID:      recipientID,
		Ciphertext:       input.Ciphertext,
		RatchetHeader:    input.RatchetHeader,
	})
}

func (s *Service) FetchMessages(ctx context.Context, sinceMessageID string, limit uint32) ([]MessageRecord, error) {
	session, ok := auth.UserSessionFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("session is required")
	}

	sinceMessageID = strings.TrimSpace(sinceMessageID)
	return s.repo.ListMessagesForRecipient(ctx, session.UserID, sinceMessageID, normalizeFetchLimit(limit))
}
