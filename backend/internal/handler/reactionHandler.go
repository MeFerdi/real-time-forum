package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"real-time/backend/internal/model"
	"real-time/backend/internal/repository"
)

type ReactionHandler struct {
	reactionRepo repository.ReactionRepository
	wsHandler    *WebSocketHandler
}

func NewReactionHandler(reactionRepo repository.ReactionRepository, wsHandler *WebSocketHandler) *ReactionHandler {
	return &ReactionHandler{
		reactionRepo: reactionRepo,
		wsHandler:    wsHandler,
	}
}

func (h *ReactionHandler) HandlePostReaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/posts/react/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(pathParts[0])
	if err != nil || postID <= 0 {
		writeError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Type != "like" && body.Type != "dislike" && body.Type != "none" {
		writeError(w, "Invalid reaction type", http.StatusBadRequest)
		return
	}

	if body.Type == "none" {
		if err := h.reactionRepo.RemoveReaction(postID, userID); err != nil {
			writeError(w, "Failed to remove reaction", http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.reactionRepo.AddReaction(postID, userID, body.Type); err != nil {
			writeError(w, "Failed to add reaction", http.StatusInternalServerError)
			return
		}
	}

	likes, dislikes, err := h.reactionRepo.GetPostReactions(postID)
	if err != nil {
		writeError(w, "Failed to fetch reactions", http.StatusInternalServerError)
		return
	}

	userReaction, err := h.reactionRepo.GetUserReaction(postID, userID)
	if err != nil {
		writeError(w, "Failed to fetch user reaction", http.StatusInternalServerError)
		return
	}

	response := model.PostDTO{
		ID:           postID,
		LikeCount:    likes,
		DislikeCount: dislikes,
		UserReaction: userReaction,
	}

	h.wsHandler.BroadcastPostUpdate(WsMessage{
		Type:   "post_reactions_updated",
		Data:   response,
		PostID: postID,
	})

	writeJSONResponse(w, http.StatusOK, response)
}
