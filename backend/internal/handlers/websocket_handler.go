package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"real-time-forum/backend/internal/auth"
	"real-time-forum/backend/internal/models"

	"github.com/gorilla/websocket"
)

// Client represents a connected websocket client
type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

// WebSocketHandler manages WebSocket connections and messages
type WebSocketHandler struct {
	db       *sql.DB
	upgrader websocket.Upgrader
	clients  map[int64]*Client // map userID to client
	mu       sync.RWMutex
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	SenderID  int64           `json:"sender_id,omitempty"`
	Timestamp string          `json:"timestamp"`
}

// ChatMessage represents a chat message payload
type ChatMessage struct {
	ReceiverID int64  `json:"receiver_id"`
	Content    string `json:"content"`
}

func NewWebSocketHandler(db *sql.DB) *WebSocketHandler {
	return &WebSocketHandler{
		db: db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		clients: make(map[int64]*Client),
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	// Create new client
	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	// Register client
	h.registerClient(client)

	// Start client routines
	go h.readPump(client)
	go h.writePump(client)

	// Notify other users that this user is online
	h.broadcastUserStatus(userID, true)
}

func (h *WebSocketHandler) registerClient(client *Client) {
	h.mu.Lock()
	h.clients[client.UserID] = client
	h.mu.Unlock()
}

func (h *WebSocketHandler) unregisterClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client.UserID]; ok {
		delete(h.clients, client.UserID)
		close(client.Send)
	}
	h.mu.Unlock()

	// Notify other users that this user is offline
	h.broadcastUserStatus(client.UserID, false)
}

func (h *WebSocketHandler) readPump(client *Client) {
	defer func() {
		h.unregisterClient(client)
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Parse the message
		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Handle different message types
		switch wsMessage.Type {
		case "chat":
			var chatMsg ChatMessage
			if err := json.Unmarshal(wsMessage.Data, &chatMsg); err != nil {
				log.Printf("Error unmarshaling chat message: %v", err)
				continue
			}
			h.handleChatMessage(client, chatMsg)
		}
	}
}

func (h *WebSocketHandler) writePump(client *Client) {
	defer func() {
		client.Conn.Close()
	}()

	for message := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error writing message: %v", err)
			return
		}
	}
	client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (h *WebSocketHandler) handleChatMessage(client *Client, msg ChatMessage) {
	// Save message to database
	dbMsg, err := models.CreateMessage(h.db, client.UserID, msg.ReceiverID, msg.Content)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		return
	}

	// Prepare message for WebSocket
	wsMessage := WebSocketMessage{
		Type:      "chat",
		SenderID:  client.UserID,
		Timestamp: dbMsg.CreatedAt.Format(time.RFC3339),
	}

	messageData, err := json.Marshal(dbMsg)
	if err != nil {
		log.Printf("Error marshaling message data: %v", err)
		return
	}
	wsMessage.Data = messageData

	// Marshal the complete message
	response, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	// Send to recipient if online
	h.mu.RLock()
	if recipient, ok := h.clients[msg.ReceiverID]; ok {
		recipient.Send <- response
	}
	h.mu.RUnlock()

	// Send back to sender for confirmation
	client.Send <- response
}

func (h *WebSocketHandler) broadcastUserStatus(userID int64, online bool) {
	status := struct {
		UserID int64 `json:"user_id"`
		Online bool  `json:"online"`
	}{
		UserID: userID,
		Online: online,
	}

	statusData, err := json.Marshal(status)
	if err != nil {
		log.Printf("Error marshaling status: %v", err)
		return
	}

	wsMessage := WebSocketMessage{
		Type: "status",
		Data: statusData,
	}

	response, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	// Broadcast to all connected clients
	h.mu.RLock()
	for _, client := range h.clients {
		select {
		case client.Send <- response:
		default:
			close(client.Send)
			delete(h.clients, client.UserID)
		}
	}
	h.mu.RUnlock()
}
