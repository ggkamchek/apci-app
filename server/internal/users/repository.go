package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("username already exists")
	ErrChallenge     = errors.New("challenge not found or expired")
)

type User struct {
	ID               string
	Username         string
	Ed25519PublicKey []byte
}

type Challenge struct {
	ID        string
	UserID    string
	Challenge []byte
	ExpiresAt time.Time
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, username string, publicKey []byte) (User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (username, ed25519_public_key)
		VALUES ($1, $2)
		RETURNING id::text, username, ed25519_public_key
	`, username, publicKey)

	var user User
	if err := row.Scan(&user.ID, &user.Username, &user.Ed25519PublicKey); err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrAlreadyExists
		}
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, username, ed25519_public_key
		FROM users
		WHERE username = $1
	`, username)

	var user User
	if err := row.Scan(&user.ID, &user.Username, &user.Ed25519PublicKey); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *Repository) UpsertDevice(ctx context.Context, userID, deviceHash string, seenAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_devices (user_id, device_hash, first_seen_at, last_seen_at)
		VALUES ($1::uuid, $2, $3, $3)
		ON CONFLICT (user_id, device_hash)
		DO UPDATE SET last_seen_at = EXCLUDED.last_seen_at
	`, userID, deviceHash, seenAt)
	if err != nil {
		return fmt.Errorf("upsert device: %w", err)
	}
	return nil
}

func (r *Repository) CreateChallenge(ctx context.Context, userID string, challenge []byte, expiresAt time.Time) (string, error) {
	var challengeID string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO login_challenges (user_id, challenge, expires_at)
		VALUES ($1::uuid, $2, $3)
		RETURNING id::text
	`, userID, challenge, expiresAt).Scan(&challengeID)
	if err != nil {
		return "", fmt.Errorf("insert challenge: %w", err)
	}
	return challengeID, nil
}

func (r *Repository) GetChallenge(ctx context.Context, challengeID, username string, now time.Time) (Challenge, User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT c.id::text, c.user_id::text, c.challenge, c.expires_at,
		       u.id::text, u.username, u.ed25519_public_key
		FROM login_challenges c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = $1::uuid
		  AND u.username = $2
		  AND c.expires_at > $3
	`, challengeID, username, now)

	var challenge Challenge
	var user User
	if err := row.Scan(
		&challenge.ID,
		&challenge.UserID,
		&challenge.Challenge,
		&challenge.ExpiresAt,
		&user.ID,
		&user.Username,
		&user.Ed25519PublicKey,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Challenge{}, User{}, ErrChallenge
		}
		return Challenge{}, User{}, fmt.Errorf("get challenge: %w", err)
	}
	return challenge, user, nil
}

func (r *Repository) DeleteChallenge(ctx context.Context, challengeID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM login_challenges WHERE id = $1::uuid`, challengeID)
	if err != nil {
		return fmt.Errorf("delete challenge: %w", err)
	}
	return nil
}

func (r *Repository) CreateSession(ctx context.Context, userID string, expiresAt time.Time) (string, error) {
	var sessionID string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, expires_at)
		VALUES ($1::uuid, $2)
		RETURNING id::text
	`, userID, expiresAt).Scan(&sessionID)
	if err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return sessionID, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
