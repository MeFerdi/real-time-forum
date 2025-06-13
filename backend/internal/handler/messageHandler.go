package handler

import (
	"log"
	"net/http"
	"strconv"

	"real-time/backend/internal/middleware"
	"real-time/backend/internal/model"
	"real-time/backend/internal/repository"

	"github.com/gorilla/websocket"
)

type MessageHandler struct {
	userRepo  repository.UserRepository
	wsHandler *WebSocketHandler
}

func NewMessageHandler(userRepo repository.UserRepository, wsHandler *WebSocketHandler) *MessageHandler {
	return &MessageHandler{
		userRepo:  userRepo,
		wsHandler: wsHandler,
	}
}

func (h *MessageHandler) HandleWebSocketMessage(msg WsMessage, conn *websocket.Conn) {
	switch msg.Type {
	case "send_message":
		h.handleSendMessage(msg, conn)
	case "get_messages":
		h.handleGetMessages(msg, conn)
	case "comment_added", "comment_updated", "comment_deleted", "post_reactions_updated":
		h.wsHandler.BroadcastPostUpdate(msg)
	default:
		log.Printf("Unknown WebSocket message type: %s", msg.Type)
	}
}

func (h *MessageHandler) handleSendMessage(msg WsMessage, _ *websocket.Conn) {
	content, ok := msg.Data.(string)
	if !ok || content == "" {
		log.Println("Invalid message content")
		return
	}

	message := model.PrivateMessage{
		SenderID:   int64(msg.SenderID),
		ReceiverID: int64(msg.ReceiverID),
		Content:    content,
	}

	if err := h.userRepo.CreatePrivateMessage(message); err != nil {
		log.Printf("Failed to create private message: %v", err)
		return
	}

	broadcastMsg := WsMessage{
		Type:       "new_message",
		Data:       message,
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
	}
	h.wsHandler.BroadcastPostUpdate(broadcastMsg)
}

func (h *MessageHandler) handleGetMessages(msg WsMessage, conn *websocket.Conn) {
	messages, err := h.userRepo.GetPrivateMessages(msg.SenderID, msg.ReceiverID)
	if err != nil {
		log.Printf("Failed to get private messages: %v", err)
		return
	}

	if err := conn.WriteJSON(WsMessage{
		Type: "messages_history",
		Data: messages,
	}); err != nil {
		log.Printf("Failed to send messages history: %v", err)
	}
}

func (h *MessageHandler) GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil || userID <= 0 {
		writeError(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || currentUserID == 0 {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	messages, err := h.userRepo.GetPrivateMessages(currentUserID, userID)
	if err != nil {
		writeError(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}
