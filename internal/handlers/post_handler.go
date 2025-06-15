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

// ListPosts handles retrieving all posts, optionally filtered by category
func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for category filter
	categoryIDStr := r.URL.Query().Get("category_id")
	var posts []models.Post
	var err error

	if categoryIDStr != "" {
		// If category_id is provided, filter by category
		categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		// Verify category exists
		var exists bool
		err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE id = ?)", categoryID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking category existence: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Category not found", http.StatusNotFound)
			return
		}

		posts, err = models.ListPostsByCategory(h.db, categoryID)
	} else {
		// If no category_id, get all posts
		posts, err = models.ListPosts(h.db)
	}

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

// LikePost handles liking/unliking a post
func (h *PostHandler) LikePost(w http.ResponseWriter, r *http.Request) {
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

	if err := models.LikePost(h.db, postID, userID); err != nil {
		log.Printf("Error liking post: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get updated like count
	likeCount, err := models.GetPostLikes(h.db, postID)
	if err != nil {
		log.Printf("Error getting post likes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get current user's like status
	hasLiked, err := models.HasUserLikedPost(h.db, postID, userID)
	if err != nil {
		log.Printf("Error checking user like status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"like_count": likeCount,
		"has_liked":  hasLiked,
	})
}

// LikeComment handles liking/unliking a comment
func (h *PostHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get comment ID from URL query parameters
	commentIDStr := r.URL.Query().Get("comment_id")
	if commentIDStr == "" {
		http.Error(w, "Missing comment ID", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := models.LikeComment(h.db, commentID, userID); err != nil {
		log.Printf("Error liking comment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get updated like count
	likeCount, err := models.GetCommentLikes(h.db, commentID)
	if err != nil {
		log.Printf("Error getting comment likes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get current user's like status
	hasLiked, err := models.HasUserLikedComment(h.db, commentID, userID)
	if err != nil {
		log.Printf("Error checking user like status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"like_count": likeCount,
		"has_liked":  hasLiked,
	})
}
