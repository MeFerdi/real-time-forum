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

	// Static file handler
	staticFileHandler := http.FileServer(http.Dir("../frontend/static"))
	staticHandler := http.StripPrefix("/", staticFileHandler)
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(r.URL.Path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		}
		staticHandler.ServeHTTP(w, r)
	}))

	// Uploads handler
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

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
