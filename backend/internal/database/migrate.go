package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// embed bakes SQL migration files into the Go binary at compile time.
// That means migrations are available even after deployment without reading loose files.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

func RunMigrations(ctx context.Context, db *sql.DB) error {
	// schema_migrations is the small bookkeeping table that remembers applied migrations.
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return err
	}

	files, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		return err
	}
	// Sorting keeps migrations deterministic: 001 runs before 002, and so on.
	sort.Strings(files)

	for _, file := range files {
		version := migrationVersion(file)
		applied, err := migrationApplied(ctx, db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := migrationFiles.ReadFile(file)
		if err != nil {
			return err
		}

		if err := runMigration(ctx, db, version, string(sqlBytes)); err != nil {
			return err
		}
	}

	return nil
}

func migrationApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM schema_migrations WHERE version = $1
		)
	`, version).Scan(&exists)
	return exists, err
}

func runMigration(ctx context.Context, db *sql.DB, version string, query string) error {
	// A transaction makes each migration all-or-nothing.
	// If the SQL fails, the version is not recorded.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("migration %s failed: %w", version, err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO schema_migrations (version) VALUES ($1)
	`, version); err != nil {
		return err
	}

	return tx.Commit()
}

func migrationVersion(path string) string {
	parts := strings.Split(path, "/")
	return strings.TrimSuffix(parts[len(parts)-1], ".sql")
}

// Why this file exists:
// This file gives Memora a reproducible way to create/update database schema.
// Instead of manually editing Neon tables, every schema change should become a versioned SQL file.
// The migration table answers: "Has this schema change already run on this database?"
