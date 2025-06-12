package handler

import (
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
	"real-time/backend/internal/middleware"
)

// writeJSONResponse writes a JSON response with the given status code
func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeError writes an error response with the given status code
func writeError(w http.ResponseWriter, message string, status int) {
	log.Printf("Error: %s", message)
	http.Error(w, message, status)
}

type PostHandler struct {
	postRepo     repository.PostRepository
	userRepo     repository.UserRepository
	categoryRepo repository.CategoryRepository
	wsHandler    *WebSocketHandler
}

func NewPostHandler(postRepo repository.PostRepository, userRepo repository.UserRepository, categoryRepo repository.CategoryRepository, wsHandler *WebSocketHandler) *PostHandler {
	return &PostHandler{
		postRepo:     postRepo,
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		wsHandler:    wsHandler,
	}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		writeError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	var req model.CreatePostRequest
	req.Title = r.FormValue("title")
	req.Content = r.FormValue("content")

	var categoryIDs []int64
	if categoryStr := r.FormValue("categories"); categoryStr != "" {
		for _, idStr := range strings.Split(categoryStr, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				writeError(w, "Invalid category ID", http.StatusBadRequest)
				return
			}
			categoryIDs = append(categoryIDs, id)
		}
	}

	var imageURL string
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		uploadsDir := "uploads"
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			writeError(w, "Failed to create uploads directory", http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
		filepath := path.Join(uploadsDir, filename)
		out, err := os.Create(filepath)
		if err != nil {
			writeError(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			writeError(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		imageURL = "/uploads/" + filename
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	post := model.Post{
		UserID:   userID,
		Title:    req.Title,
		Content:  req.Content,
		ImageURL: imageURL,
	}

	if err := h.postRepo.Create(&post); err != nil {
		writeError(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	if len(categoryIDs) > 0 {
		if err := h.categoryRepo.AddPostCategories(post.ID, categoryIDs); err != nil {
			writeError(w, "Failed to add categories", http.StatusInternalServerError)
			return
		}
	}

	createdPost, err := h.postRepo.GetByID(post.ID)
	if err != nil {
		writeError(w, "Failed to fetch created post", http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.GetByID(createdPost.UserID)
	if err != nil {
		writeError(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	h.wsHandler.BroadcastPostUpdate(WsMessage{
		Type:   "post_created",
		Data:   createdPost.ToDTO(user.ToDTO()),
		PostID: createdPost.ID,
	})

	writeJSONResponse(w, http.StatusCreated, createdPost.ToDTO(user.ToDTO()))
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
		writeError(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	response := &model.PostListDTO{
		Posts:      make([]model.PostDTO, 0, len(posts)),
		TotalPosts: int64(total),
		Page:       page,
		PageSize:   limit,
	}

	for _, post := range posts {
		user, err := h.userRepo.GetByID(post.UserID)
		if err != nil || user == nil {
			continue
		}
		response.Posts = append(response.Posts, post.ToDTO(user.ToDTO()))
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *PostHandler) GetPostsByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.URL.Query().Get("userID"))
	if err != nil || userID <= 0 {
		writeError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	posts, total, err := h.postRepo.GetByUserID(userID, offset, limit)
	if err != nil {
		writeError(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	response := &model.PostListDTO{
		Posts:      make([]model.PostDTO, 0, len(posts)),
		TotalPosts: int64(total),
		Page:       page,
		PageSize:   limit,
	}

	for _, post := range posts {
		user, err := h.userRepo.GetByID(post.UserID)
		if err != nil || user == nil {
			continue
		}
		response.Posts = append(response.Posts, post.ToDTO(user.ToDTO()))
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/posts/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil || postID <= 0 {
		writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.postRepo.GetByID(int(postID))
	if err != nil {
		writeError(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}
	if post == nil {
		writeError(w, "Post not found", http.StatusNotFound)
		return
	}

	user, err := h.userRepo.GetByID(post.UserID)
	if err != nil || user == nil {
		writeError(w, "Failed to retrieve user information", http.StatusInternalServerError)
		return
	}

	response := post.ToDTO(user.ToDTO())
	writeJSONResponse(w, http.StatusOK, response)
}
