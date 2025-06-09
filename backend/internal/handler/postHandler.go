package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"real-time/backend/internal/model"
	"real-time/backend/internal/repository"
)

// Helper function to get userID from context
func getUserIDFromContext(ctx context.Context) int {
	if userID, ok := ctx.Value("userID").(int64); ok {
		return int(userID)
	}
	return 0
}

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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	title := r.FormValue("title")
	content := r.FormValue("content")
	categoriesStr := r.FormValue("categories")

	log.Printf("Received post data - Title: %s, Content length: %d, Categories: %s",
		title, len(content), categoriesStr)

	// Process categories - temporarily skip category processing until we implement proper category management
	var categoryIDs []int64 // Empty slice for now

	// Handle image upload
	var imageURL string
	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("No image file uploaded: %v", err)
	} else {
		defer file.Close()

		// Create uploads directory if it doesn't exist
		if err := os.MkdirAll("uploads", 0755); err != nil {
			log.Printf("Error creating uploads directory: %v", err)
			http.Error(w, "Failed to create uploads directory", http.StatusInternalServerError)
			return
		}

		// Generate unique filename and ensure uploads directory exists
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
		uploadsDir := "uploads"
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			log.Printf("Error creating uploads directory: %v", err)
			http.Error(w, "Failed to create uploads directory", http.StatusInternalServerError)
			return
		}

		filepath := path.Join(uploadsDir, filename)
		log.Printf("Saving file to: %s", filepath)

		// Save file
		out, err := os.Create(filepath)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			http.Error(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			log.Printf("Error saving file: %v", err)
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		imageURL = "/uploads/" + filename
		log.Printf("Image uploaded successfully: %s", imageURL)
	}

	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		log.Printf("Error: UserID is 0, context may be missing user information")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	post := &model.Post{
		UserID:   userID,
		Title:    title,
		Content:  content,
		ImageURL: imageURL,
	}

	log.Printf("Creating post with UserID: %d", userID)
	if err := h.postRepo.Create(post, categoryIDs); err != nil {
		log.Printf("Error creating post in repository: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	log.Printf("Post created successfully with ID: %d", post.ID)

	// Get the created post with user information
	createdPost, err := h.postRepo.GetByID(post.ID)
	if err != nil {
		log.Printf("Error fetching created post: %v", err)
		http.Error(w, "Failed to fetch created post", http.StatusInternalServerError)
		return
	}

	// Get user info
	user, err := h.userRepo.GetByID(createdPost.UserID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPost.ToDTO(user.ToDTO()))
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit

	log.Printf("Fetching posts - Page: %d, Limit: %d, Offset: %d", page, limit, offset)
	posts, total, err := h.postRepo.List(offset, limit)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
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
		user, err := h.userRepo.GetByID(post.UserID)
		if err != nil {
			log.Printf("Error fetching user for post %d: %v", post.ID, err)
			continue
		}
		if user != nil {
			response.Posts = append(response.Posts, post.ToDTO(user.ToDTO()))
		}
	}

	log.Printf("Returning %d posts", len(response.Posts))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) GetPostsByUserID(w http.ResponseWriter, r *http.Request) {
	// Get userID from query parameter
	userIDStr := r.URL.Query().Get("userID")
	userID, err := strconv.Atoi(userIDStr)
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
	post, err := h.postRepo.GetByID(int(postID))
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
	comments, err := h.postRepo.GetComments(post.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
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

	// Set response headers and encode JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {
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

	// Get userID from context
	contextUserID := r.Context().Value("userID").(int64)

	comment := &model.Comment{
		PostID:    int(postID),
		UserID:    int(contextUserID),
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := h.postRepo.AddComment(comment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve user information for response
	user, err := h.userRepo.GetByID(int(contextUserID))
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	offset := 0
	limit := 10

	posts, _, err := h.postRepo.List(offset, limit)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Create response DTOs with user information
	var response []model.PostDTO
	for _, post := range posts {
		user, err := h.userRepo.GetByID(post.UserID)
		if err != nil {
			continue
		}
		response = append(response, post.ToDTO(user.ToDTO()))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PostHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Query().Get("postId")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	comments, err := h.postRepo.GetComments(postID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	// Create response DTOs with user information
	var response []model.CommentDTO
	for _, comment := range comments {
		user, err := h.userRepo.GetByID(comment.UserID)
		if err != nil {
			continue
		}
		response = append(response, comment.ToDTO(user.ToDTO()))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
