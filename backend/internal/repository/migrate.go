package repository

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schema string

func MigrateDB(db *sql.DB) error {
	// Enable foreign keys for SQLite
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Execute schema
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	// Verify tables
	tables := []string{"users", "sessions", "posts", "comments", "private_messages"}
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %s LIMIT 1;", table)); err != nil {
			return fmt.Errorf("verification failed for table %s: %w", table, err)
		}
	}

	return nil
}
