package users

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ed25519"
)

type Service struct {
	repo         *Repository
	challengeTTL time.Duration
	sessionTTL   time.Duration
	now          func() time.Time
}

func NewService(repo *Repository, challengeTTL, sessionTTL time.Duration) *Service {
	return &Service{
		repo:         repo,
		challengeTTL: challengeTTL,
		sessionTTL:   sessionTTL,
		now:          time.Now,
	}
}

type RegisterInput struct {
	Username         string
	Ed25519PublicKey []byte
	DeviceHash       string
}

type RegisterResult struct {
	UserID   string
	Username string
}

type BeginLoginResult struct {
	ChallengeID string
	Challenge   []byte
}

type CompleteLoginInput struct {
	Username    string
	ChallengeID string
	Signature   []byte
	DeviceHash  string
}

type CompleteLoginResult struct {
	SessionID string
	UserID    string
	Username  string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (RegisterResult, error) {
	username := normalizeUsername(input.Username)
	if err := validateUsername(username); err != nil {
		return RegisterResult{}, err
	}
	if err := validatePublicKey(input.Ed25519PublicKey); err != nil {
		return RegisterResult{}, err
	}
	if err := validateDeviceHash(input.DeviceHash); err != nil {
		return RegisterResult{}, err
	}

	user, err := s.repo.CreateUser(ctx, username, input.Ed25519PublicKey)
	if err != nil {
		return RegisterResult{}, err
	}

	now := s.now().UTC()
	if err := s.repo.UpsertDevice(ctx, user.ID, input.DeviceHash, now); err != nil {
		return RegisterResult{}, err
	}

	return RegisterResult{
		UserID:   user.ID,
		Username: user.Username,
	}, nil
}

func (s *Service) BeginLogin(ctx context.Context, username string) (BeginLoginResult, error) {
	username = normalizeUsername(username)
	if err := validateUsername(username); err != nil {
		return BeginLoginResult{}, err
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return BeginLoginResult{}, err
	}

	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return BeginLoginResult{}, fmt.Errorf("generate challenge: %w", err)
	}

	now := s.now().UTC()
	challengeID, err := s.repo.CreateChallenge(ctx, user.ID, challenge, now.Add(s.challengeTTL))
	if err != nil {
		return BeginLoginResult{}, err
	}

	return BeginLoginResult{
		ChallengeID: challengeID,
		Challenge:   challenge,
	}, nil
}

func (s *Service) CompleteLogin(ctx context.Context, input CompleteLoginInput) (CompleteLoginResult, error) {
	username := normalizeUsername(input.Username)
	if err := validateUsername(username); err != nil {
		return CompleteLoginResult{}, err
	}
	if err := validateSignature(input.Signature); err != nil {
		return CompleteLoginResult{}, err
	}
	if err := validateDeviceHash(input.DeviceHash); err != nil {
		return CompleteLoginResult{}, err
	}
	if input.ChallengeID == "" {
		return CompleteLoginResult{}, fmt.Errorf("challenge_id is required")
	}

	now := s.now().UTC()
	challenge, user, err := s.repo.GetChallenge(ctx, input.ChallengeID, username, now)
	if err != nil {
		return CompleteLoginResult{}, err
	}

	if !ed25519.Verify(user.Ed25519PublicKey, challenge.Challenge, input.Signature) {
		return CompleteLoginResult{}, fmt.Errorf("invalid signature")
	}

	if err := s.repo.DeleteChallenge(ctx, challenge.ID); err != nil {
		return CompleteLoginResult{}, err
	}

	sessionID, err := s.repo.CreateSession(ctx, user.ID, now.Add(s.sessionTTL))
	if err != nil {
		return CompleteLoginResult{}, err
	}

	if err := s.repo.UpsertDevice(ctx, user.ID, input.DeviceHash, now); err != nil {
		return CompleteLoginResult{}, err
	}

	return CompleteLoginResult{
		SessionID: sessionID,
		UserID:    user.ID,
		Username:  user.Username,
	}, nil
}

func (s *Service) ValidateSession(ctx context.Context, sessionID string) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", fmt.Errorf("session_id is required")
	}
	return s.repo.GetSessionUserID(ctx, sessionID, s.now().UTC())
}

func normalizeUsername(username string) string {
	return strings.TrimSpace(username)
}
