package main

import (
	"database/sql"
	"log"
	"net/http"

	"real-time/backend/internal/config"
	"real-time/backend/internal/handler"
	"real-time/backend/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg := config.MustLoad()

	db, err := sql.Open("sqlite3", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := repository.MigrateDB(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	router := setupRouter(db, cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}

func setupRouter(db *sql.DB, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	authHandler := handler.NewAuthHandler(cfg, userRepo, sessionRepo)

	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)

	return mux
}
