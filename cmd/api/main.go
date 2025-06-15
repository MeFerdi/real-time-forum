package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"real-time-forum/internal/auth"
	"real-time-forum/internal/database"
	"real-time-forum/internal/handlers"
	"real-time-forum/internal/websocket"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "./internal/database/forum.db"

func main() {
	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Initialize database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize schema
	if err := database.InitializeSchema(db); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()
	wsHandler := websocket.NewHandler(hub)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db)
	postHandler := handlers.NewPostHandler(db)

	// Register public routes
	http.HandleFunc("/api/register", userHandler.Register)
	http.HandleFunc("/api/login", userHandler.Login)
	http.HandleFunc("/api/logout", userHandler.Logout)
	http.HandleFunc("/api/categories", postHandler.ListCategories)

	// Register protected routes
	http.HandleFunc("/api/profile", auth.RequireAuth(userHandler.Profile, db))
	http.HandleFunc("/api/posts/create", auth.RequireAuth(postHandler.CreatePost, db))
	http.HandleFunc("/api/posts/get", auth.RequireAuth(postHandler.GetPost, db))
	http.HandleFunc("/api/posts", auth.RequireAuth(postHandler.ListPosts, db))
	http.HandleFunc("/api/comments", auth.RequireAuth(postHandler.CreateComment, db))
	http.HandleFunc("/api/posts/like", auth.RequireAuth(postHandler.LikePost, db))
	http.HandleFunc("/api/comments/like", auth.RequireAuth(postHandler.LikeComment, db))

	// Register WebSocket route
	http.HandleFunc("/api/ws", auth.RequireAuth(wsHandler.ServeWS, db))

	// Start HTTP server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
