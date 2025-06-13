package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func MigrateDB(db *sql.DB) error {
	log.Printf("Starting database migration...")

	statements := []string{
		// Enable foreign keys
		"PRAGMA foreign_keys = ON;",

		// Users table
		`CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            uuid TEXT UNIQUE NOT NULL,
            nickname TEXT UNIQUE NOT NULL,
            email TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            first_name TEXT NOT NULL,
            last_name TEXT NOT NULL,
            age INTEGER NOT NULL,
            gender TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            last_online DATETIME,
            is_online BOOLEAN DEFAULT FALSE
        );`,

		// Sessions table
		`CREATE TABLE IF NOT EXISTS sessions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            token TEXT UNIQUE NOT NULL,
            expires_at DATETIME NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
        );`,

		// Posts table
		`CREATE TABLE IF NOT EXISTS posts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            title TEXT,
            content TEXT NOT NULL,
            image_url TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
        );`,

		// Comments table
		`CREATE TABLE IF NOT EXISTS comments (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            post_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            content TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
            FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
        );`,

		// Private messages table
		`CREATE TABLE IF NOT EXISTS private_messages (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            sender_id INTEGER NOT NULL,
            receiver_id INTEGER NOT NULL,
            content TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            is_read BOOLEAN DEFAULT FALSE,
            FOREIGN KEY(sender_id) REFERENCES users(id) ON DELETE CASCADE,
            FOREIGN KEY(receiver_id) REFERENCES users(id) ON DELETE CASCADE
        );`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);`,

		// Categories table
		`CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL
        );`,

		// Default categories
		`INSERT OR IGNORE INTO categories (name) VALUES
        ('Technology'),
        ('Science'),
        ('Business'),
        ('Finance'),
        ('Health'),
        ('Education'),
        ('Career'),
        ('News'),
        ('Environment'),
        ('Innovation');`,

		// Post-Categories (many-to-many)
		`CREATE TABLE IF NOT EXISTS post_categories (
            post_id INTEGER NOT NULL,
            category_id INTEGER NOT NULL,
            PRIMARY KEY (post_id, category_id),
            FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
            FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE CASCADE
        );`,

		// More indexes
		`CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_private_messages_sender ON private_messages(sender_id);`,
		`CREATE INDEX IF NOT EXISTS idx_private_messages_receiver ON private_messages(receiver_id);`,
		`CREATE INDEX IF NOT EXISTS idx_post_categories_post_id ON post_categories(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_post_categories_category_id ON post_categories(category_id);`,

		// Post reactions table
		`CREATE TABLE IF NOT EXISTS post_reactions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            post_id INTEGER NOT NULL,
            user_id INTEGER NOT NULL,
            reaction_type TEXT NOT NULL CHECK(reaction_type IN ('like', 'dislike')),
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
            FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
            UNIQUE(post_id, user_id)
        );`,
		`CREATE INDEX IF NOT EXISTS idx_post_reactions_post_id ON post_reactions(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_post_reactions_user_id ON post_reactions(user_id);`,
	}

	// Execute each statement
	for _, stmt := range statements {
		preview := stmt
		if len(stmt) > 50 {
			preview = stmt[:50] + "..."
		}
		log.Printf("Executing statement: %s", preview)
		if _, err := db.Exec(stmt); err != nil {
			log.Printf("Error executing statement: %v", err)
			return fmt.Errorf("failed to execute schema statement: %w", err)
		}
	}

	// Verify tables
	tables := []string{
		"users", "sessions", "posts", "comments", "private_messages",
		"categories", "post_categories", "post_reactions",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err == sql.ErrNoRows {
			log.Printf("Table %s does not exist after migration", table)
			return fmt.Errorf("table %s was not created", table)
		}
		if err != nil {
			log.Printf("Error verifying table %s: %v", table, err)
			return fmt.Errorf("verification failed for table %s: %w", table, err)
		}
		log.Printf("Verified table: %s", table)
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := filepath.Join("uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("Error creating uploads directory: %v", err)
		return fmt.Errorf("failed to create uploads directory: %w", err)
	}

	log.Printf("Database migration completed successfully")
	return nil
}
