package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"real-time-forum/backend/internal/auth"
	"real-time-forum/backend/internal/database"
	"real-time-forum/backend/internal/handlers"
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

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db)
	postHandler := handlers.NewPostHandler(db)
	messageHandler := handlers.NewMessageHandler(db)

	// Initialize WebSocket hub
	hub := handlers.NewHub(db)
	go hub.Run()

	// Create router
	mux := http.NewServeMux()

	// Register public API routes
	mux.HandleFunc("/api/register", userHandler.Register)
	mux.HandleFunc("/api/login", userHandler.Login)
	mux.HandleFunc("/api/logout", userHandler.Logout)
	mux.HandleFunc("/api/categories", postHandler.ListCategories)

	// Register protected routes
	mux.HandleFunc("/api/profile", auth.RequireAuth(userHandler.Profile, db))
	mux.HandleFunc("/api/posts/create", auth.RequireAuth(postHandler.CreatePost, db))
	mux.HandleFunc("/api/posts/get", auth.RequireAuth(postHandler.GetPost, db))
	mux.HandleFunc("/api/posts", auth.RequireAuth(postHandler.ListPosts, db))
	mux.HandleFunc("/api/posts/", auth.RequireAuth(postHandler.HandlePostRoutes, db))
	mux.HandleFunc("/api/posts/like", auth.RequireAuth(postHandler.LikePost, db))
	mux.HandleFunc("/api/comments/like", auth.RequireAuth(postHandler.LikeComment, db))

	// Register WebSocket and message routes
	mux.HandleFunc("/ws", hub.WebSocketHandler)
	mux.HandleFunc("/api/messages/conversations", auth.RequireAuth(messageHandler.GetConversations, db))
	mux.HandleFunc("/api/messages/history", auth.RequireAuth(messageHandler.GetConversationHistory, db))
	mux.HandleFunc("/api/messages/mark-read", auth.RequireAuth(messageHandler.MarkAsRead, db))
	mux.HandleFunc("/api/messages/users", auth.RequireAuth(messageHandler.GetAllUsers, db))
	mux.HandleFunc("/api/messages/send", auth.RequireAuth(messageHandler.SendMessage, db))

	// Create a custom handler that wraps the file server for SPA support
	fs := http.FileServer(http.Dir("../frontend"))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested file exists
		filePath := "../frontend" + r.URL.Path
		if _, err := os.Stat(filePath); os.IsNotExist(err) && r.URL.Path != "/" {
			// File doesn't exist and it's not the root, serve index.html for SPA routing
			http.ServeFile(w, r, "../frontend/index.html")
			return
		}

		// Serve static files normally
		fs.ServeHTTP(w, r)
	}))

	// Start HTTP server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
