package routes

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"real-time/backend/internal/config"
	"real-time/backend/internal/handler"
	"real-time/backend/internal/middleware"
	"real-time/backend/internal/repository"
)

func SetupRoutes(db *sql.DB, cfg *config.Config) *http.Server {
	mux := http.NewServeMux()

	staticBase, _ := filepath.Abs("../frontend/static")
	jsDir := filepath.Join(staticBase, "js")
	cssDir := filepath.Join(staticBase, "css")
	uploadsDir, _ := filepath.Abs("uploads")

	serveStatic := func(dir string, exts map[string]string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rel := strings.TrimPrefix(r.URL.Path, "/")
			abs := filepath.Join(dir, rel)
			if !strings.HasPrefix(abs, dir) {
				http.NotFound(w, r)
				return
			}
			info, err := os.Stat(abs)
			if err != nil || info.IsDir() {
				http.NotFound(w, r)
				return
			}
			ext := strings.ToLower(filepath.Ext(abs))
			if ct, ok := exts[ext]; ok {
				w.Header().Set("Content-Type", ct)
				http.ServeFile(w, r, abs)
				return
			}
			http.NotFound(w, r)
		})
	}
	jsExts := map[string]string{".js": "application/javascript"}
	cssExts := map[string]string{".css": "text/css"}
	uploadExts := map[string]string{".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif"}

	mux.Handle("/js/", http.StripPrefix("/js/", serveStatic(jsDir, jsExts)))
	mux.Handle("/css/", http.StripPrefix("/css/", serveStatic(cssDir, cssExts)))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", serveStatic(uploadsDir, uploadExts)))

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	postRepo := repository.NewPostRepository(db, categoryRepo)
	commentRepo := repository.NewCommentRepository(db)
	reactionRepo := repository.NewReactionRepository(db)

	authHandler := handler.NewAuthHandler(cfg, userRepo, sessionRepo)
	wsHandler := handler.NewWebSocketHandler()
	postHandler := handler.NewPostHandler(postRepo, userRepo, categoryRepo, wsHandler)
	commentHandler := handler.NewCommentHandler(commentRepo, userRepo, wsHandler)
	reactionHandler := handler.NewReactionHandler(reactionRepo, wsHandler)
	messageHandler := handler.NewMessageHandler(userRepo, wsHandler)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)

	// Auth
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)
	mux.Handle("/api/auth/me", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(authHandler.Me)))

	// WebSocket
	mux.Handle("/ws/messages", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(wsHandler.ServeHTTP)))

	// Messages
	mux.Handle("/api/messages/history", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(messageHandler.GetMessageHistory)))

	// Posts
	mux.Handle("/api/posts", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(postHandler.ListPosts)))
	mux.Handle("/api/posts/create", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(postHandler.CreatePost)))
	mux.Handle("/api/posts/by-user", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(postHandler.GetPostsByUserID)))
	mux.Handle("/api/posts/", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(postHandler.GetPost)))

	// Categories
	mux.Handle("/api/categories", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(categoryHandler.GetAllCategories)))

	// Comments
	mux.Handle("/api/posts/comments", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(commentHandler.GetComments)))
	mux.Handle("/api/posts/comments/", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(commentHandler.AddComment)))
	mux.Handle("/api/comments/", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(commentHandler.UpdateOrDeleteComment)))

	// Reactions
	mux.Handle("/api/posts/react/", middleware.SessionAuthMiddleware(sessionRepo)(http.HandlerFunc(reactionHandler.HandlePostReaction)))

	// SPA catch-all
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
