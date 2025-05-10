package handler

import (
	"log"
	"net/http"
	"real-time/backend/internal/repository"
	"sync"

	"encoding/json"
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
	Type       string      `json:"type"`
	Data       interface{} `json:"data"`
	SenderID   int         `json:"sender_id"`
	ReceiverID int         `json:"receiver_id"`
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

func (h *WsHandler) handleSendMessage(msg WsMessage, conn *websocket.Conn) {
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

	// Broadcast to receiver
	h.broadcast <- WsMessage{
		Type:       "new_message",
		Data:       message,
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
	}
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
