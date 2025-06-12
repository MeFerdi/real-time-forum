package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"real-time/backend/internal/middleware"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust for production
	},
}

type WsMessage struct {
	Type       string      `json:"type"`
	Data       interface{} `json:"data"`
	SenderID   int         `json:"sender_id,omitempty"`
	ReceiverID int         `json:"receiver_id,omitempty"`
	PostID     int         `json:"post_id,omitempty"`
}

type WebSocketHandler struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan WsMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan WsMessage),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	go h.broadcastMessages()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	h.register <- conn
	defer func() {
		h.unregister <- conn
		conn.Close()
	}()

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || userID == 0 {
		log.Println("Unauthorized WebSocket connection attempt")
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg WsMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal WebSocket message: %v", err)
			continue
		}

		msg.SenderID = userID
		h.broadcast <- msg
	}
}

func (h *WebSocketHandler) BroadcastPostUpdate(message WsMessage) {
	if h != nil {
		h.broadcast <- message
	}
}

func (h *WebSocketHandler) broadcastMessages() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				if err := client.WriteJSON(message); err != nil {
					log.Printf("Failed to broadcast message: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}
