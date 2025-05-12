package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"real-time/backend/internal/repository"

	"real-time/backend/internal/model"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsHandler struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan WsMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	userRepo   repository.UserRepository
	mutex      sync.Mutex
}

type WsMessage struct {
	Type       string `json:"type"`
	Data       any    `json:"data"`
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
}

func NewWsHandler(userRepo repository.UserRepository) *WsHandler {
	return &WsHandler{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan WsMessage),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		userRepo:   userRepo,
	}
}

func (h *WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	go h.broadcastMessages()

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Handle connection
	h.register <- conn
	defer func() {
		h.unregister <- conn
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Handle incoming message
		var msg WsMessage
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}

		switch msg.Type {
		case "send_message":
			h.handleSendMessage(msg, conn)
		case "get_messages":
			h.handleGetMessages(msg, conn)
		}
	}
}

func (h *WsHandler) handleSendMessage(msg WsMessage, _ *websocket.Conn) {
	// Create message in database
	message := model.PrivateMessage{
		SenderID:   int64(msg.SenderID),
		ReceiverID: int64(msg.ReceiverID),
		Content:    msg.Data.(string),
	}

	err := h.userRepo.CreatePrivateMessage(message)
	if err != nil {
		log.Printf("error creating message: %v", err)
		return
	}

	// Create broadcast message
	broadcastMsg := WsMessage{
		Type:       "new_message",
		Data:       message.Content,
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
	}

	// Broadcast to all clients
	h.broadcast <- broadcastMsg
}

func (h *WsHandler) handleGetMessages(msg WsMessage, conn *websocket.Conn) {
	messages, err := h.userRepo.GetPrivateMessages(msg.SenderID, msg.ReceiverID)
	if err != nil {
		log.Printf("error getting messages: %v", err)
		return
	}

	// Send messages to client
	err = conn.WriteJSON(WsMessage{
		Type: "messages_history",
		Data: messages,
	})
	if err != nil {
		log.Printf("error sending messages: %v", err)
	}
}

// GetMessageHistory handles HTTP requests for getting message history
func GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	response := struct {
		Success  bool `json:"success"`
		Messages []struct {
			ID         int    `json:"id"`
			Content    string `json:"content"`
			SenderID   int    `json:"sender_id"`
			ReceiverID int    `json:"receiver_id"`
			CreatedAt  string `json:"created_at"`
		} `json:"messages"`
	}{
		Success: true,
		Messages: []struct {
			ID         int    `json:"id"`
			Content    string `json:"content"`
			SenderID   int    `json:"sender_id"`
			ReceiverID int    `json:"receiver_id"`
			CreatedAt  string `json:"created_at"`
		}{},
	}

	json.NewEncoder(w).Encode(response)
}

func (h *WsHandler) broadcastMessages() {
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
					log.Printf("error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}
