package handler

import (
	"net/http"
	"real-time/backend/internal/repository"
	"sync"

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
