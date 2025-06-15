package websocket

import (
	"log"
	"net/http"

	"real-time-forum/internal/auth"
)

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeWS handles WebSocket requests from clients
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is already connected
	if h.hub.IsUserOnline(userID) {
		http.Error(w, "User already connected", http.StatusConflict)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		hub:    h.hub,
		conn:   conn,
		send:   make(chan Message, 256),
		userID: userID,
	}
	client.hub.register <- client

	// Start goroutines for pumping messages
	go client.writePump()
	go client.readPump()
}
