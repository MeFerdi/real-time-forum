package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"real-time-forum/backend/internal/auth"
	"real-time-forum/backend/internal/models"
)

type MessageHandler struct {
	db  *sql.DB
	hub *Hub
}

func NewMessageHandler(db *sql.DB, hub *Hub) *MessageHandler {
	return &MessageHandler{db: db, hub: hub}
}

// GetConversations retrieves all conversations for the authenticated user
func (h *MessageHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversations, err := models.GetUserConversations(h.db, userID)
	if err != nil {
		log.Printf("Error getting conversations: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// GetConversationHistory retrieves message history between the authenticated user and another user
func (h *MessageHandler) GetConversationHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get other user ID from query parameters
	otherUserIDStr := r.URL.Query().Get("user_id")
	if otherUserIDStr == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	otherUserID, err := strconv.ParseInt(otherUserIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
		return
	}

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	messages, err := models.GetConversationHistory(h.db, userID, otherUserID, limit, offset)
	if err != nil {
		log.Printf("Error getting conversation history: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// MarkAsRead marks messages from a specific user as read
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		SenderID int64 `json:"sender_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := models.MarkMessagesAsRead(h.db, userID, req.SenderID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// GetAllUsers retrieves all users for the chat user list
func (h *MessageHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, err := models.GetAllUsers(h.db, userID)
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// SendMessage handles sending a private message via HTTP (fallback for non-WebSocket clients)
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		ReceiverID int64  `json:"receiver_id"`
		Content    string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	message, err := models.CreatePrivateMessage(h.db, userID, req.ReceiverID, req.Content)
	if err != nil {
		log.Printf("Error creating private message: %v", err)
		switch err.Error() {
		case "message content cannot be empty":
			http.Error(w, err.Error(), http.StatusBadRequest)
		case "cannot send message to yourself":
			http.Error(w, err.Error(), http.StatusBadRequest)
		case "receiver not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Broadcast message via WebSocket if hub is available
	if h.hub != nil {
		// Get sender info for WebSocket message
		sender, err := models.GetUserByID(h.db, userID)
		if err == nil {
			// Create WebSocket message data
			messageData := PrivateMessageData{
				ID:         message.ID,
				SenderID:   message.SenderID,
				ReceiverID: message.ReceiverID,
				Content:    message.Content,
				IsRead:     message.IsRead,
				CreatedAt:  message.CreatedAt,
				Sender:     sender,
			}

			// Create WebSocket message
			wsMessage := WSMessage{
				Type:      MessageTypePrivateMessage,
				Data:      messageData,
				Timestamp: time.Now(),
			}

			// Send to receiver
			h.hub.sendToUser(req.ReceiverID, wsMessage)
			// Send to sender (for multi-tab support)
			h.hub.sendToUser(userID, wsMessage)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}
