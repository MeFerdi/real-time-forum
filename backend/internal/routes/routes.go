package routes

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"real-time/backend/internal/config"
	"real-time/backend/internal/handler"
	"real-time/backend/internal/repository"
	"real-time/backend/internal/utils"
)

// SetupRoutes configures the HTTP router with all application routes
func SetupRoutes(db *sql.DB, cfg *config.Config) *http.Server {
	mux := http.NewServeMux()

	// Resolve absolute paths for static directories
	staticBase, err := filepath.Abs("../frontend/static")
	if err != nil {
		log.Fatalf("Failed to resolve static base: %v", err)
	}
	jsDir := filepath.Join(staticBase, "js")
	cssDir := filepath.Join(staticBase, "css")
	uploadsDir, err := filepath.Abs("uploads")
	if err != nil {
		log.Fatalf("Failed to resolve uploads dir: %v", err)
	}

	// Secure static file server (no directory listing, only allowed extensions)
	secureFileServer := func(absDir string, allowedExts map[string]string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			relPath := strings.TrimPrefix(r.URL.Path, "/")
			absPath := filepath.Join(absDir, relPath)
			if !strings.HasPrefix(absPath, absDir) {
				http.NotFound(w, r)
				return
			}
			fileInfo, err := os.Stat(absPath)
			if err != nil || fileInfo.IsDir() {
				http.NotFound(w, r)
				return
			}
			ext := strings.ToLower(filepath.Ext(absPath))
			contentType, ok := allowedExts[ext]
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", contentType)
			http.ServeFile(w, r, absPath)
		})
	}

	// Allowed extensions and MIME types
	jsExts := map[string]string{".js": "application/javascript"}
	cssExts := map[string]string{".css": "text/css"}
	uploadExts := map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
	}

	// Serve static assets using absolute paths
	mux.Handle("/js/", http.StripPrefix("/js/", secureFileServer(jsDir, jsExts)))
	mux.Handle("/css/", http.StripPrefix("/css/", secureFileServer(cssDir, cssExts)))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", secureFileServer(uploadsDir, uploadExts)))

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
	mux.HandleFunc("/api/auth/me", withAuth(sessionRepo, authHandler.Me))

	// WebSocket route
	mux.Handle("/ws/messages", withAuth(sessionRepo, wsHandler.ServeHTTP))

	// Message routes
	mux.HandleFunc("/api/messages/history", withAuth(sessionRepo, messageHandler.GetMessageHistory))

	// Post routes
	mux.Handle("/api/posts", withAuth(sessionRepo, postHandler.ListPosts))
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

	// SPA catch-all: serve main.html for all other GET requests
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet ||
			strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/js/") ||
			strings.HasPrefix(r.URL.Path, "/css/") ||
			strings.HasPrefix(r.URL.Path, "/uploads/") ||
			strings.Contains(r.URL.Path, "..") {
			http.NotFound(w, r)
			return
		}
		mainHTMLPath := filepath.Join(staticBase, "main.html")
		if _, err := os.Stat(mainHTMLPath); err != nil {
			http.Error(w, "Page Not Found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, mainHTMLPath)
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
