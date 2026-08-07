package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseURL string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return conn, nil
}

// Migrate applies every .sql file in the migrations directory, in filename
// order, tracking which have already run in schema_migrations so each one
// executes exactly once — required for migrations that aren't naturally
// idempotent (e.g. renames), not just the CREATE TABLE IF NOT EXISTS ones.
func Migrate(conn *sql.DB, migrationsDir string) error {
	if _, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		var alreadyApplied bool
		row := conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`, entry.Name())
		if err := row.Scan(&alreadyApplied); err != nil {
			return fmt.Errorf("check migration %s: %w", entry.Name(), err)
		}
		if alreadyApplied {
			continue
		}

		path := filepath.Join(migrationsDir, entry.Name())
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		tx, err := conn.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (filename) VALUES ($1)`, entry.Name()); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}
