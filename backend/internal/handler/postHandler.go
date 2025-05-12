package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"real-time/backend/internal/model"
	"real-time/backend/internal/repository"
)

type PostHandler struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

func NewPostHandler(postRepo repository.PostRepository, userRepo repository.UserRepository) *PostHandler {
	return &PostHandler{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(int64)
	post := &model.Post{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.postRepo.Create(post, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit

	posts, total, err := h.postRepo.List(offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := &model.PostListDTO{
		Posts:      make([]model.PostDTO, 0),
		TotalPosts: int64(total),
		Page:       page,
		PageSize:   limit,
	}

	for _, post := range posts {
		user, _ := h.userRepo.GetByID(post.UserID)
		if user != nil {
			response.Posts = append(response.Posts, post.ToDTO(user.ToDTO()))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) GetPostsByUserID(w http.ResponseWriter, r *http.Request) {
	// Get userID from query parameter
	userIDStr := r.URL.Query().Get("userID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	// Retrieve posts from repository
	posts, total, err := h.postRepo.GetByUserID(userID, offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create response
	response := &model.PostListDTO{
		Posts:      make([]model.PostDTO, 0),
		TotalPosts: int64(total),
		Page:       page,
		PageSize:   limit,
	}

	// Convert posts to DTOs with user information
	for _, post := range posts {
		user, err := h.userRepo.GetByID(post.UserID)
		if err != nil {
			continue
		}
		response.Posts = append(response.Posts, post.ToDTO(user.ToDTO()))
	}

	// Set response headers and encode JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path (e.g., /api/posts/123)
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/posts/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil || postID <= 0 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Retrieve post from repository
	post, err := h.postRepo.GetByID(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if post == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Retrieve user information
	user, err := h.userRepo.GetByID(post.UserID)
	if err != nil {
		http.Error(w, "Failed to retrieve user information", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Retrieve comments and categories
	comments, err := h.postRepo.GetComments(postID)
	if err != nil {
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	categories, err := h.postRepo.GetCategories(postID)
	if err != nil {
		http.Error(w, "Failed to retrieve categories", http.StatusInternalServerError)
		return
	}

	// Create response DTO
	response := post.ToDTO(user.ToDTO())
	response.Comments = make([]model.CommentDTO, 0, len(comments))
	for _, comment := range comments {
		commentUser, err := h.userRepo.GetByID(comment.UserID)
		if err != nil || commentUser == nil {
			continue
		}
		response.Comments = append(response.Comments, comment.ToDTO(commentUser.ToDTO()))
	}
	response.Categories = make([]model.CategoryDTO, 0, len(categories))
	for _, category := range categories {
		response.Categories = append(response.Categories, category.ToDTO())
	}

	// Set response headers and encode JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path (e.g., /api/posts/123/comments)
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/posts/"), "/")
	if len(pathParts) < 2 || pathParts[0] == "" {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil || postID <= 0 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Decode comment request
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	// Get userID from context (set by withAuth middleware)
	userID := r.Context().Value("userID").(int64)

	// Create comment
	comment := &model.Comment{
		PostID:    postID,
		UserID:    userID,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	// Save comment to repository
	if err := h.postRepo.AddComment(comment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve user information for response
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve user information", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Create response DTO
	response := comment.ToDTO(user.ToDTO())

	// Set response headers and encode JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
