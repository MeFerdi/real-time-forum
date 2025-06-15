package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"real-time-forum/internal/auth"
	"real-time-forum/internal/models"
)

type PostHandler struct {
	db *sql.DB
}

func NewPostHandler(db *sql.DB) *PostHandler {
	return &PostHandler{db: db}
}

// CreatePost handles post creation
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserID(r)
	if !ok {
		log.Printf("Failed to get user ID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("Creating post for user ID: %d", userID)

	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("Received post request: %+v", req)

	post, err := models.CreatePost(h.db, userID, req)
	if err != nil {
		switch err {
		case models.ErrEmptyTitle, models.ErrEmptyContent, models.ErrNoCategories, models.ErrInvalidCategory:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			log.Printf("Error creating post: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// GetPost handles retrieving a single post
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get post ID from URL query parameters
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := models.GetPostByID(h.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting post: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// ListCategories handles retrieving all categories
func (h *PostHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := models.ListCategories(h.db)
	if err != nil {
		log.Printf("Error listing categories: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// ListPosts handles retrieving all posts
func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	posts, err := models.ListPosts(h.db)
	if err != nil {
		log.Printf("Error listing posts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// CreateComment handles comment creation for a post
func (h *PostHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get post ID from URL query parameters
	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	comment, err := models.CreateComment(h.db, postID, userID, req)
	if err != nil {
		switch err {
		case models.ErrEmptyComment:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case models.ErrPostNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			log.Printf("Error creating comment: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}
