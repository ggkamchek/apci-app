package e2e

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrBundleNotFound = errors.New("prekey bundle not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type SignedPreKeyRecord struct {
	KeyID      uint32
	PublicKey  []byte
	Signature  []byte
	ExpiresAt  time.Time
}

type OneTimePreKeyRecord struct {
	KeyID     uint32
	PublicKey []byte
}

type MessageRecord struct {
	ID               string
	ChatID           string
	RatchetSessionID string
	SenderID         string
	RecipientID      string
	Ciphertext       []byte
	RatchetHeader    []byte
	SentAt           time.Time
}

func (r *Repository) UpsertBundle(
	ctx context.Context,
	userID string,
	identityKey []byte,
	signed SignedPreKeyRecord,
	oneTime []OneTimePreKeyRecord,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO e2e_identity_keys (user_id, identity_key, updated_at)
		VALUES ($1::uuid, $2, now())
		ON CONFLICT (user_id) DO UPDATE
		SET identity_key = EXCLUDED.identity_key, updated_at = now()
	`, userID, identityKey); err != nil {
		return fmt.Errorf("upsert identity key: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO e2e_signed_prekeys (user_id, key_id, public_key, signature, expires_at, updated_at)
		VALUES ($1::uuid, $2, $3, $4, $5, now())
		ON CONFLICT (user_id) DO UPDATE
		SET key_id = EXCLUDED.key_id,
		    public_key = EXCLUDED.public_key,
		    signature = EXCLUDED.signature,
		    expires_at = EXCLUDED.expires_at,
		    updated_at = now()
	`, userID, signed.KeyID, signed.PublicKey, signed.Signature, signed.ExpiresAt); err != nil {
		return fmt.Errorf("upsert signed prekey: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM e2e_one_time_prekeys WHERE user_id = $1::uuid AND used_at IS NULL`, userID); err != nil {
		return fmt.Errorf("delete old one-time prekeys: %w", err)
	}

	for _, key := range oneTime {
		if _, err := tx.Exec(ctx, `
			INSERT INTO e2e_one_time_prekeys (user_id, key_id, public_key)
			VALUES ($1::uuid, $2, $3)
		`, userID, key.KeyID, key.PublicKey); err != nil {
			return fmt.Errorf("insert one-time prekey: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit bundle: %w", err)
	}
	return nil
}

type PreKeyBundle struct {
	UserID                string
	IdentityKey           []byte
	SignedPreKeyID        uint32
	SignedPreKey          []byte
	SignedPreKeySignature []byte
	OneTimePreKeyID       uint32
	OneTimePreKey         []byte
}

func (r *Repository) GetPreKeyBundle(ctx context.Context, userID string) (PreKeyBundle, error) {
	var bundle PreKeyBundle
	bundle.UserID = userID

	err := r.pool.QueryRow(ctx, `
		SELECT identity_key
		FROM e2e_identity_keys
		WHERE user_id = $1::uuid
	`, userID).Scan(&bundle.IdentityKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PreKeyBundle{}, ErrBundleNotFound
		}
		return PreKeyBundle{}, fmt.Errorf("get identity key: %w", err)
	}

	err = r.pool.QueryRow(ctx, `
		SELECT key_id, public_key, signature
		FROM e2e_signed_prekeys
		WHERE user_id = $1::uuid
	`, userID).Scan(&bundle.SignedPreKeyID, &bundle.SignedPreKey, &bundle.SignedPreKeySignature)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PreKeyBundle{}, ErrBundleNotFound
		}
		return PreKeyBundle{}, fmt.Errorf("get signed prekey: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return PreKeyBundle{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var oneTimeID string
	err = tx.QueryRow(ctx, `
		SELECT id::text, key_id, public_key
		FROM e2e_one_time_prekeys
		WHERE user_id = $1::uuid AND used_at IS NULL
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, userID).Scan(&oneTimeID, &bundle.OneTimePreKeyID, &bundle.OneTimePreKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return PreKeyBundle{}, fmt.Errorf("get one-time prekey: %w", err)
	}
	if err == nil {
		if _, err := tx.Exec(ctx, `
			UPDATE e2e_one_time_prekeys SET used_at = now() WHERE id = $1::uuid
		`, oneTimeID); err != nil {
			return PreKeyBundle{}, fmt.Errorf("mark one-time prekey used: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return PreKeyBundle{}, fmt.Errorf("commit prekey bundle: %w", err)
	}

	return bundle, nil
}

func (r *Repository) InsertMessage(ctx context.Context, msg MessageRecord) (MessageRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO e2e_messages (
			chat_id, ratchet_session_id, sender_id, recipient_id, ciphertext, ratchet_header
		) VALUES ($1, $2, $3::uuid, $4::uuid, $5, $6)
		RETURNING id::text, sent_at
	`, msg.ChatID, msg.RatchetSessionID, msg.SenderID, msg.RecipientID, msg.Ciphertext, msg.RatchetHeader)

	if err := row.Scan(&msg.ID, &msg.SentAt); err != nil {
		return MessageRecord{}, fmt.Errorf("insert message: %w", err)
	}
	return msg, nil
}

func (r *Repository) ListMessagesForRecipient(ctx context.Context, recipientID, sinceMessageID string, limit int) ([]MessageRecord, error) {
	query := `
		SELECT id::text, chat_id, ratchet_session_id, sender_id::text, recipient_id::text,
		       ciphertext, ratchet_header, sent_at
		FROM e2e_messages
		WHERE recipient_id = $1::uuid
	`
	args := []any{recipientID}

	if sinceMessageID != "" {
		query += `
			AND sent_at > COALESCE(
				(SELECT sent_at FROM e2e_messages WHERE id = $2::uuid AND recipient_id = $1::uuid),
				'-infinity'::timestamptz
			)
		`
		args = append(args, sinceMessageID)
	}

	query += ` ORDER BY sent_at ASC, id ASC LIMIT ` + fmt.Sprintf("%d", limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []MessageRecord
	for rows.Next() {
		var msg MessageRecord
		if err := rows.Scan(
			&msg.ID, &msg.ChatID, &msg.RatchetSessionID,
			&msg.SenderID, &msg.RecipientID,
			&msg.Ciphertext, &msg.RatchetHeader, &msg.SentAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}
	return messages, nil
}
