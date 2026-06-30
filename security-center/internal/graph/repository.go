package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM ingested_events WHERE event_id = $1::uuid)
	`, eventID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check event: %w", err)
	}
	return exists, nil
}

func (r *Repository) MarkEventProcessed(ctx context.Context, eventID, eventType string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ingested_events (event_id, event_type)
		VALUES ($1::uuid, $2)
	`, eventID, eventType)
	if err != nil {
		return fmt.Errorf("record event: %w", err)
	}
	return nil
}

func (r *Repository) TouchAccount(ctx context.Context, userID string, seenAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO account_nodes (user_id, first_seen_at, last_seen_at)
		VALUES ($1::uuid, $2, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET last_seen_at = EXCLUDED.last_seen_at
	`, userID, seenAt)
	if err != nil {
		return fmt.Errorf("touch account: %w", err)
	}
	return nil
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

func (r *Repository) FindUsersByDevice(ctx context.Context, deviceHash, excludeUserID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id::text
		FROM user_devices
		WHERE device_hash = $1 AND user_id <> $2::uuid
	`, deviceHash, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("find users by device: %w", err)
	}
	defer rows.Close()

	return scanUserIDs(rows)
}

func (r *Repository) UpsertLink(ctx context.Context, userA, userB, linkType string, weight int, seenAt time.Time) error {
	accountA, accountB := orderAccounts(userA, userB)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO account_links (account_a, account_b, link_type, weight, detected_at, last_seen_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $5)
		ON CONFLICT (account_a, account_b, link_type)
		DO UPDATE SET last_seen_at = EXCLUDED.last_seen_at
	`, accountA, accountB, linkType, weight, seenAt)
	if err != nil {
		return fmt.Errorf("upsert link: %w", err)
	}
	return nil
}

func (r *Repository) ListLinks(ctx context.Context) ([]Link, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT account_a::text, account_b::text, link_type, weight, detected_at
		FROM account_links
		ORDER BY detected_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		var detectedAt time.Time
		if err := rows.Scan(&link.AccountA, &link.AccountB, &link.LinkType, &link.Weight, &detectedAt); err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}
		link.DetectedAt = detectedAt.UTC().Format(time.RFC3339)
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate links: %w", err)
	}
	return links, nil
}

func scanUserIDs(rows pgx.Rows) ([]string, error) {
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user ids: %w", err)
	}
	return ids, nil
}

func orderAccounts(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}
