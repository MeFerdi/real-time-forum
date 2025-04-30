package repository

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func MigrateDB(db *sql.DB) error {
	// Read migration files
	migrations, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Filter and sort .up.sql files
	var upMigrations []fs.DirEntry
	for _, m := range migrations {
		if strings.HasSuffix(m.Name(), ".up.sql") {
			upMigrations = append(upMigrations, m)
		}
	}
	sort.Slice(upMigrations, func(i, j int) bool {
		return upMigrations[i].Name() < upMigrations[j].Name()
	})

	// Execute each migration in a transaction
	for _, migration := range upMigrations {
		path := filepath.Join("migrations", migration.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", path, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", path, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", path, err)
		}

		fmt.Printf("Applied migration: %s\n", path)
	}

	return nil
}
