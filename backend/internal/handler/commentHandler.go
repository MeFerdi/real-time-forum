package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"real-time/backend/internal/middleware"
	"real-time/backend/internal/model"
	"real-time/backend/internal/repository"
)

type CommentHandler struct {
	commentRepo repository.CommentRepository
	userRepo    repository.UserRepository
	wsHandler   *WebSocketHandler
}

func NewCommentHandler(commentRepo repository.CommentRepository, userRepo repository.UserRepository, wsHandler *WebSocketHandler) *CommentHandler {
	return &CommentHandler{
		commentRepo: commentRepo,
		userRepo:    userRepo,
		wsHandler:   wsHandler,
	}
}

func (h *CommentHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/posts/comments/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(pathParts[0])
	if err != nil || postID <= 0 {
		writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Content == "" {
		writeError(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	comment := model.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: body.Content,
	}

	if err := h.commentRepo.AddComment(&comment); err != nil {
		writeError(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		writeError(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	comment.User = *user // Attach user info if needed by frontend

	h.wsHandler.BroadcastPostUpdate(WsMessage{
		Type:   "comment_added",
		Data:   comment,
		PostID: postID,
	})

	writeJSONResponse(w, http.StatusCreated, comment)
}

func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	comments, err := h.commentRepo.GetComments(postID)
	if err != nil {
		writeError(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	if comments == nil {
		comments = []*model.Comment{}
	}

	// Attach user info to each comment if needed
	for i := range comments {
		user, err := h.userRepo.GetByID(comments[i].UserID)
		if err == nil && user != nil {
			comments[i].User = *user
		}
	}

	writeJSONResponse(w, http.StatusOK, comments)
}
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/comments/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		writeError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(pathParts[0])
	if err != nil || commentID <= 0 {
		writeError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Content == "" {
		writeError(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	comment, err := h.commentRepo.GetCommentByID(commentID)
	if err != nil {
		writeError(w, "Failed to fetch comment", http.StatusInternalServerError)
		return
	}
	if comment == nil {
		writeError(w, "Comment not found", http.StatusNotFound)
		return
	}

	if comment.UserID != userID {
		writeError(w, "Unauthorized to update this comment", http.StatusForbidden)
		return
	}

	comment.Content = body.Content
	if err := h.commentRepo.UpdateComment(comment); err != nil {
		writeError(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		writeError(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	comment.User = *user

	h.wsHandler.BroadcastPostUpdate(WsMessage{
		Type:   "comment_updated",
		Data:   comment,
		PostID: comment.PostID,
	})

	writeJSONResponse(w, http.StatusOK, comment)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/comments/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		writeError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(pathParts[0])
	if err != nil || commentID <= 0 {
		writeError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	comment, err := h.commentRepo.GetCommentByID(commentID)
	if err != nil {
		writeError(w, "Failed to fetch comment", http.StatusInternalServerError)
		return
	}
	if comment == nil {
		writeError(w, "Comment not found", http.StatusNotFound)
		return
	}

	if comment.UserID != userID {
		writeError(w, "Unauthorized to delete this comment", http.StatusForbidden)
		return
	}

	if err := h.commentRepo.DeleteComment(commentID); err != nil {
		writeError(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	h.wsHandler.BroadcastPostUpdate(WsMessage{
		Type:   "comment_deleted",
		Data:   map[string]int{"id": commentID},
		PostID: comment.PostID,
	})

	w.WriteHeader(http.StatusOK)
}
func (h *CommentHandler) UpdateOrDeleteComment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		h.UpdateComment(w, r)
	case http.MethodDelete:
		h.DeleteComment(w, r)
	default:
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
