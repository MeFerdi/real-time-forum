package main

import (
	"database/sql"
	"log"

	"real-time/backend/internal/config"
	"real-time/backend/internal/repository"
	"real-time/backend/internal/routes"

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

	log.Printf("Starting server on %s", cfg.ServerAddress)
	router := routes.SetupRoutes(db, cfg)
	log.Fatal(router.ListenAndServe())
}
