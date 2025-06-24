package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"real-time-forum/backend/internal/models"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin in development
	},
}

// Message types
const (
	MessageTypePrivateMessage = "private_message"
	MessageTypeUserStatus     = "user_status"
	MessageTypeTyping         = "typing"
	MessageTypeOnlineUsers    = "online_users"
	MessageTypeError          = "error"
)

// WebSocket message structure
type WSMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Private message data structure
type PrivateMessageData struct {
	ID         int64        `json:"id"`
	SenderID   int64        `json:"sender_id"`
	ReceiverID int64        `json:"receiver_id"`
	Content    string       `json:"content"`
	IsRead     bool         `json:"is_read"`
	CreatedAt  time.Time    `json:"created_at"`
	Sender     *models.User `json:"sender,omitempty"`
}

// User status data structure
type UserStatusData struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Status   string `json:"status"` // "online" or "offline"
}

// Client represents a WebSocket connection
type Client struct {
	conn   *websocket.Conn
	send   chan WSMessage
	hub    *Hub
	userID int64
	user   *models.User
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	clients     map[*Client]bool
	userClients map[int64]*Client // Map user ID to client
	broadcast   chan WSMessage
	register    chan *Client
	unregister  chan *Client
	db          *sql.DB
	mutex       sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub(db *sql.DB) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[int64]*Client),
		broadcast:   make(chan WSMessage),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		db:          db,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.userClients[client.userID] = client
			h.mutex.Unlock()

			log.Printf("User %d connected", client.userID)

			// Notify all clients about user coming online
			h.broadcastUserStatus(client.userID, client.user.Username, "online")

			// Send current online users to the new client
			h.sendOnlineUsers(client)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.userClients, client.userID)
				close(client.send)
				h.mutex.Unlock()

				log.Printf("User %d disconnected", client.userID)

				// Notify all clients about user going offline
				h.broadcastUserStatus(client.userID, client.user.Username, "offline")
			} else {
				h.mutex.Unlock()
			}

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					delete(h.userClients, client.userID)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// broadcastUserStatus sends user status update to all connected clients
func (h *Hub) broadcastUserStatus(userID int64, username, status string) {
	message := WSMessage{
		Type: MessageTypeUserStatus,
		Data: UserStatusData{
			UserID:   userID,
			Username: username,
			Status:   status,
		},
		Timestamp: time.Now(),
	}
	h.broadcast <- message
}

// sendOnlineUsers sends the list of online users to a specific client
func (h *Hub) sendOnlineUsers(client *Client) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var onlineUsers []UserStatusData
	for _, c := range h.userClients {
		if c.user != nil {
			onlineUsers = append(onlineUsers, UserStatusData{
				UserID:   c.userID,
				Username: c.user.Username,
				Status:   "online",
			})
		}
	}

	message := WSMessage{
		Type:      MessageTypeOnlineUsers,
		Data:      onlineUsers,
		Timestamp: time.Now(),
	}

	select {
	case client.send <- message:
	default:
		close(client.send)
		delete(h.clients, client)
		delete(h.userClients, client.userID)
	}
}

// sendToUser sends a message to a specific user
func (h *Hub) sendToUser(userID int64, message WSMessage) {
	h.mutex.RLock()
	client, exists := h.userClients[userID]
	h.mutex.RUnlock()

	if exists {
		select {
		case client.send <- message:
		default:
			h.mutex.Lock()
			close(client.send)
			delete(h.clients, client)
			delete(h.userClients, client.userID)
			h.mutex.Unlock()
		}
	}
}

// WebSocketHandler handles WebSocket connections
func (h *Hub) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := models.GetUserBySessionToken(h.db, cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create new client
	client := &Client{
		conn:   conn,
		send:   make(chan WSMessage, 256),
		hub:    h,
		userID: user.ID,
		user:   user,
	}

	// Register client with hub
	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WSMessage
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(message WSMessage) {
	switch message.Type {
	case MessageTypePrivateMessage:
		c.handlePrivateMessage(message)
	case MessageTypeTyping:
		c.handleTyping(message)
	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

// handlePrivateMessage processes private message sending
func (c *Client) handlePrivateMessage(message WSMessage) {
	// Parse message data
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		c.sendError("Invalid message data")
		return
	}

	receiverID, ok := data["receiver_id"].(float64)
	if !ok {
		c.sendError("Invalid receiver ID")
		return
	}

	content, ok := data["content"].(string)
	if !ok || content == "" {
		c.sendError("Invalid message content")
		return
	}

	// Save message to database
	privateMessage, err := models.CreatePrivateMessage(c.hub.db, c.userID, int64(receiverID), content)
	if err != nil {
		log.Printf("Error creating private message: %v", err)
		c.sendError("Failed to send message")
		return
	}

	// Create message data with sender info
	messageData := PrivateMessageData{
		ID:         privateMessage.ID,
		SenderID:   privateMessage.SenderID,
		ReceiverID: privateMessage.ReceiverID,
		Content:    privateMessage.Content,
		IsRead:     privateMessage.IsRead,
		CreatedAt:  privateMessage.CreatedAt,
		Sender:     c.user,
	}

	// Send message to receiver
	wsMessage := WSMessage{
		Type:      MessageTypePrivateMessage,
		Data:      messageData,
		Timestamp: time.Now(),
	}

	c.hub.sendToUser(int64(receiverID), wsMessage)

	// Also send to sender for real-time update (in case they have multiple tabs open)
	c.hub.sendToUser(c.userID, wsMessage)
}

// handleTyping processes typing indicators
func (c *Client) handleTyping(message WSMessage) {
	// Parse message data
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		return
	}

	receiverID, ok := data["receiver_id"].(float64)
	if !ok {
		return
	}

	// Forward typing indicator to receiver
	typingMessage := WSMessage{
		Type: MessageTypeTyping,
		Data: map[string]interface{}{
			"sender_id": c.userID,
			"username":  c.user.Username,
		},
		Timestamp: time.Now(),
	}

	c.hub.sendToUser(int64(receiverID), typingMessage)
}

// sendError sends an error message to the client
func (c *Client) sendError(errorMsg string) {
	message := WSMessage{
		Type: MessageTypeError,
		Data: map[string]string{
			"error": errorMsg,
		},
		Timestamp: time.Now(),
	}

	select {
	case c.send <- message:
	default:
		close(c.send)
	}
}
