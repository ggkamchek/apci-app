package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationVersion int64 = 1

func Up(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	if err := ensureMigrationsTable(ctx, pool); err != nil {
		return err
	}

	applied, err := currentVersion(ctx, pool)
	if err != nil {
		return err
	}

	if applied >= migrationVersion {
		return nil
	}

	sqlPath := filepath.Join(migrationsDir, "001_users.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, migrationVersion); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}
	return nil
}

func currentVersion(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	var versions []int64
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return 0, fmt.Errorf("list migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return 0, fmt.Errorf("scan migration version: %w", err)
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate migrations: %w", err)
	}

	if len(versions) == 0 {
		return 0, nil
	}

	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	return versions[len(versions)-1], nil
}
