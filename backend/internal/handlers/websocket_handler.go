package handlers

import (
	"database/sql"
	"fmt"
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
	userClients map[int64]map[*Client]bool // Map user ID to client
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
		userClients: make(map[int64]map[*Client]bool),
		broadcast:   make(chan WSMessage),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		db:          db,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	log.Println("Hub started")
	for {
		select {
		case client := <-h.register:
			log.Printf("[Hub] Registering client for user %d", client.userID)
			h.mutex.Lock()
			h.clients[client] = true
			if h.userClients[client.userID] == nil {
				h.userClients[client.userID] = make(map[*Client]bool)
			}
			h.userClients[client.userID][client] = true
			var userSummary string
			for uid, conns := range h.userClients {
				userSummary += fmt.Sprintf("[User %d: %d connections] ", uid, len(conns))
			}
			log.Printf("[Hub] User %d connected. Current users: %s", client.userID, userSummary)
			h.mutex.Unlock()
			h.broadcastUserStatus(client.userID, client.user.Username, "online")
			h.sendOnlineUsers(client)

		case client := <-h.unregister:
			log.Printf("[Hub] Unregistering client for user %d", client.userID)
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if userSet, exists := h.userClients[client.userID]; exists {
					delete(userSet, client)
					if len(userSet) == 0 {
						delete(h.userClients, client.userID)
						h.broadcastUserStatus(client.userID, client.user.Username, "offline")
					}
				}
				close(client.send)
				log.Printf("[Hub] User %d disconnected", client.userID)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			log.Printf("[Hub] Broadcasting message of type %s", message.Type)
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					log.Printf("[Hub] Closing send channel for client user %d", client.userID)
					close(client.send)
					delete(h.clients, client)
					if userClients, exists := h.userClients[client.userID]; exists {
						delete(userClients, client)
						if len(userClients) == 0 {
							delete(h.userClients, client.userID)
						}
					}
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
	for userID, userSet := range h.userClients {
		for c := range userSet {
			if c.user != nil {
				onlineUsers = append(onlineUsers, UserStatusData{
					UserID:   userID,
					Username: c.user.Username,
					Status:   "online",
				})
				break // Only need one entry per user
			}
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
		h.mutex.RUnlock()
		h.mutex.Lock()
		delete(h.clients, client)
		if userSet, exists := h.userClients[client.userID]; exists {
			delete(userSet, client)
			if len(userSet) == 0 {
				delete(h.userClients, client.userID)
			}
		}
		h.mutex.Unlock()
		h.mutex.RLock()
	}
}

func (h *Hub) sendToUser(userID int64, message WSMessage) {
	log.Printf("[Hub] sendToUser: Sending message of type %s to user %d", message.Type, userID)
	h.mutex.RLock()
	clients, exists := h.userClients[userID]
	h.mutex.RUnlock()
	if exists {
		for client := range clients {
			select {
			case client.send <- message:
				log.Printf("[Hub] sendToUser: Message sent to client for user %d", userID)
			default:
				log.Printf("[Hub] sendToUser: Closing send channel for client user %d", userID)
				h.mutex.Lock()
				close(client.send)
				delete(h.clients, client)
				delete(h.userClients[userID], client)
				h.mutex.Unlock()
			}
		}
	} else {
		log.Printf("[Hub] sendToUser: No clients found for user %d", userID)
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
		log.Printf("[Client] readPump exiting for user %d", c.userID)
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
			log.Printf("[Client] WebSocket read error for user %d: %v", c.userID, err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Client] WebSocket unexpected close error: %v", err)
			}
			break
		}
		log.Printf("[Client] Received message of type %s from user %d", message.Type, c.userID)
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		log.Printf("[Client] writePump exiting for user %d", c.userID)
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				log.Printf("[Client] writePump: send channel closed for user %d", c.userID)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Printf("[Client] writePump: Sending message of type %s to user %d", message.Type, c.userID)
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("[Client] WebSocket write error for user %d: %v", c.userID, err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[Client] writePump: Ping error for user %d: %v", c.userID, err)
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
	log.Printf("[Client] handlePrivateMessage: user %d", c.userID)
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		log.Printf("[Client] handlePrivateMessage: Invalid message data for user %d", c.userID)
		c.sendError("Invalid message data")
		return
	}

	receiverID, ok := data["receiver_id"].(float64)
	if !ok {
		log.Printf("[Client] handlePrivateMessage: Invalid receiver ID for user %d", c.userID)
		c.sendError("Invalid receiver ID")
		return
	}

	content, ok := data["content"].(string)
	if !ok || content == "" {
		log.Printf("[Client] handlePrivateMessage: Invalid message content for user %d", c.userID)
		c.sendError("Invalid message content")
		return
	}

	log.Printf("[Client] handlePrivateMessage: Creating message from user %d to user %d", c.userID, int64(receiverID))
	privateMessage, err := models.CreatePrivateMessage(c.hub.db, c.userID, int64(receiverID), content)
	if err != nil {
		log.Printf("[Client] Error creating private message from user %d to user %d: %v", c.userID, int64(receiverID), err)
		c.sendError("Failed to send message")
		return
	}

	messageData := PrivateMessageData{
		ID:         privateMessage.ID,
		SenderID:   privateMessage.SenderID,
		ReceiverID: privateMessage.ReceiverID,
		Content:    privateMessage.Content,
		IsRead:     privateMessage.IsRead,
		CreatedAt:  privateMessage.CreatedAt,
		Sender:     c.user,
	}

	wsMessage := WSMessage{
		Type:      MessageTypePrivateMessage,
		Data:      messageData,
		Timestamp: time.Now(),
	}

	log.Printf("[Client] handlePrivateMessage: Sending to receiver %d and sender %d", int64(receiverID), c.userID)
	c.hub.sendToUser(int64(receiverID), wsMessage)
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
