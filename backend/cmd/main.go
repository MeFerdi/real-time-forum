package main

import (
	"database/sql"
	"log"
	"net/http"

	"real-time/backend/internal/config"
	"real-time/backend/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	// Initialize database
	db, err := sql.Open("sqlite3", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Run migrations
	if err := repository.MigrateDB(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.ServerAddress,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		// 	Handler:      setupRouter(db, cfg), // The Application router setup function *Commented for now, i'm still doing setup lol
	}

	log.Printf("Server starting on %s", cfg.ServerAddress)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// func setupRouter(db *sql.DB, cfg *config.Config) http.Handler {
// 	return nil
// }
