package websocket

import (
	"log"
	"sync"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Map of user ID to their client connection
	clients map[int64]*Client

	// Lock for clients map
	mu sync.RWMutex

	// Channel for broadcasting messages to all clients
	broadcast chan Message

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			log.Printf("User %d connected. Total connected users: %d", client.userID, len(h.clients))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				log.Printf("User %d disconnected. Total connected users: %d", client.userID, len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.userID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message Message) {
	message.Timestamp = time.Now()
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client.userID)
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID int64, message Message) bool {
	message.Timestamp = time.Now()
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.clients[userID]; ok {
		client.send <- message
		return true
	}
	return false
}

// NotifyUserOnline broadcasts a user's online status to all clients
func (h *Hub) NotifyUserOnline(userID int64, online bool) {
	status := "offline"
	if online {
		status = "online"
	}

	h.Broadcast(Message{
		Type: MessageTypeStatus,
		Content: map[string]interface{}{
			"user_id": userID,
			"status":  status,
		},
		Timestamp: time.Now(),
	})
}

// IsUserOnline checks if a user is currently connected
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// GetOnlineUsers returns a list of all online user IDs
func (h *Hub) GetOnlineUsers() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]int64, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}
