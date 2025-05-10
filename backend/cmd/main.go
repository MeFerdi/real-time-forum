package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"real-time/backend/internal/config"
	"real-time/backend/internal/handler"
	"real-time/backend/internal/repository"
	"real-time/backend/internal/utils"

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

	// Repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	postRepo := repository.NewPostRepository(db)

	// Handlers
	authHandler := handler.NewAuthHandler(cfg, userRepo, sessionRepo)
	postHandler := handler.NewPostHandler(postRepo, userRepo)

	// Auth routes
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)

	// websocket handler
	wsHandler := handler.NewWsHandler(userRepo)
	mux.HandleFunc("/ws/messages", wsHandler.ServeHTTP)
	mux.HandleFunc("/api/messages/history", withAuth(sessionRepo, handler.GetMessageHistory))

	// Post routes with auth middleware
	mux.HandleFunc("/api/posts", withAuth(sessionRepo, postHandler.ListPosts))
	mux.HandleFunc("/api/posts/create", withAuth(sessionRepo, postHandler.CreatePost))
	mux.HandleFunc("/api/posts/by-user", withAuth(sessionRepo, postHandler.GetPostsByUserID)) // New route
	mux.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/posts/") {
			http.NotFound(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/posts/")

		if strings.HasSuffix(path, "/comments") {
			withAuth(sessionRepo, postHandler.AddComment)(w, r)
			return
		}

		withAuth(sessionRepo, postHandler.GetPost)(w, r)
	})

	return mux
}

func withAuth(sessionRepo repository.SessionRepository, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := utils.GetAuthToken(r)
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := sessionRepo.Get(token)
		if err != nil {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next(w, r.WithContext(ctx))
	}
}
