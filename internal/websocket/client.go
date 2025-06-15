package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client represents a single WebSocket connection
type Client struct {
	hub *Hub

	// The WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan Message

	// User ID associated with this connection
	userID int64
}

// MessageType represents different types of WebSocket messages
const (
	MessageTypeChat   = "chat"
	MessageTypeNotify = "notification"
	MessageTypeStatus = "status"
	MessageTypeTyping = "typing"
)

// Message represents a WebSocket message with different possible types
type Message struct {
	Type       string      `json:"type"`              // Message type (chat, notification, status, typing)
	Content    interface{} `json:"content,omitempty"` // Message content (can be string or structured data)
	SenderID   int64       `json:"sender_id"`         // ID of the sending user
	ReceiverID int64       `json:"receiver_id"`       // ID of the receiving user (optional)
	Timestamp  time.Time   `json:"timestamp"`         // Message timestamp
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Set sender ID and timestamp
		msg.SenderID = c.userID
		msg.Timestamp = time.Now()

		// Handle message based on type
		switch msg.Type {
		case MessageTypeChat:
			// Handle private chat message
			if msg.ReceiverID != 0 {
				if targetClient := c.hub.clients[msg.ReceiverID]; targetClient != nil {
					targetClient.send <- msg
				}
			}
		case MessageTypeTyping:
			// Handle typing indicator
			if msg.ReceiverID != 0 {
				if targetClient := c.hub.clients[msg.ReceiverID]; targetClient != nil {
					targetClient.send <- msg
				}
			}
		case MessageTypeStatus:
			// Broadcast online status to all clients
			c.hub.broadcast <- msg
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteJSON(message)
			if err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
