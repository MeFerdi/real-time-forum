package routes

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
)

// SetupRoutes configures the HTTP router with all application routes
func SetupRoutes(db *sql.DB, cfg *config.Config) *http.Server {
	mux := http.NewServeMux()

	// Serve static assets (js, css, images, etc.)
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../frontend/static/js"))))
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../frontend/static/css"))))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	// Add other static asset folders as needed

	// Repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	postRepo := repository.NewPostRepository(db, categoryRepo)
	commentRepo := repository.NewCommentRepository(db)
	reactionRepo := repository.NewReactionRepository(db)

	// Handlers
	authHandler := handler.NewAuthHandler(cfg, userRepo, sessionRepo)
	wsHandler := handler.NewWebSocketHandler()
	postHandler := handler.NewPostHandler(postRepo, userRepo, categoryRepo, wsHandler)
	commentHandler := handler.NewCommentHandler(commentRepo, userRepo, wsHandler)
	reactionHandler := handler.NewReactionHandler(reactionRepo, wsHandler)
	messageHandler := handler.NewMessageHandler(userRepo, wsHandler)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)

	// Auth routes
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)

	// WebSocket route
	mux.Handle("/ws/messages", withAuth(sessionRepo, wsHandler.ServeHTTP))

	// Message routes
	mux.HandleFunc("/api/messages/history", withAuth(sessionRepo, messageHandler.GetMessageHistory))

	// Post routes
	mux.HandleFunc("/api/posts", withAuth(sessionRepo, postHandler.ListPosts))
	mux.HandleFunc("/api/posts/create", withAuth(sessionRepo, postHandler.CreatePost))
	mux.HandleFunc("/api/posts/by-user", withAuth(sessionRepo, postHandler.GetPostsByUserID))
	mux.HandleFunc("/api/posts/", withAuth(sessionRepo, postHandler.GetPost))
	mux.HandleFunc("/api/categories", categoryHandler.GetAllCategories)

	// Comment routes
	mux.HandleFunc("/api/posts/comments", withAuth(sessionRepo, commentHandler.GetComments))
	mux.HandleFunc("/api/posts/comments/", withAuth(sessionRepo, commentHandler.AddComment))
	mux.HandleFunc("/api/comments/", withAuth(sessionRepo, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			commentHandler.UpdateComment(w, r)
		case http.MethodDelete:
			commentHandler.DeleteComment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Reaction routes
	mux.HandleFunc("/api/posts/react/", withAuth(sessionRepo, reactionHandler.HandlePostReaction))

	// Catch-all: serve main.html for all other requests (SPA entry)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only serve main.html for GET requests that are not for static assets or API
		if r.Method != http.MethodGet ||
			strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/js/") ||
			strings.HasPrefix(r.URL.Path, "/css/") ||
			strings.HasPrefix(r.URL.Path, "/uploads/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "../frontend/static/main.html")
	})

	log.Println("Routes registered successfully")

	return &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: mux,
	}
}

// withAuth is middleware to authenticate requests using session tokens
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
