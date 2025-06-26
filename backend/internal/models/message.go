package models

import (
	"database/sql"
	"errors"
	"time"
)

// PrivateMessage represents a private message between users
type PrivateMessage struct {
	ID         int64     `json:"id"`
	SenderID   int64     `json:"sender_id"`
	ReceiverID int64     `json:"receiver_id"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
	Sender     *User     `json:"sender,omitempty"`
	Receiver   *User     `json:"receiver,omitempty"`
}

// Conversation represents a conversation between two users
type Conversation struct {
	UserID      int64           `json:"user_id"`
	Username    string          `json:"username"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	LastMessage *PrivateMessage `json:"last_message"`
	UnreadCount int             `json:"unread_count"`
	IsOnline    bool            `json:"is_online"`
}

// CreatePrivateMessage creates a new private message
func CreatePrivateMessage(db *sql.DB, senderID, receiverID int64, content string) (*PrivateMessage, error) {
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	if senderID == receiverID {
		return nil, errors.New("cannot send message to yourself")
	}

	// Check if receiver exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", receiverID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("receiver not found")
	}

	// Insert the message
	result, err := db.Exec(`
		INSERT INTO messages (sender_id, receiver_id, content, created_at)
		VALUES (?, ?, ?, ?)`,
		senderID, receiverID, content, time.Now())
	if err != nil {
		return nil, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Get the created message with sender info
	message, err := GetPrivateMessage(db, messageID)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// GetPrivateMessage retrieves a private message by ID
func GetPrivateMessage(db *sql.DB, messageID int64) (*PrivateMessage, error) {
	query := `
		SELECT m.id, m.sender_id, m.receiver_id, m.content, m.is_read, m.created_at,
		       u.username, u.first_name, u.last_name
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.id = ?`

	var message PrivateMessage
	var sender User
	err := db.QueryRow(query, messageID).Scan(
		&message.ID,
		&message.SenderID,
		&message.ReceiverID,
		&message.Content,
		&message.IsRead,
		&message.CreatedAt,
		&sender.Username,
		&sender.FirstName,
		&sender.LastName,
	)
	if err != nil {
		return nil, err
	}

	sender.ID = message.SenderID
	message.Sender = &sender

	return &message, nil
}

// GetConversationHistory retrieves message history between two users with pagination
func GetConversationHistory(db *sql.DB, userID1, userID2 int64, limit, offset int) ([]PrivateMessage, error) {
	query := `
		SELECT m.id, m.sender_id, m.receiver_id, m.content, m.is_read, m.created_at,
		       u.username, u.first_name, u.last_name
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE (m.sender_id = ? AND m.receiver_id = ?) OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := db.Query(query, userID1, userID2, userID2, userID1, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []PrivateMessage
	for rows.Next() {
		var message PrivateMessage
		var sender User
		err := rows.Scan(
			&message.ID,
			&message.SenderID,
			&message.ReceiverID,
			&message.Content,
			&message.IsRead,
			&message.CreatedAt,
			&sender.Username,
			&sender.FirstName,
			&sender.LastName,
		)
		if err != nil {
			// log.Printf("GetConversationHistory scan error: %v", err)
			return nil, err
		}

		sender.ID = message.SenderID
		message.Sender = &sender

		messages = append(messages, message)
	}

	// Reverse the slice to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetUserConversations retrieves all conversations for a user
func GetUserConversations(db *sql.DB, userID int64) ([]Conversation, error) {
	query := `
		SELECT DISTINCT
			CASE 
				WHEN m.sender_id = ? THEN m.receiver_id 
				ELSE m.sender_id 
			END as other_user_id,
			u.username, u.first_name, u.last_name
		FROM messages m
		JOIN users u ON (
			CASE 
				WHEN m.sender_id = ? THEN m.receiver_id = u.id
				ELSE m.sender_id = u.id
			END
		)
		WHERE m.sender_id = ? OR m.receiver_id = ?`

	rows, err := db.Query(query, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(
			&conv.UserID,
			&conv.Username,
			&conv.FirstName,
			&conv.LastName,
		)
		if err != nil {
			return nil, err
		}

		// Get last message for this conversation
		lastMessage, err := getLastMessage(db, userID, conv.UserID)
		if err == nil {
			conv.LastMessage = lastMessage
		}

		// Get unread count
		conv.UnreadCount, _ = getUnreadCount(db, userID, conv.UserID)

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// getLastMessage retrieves the last message between two users
func getLastMessage(db *sql.DB, userID1, userID2 int64) (*PrivateMessage, error) {
	query := `
        SELECT m.id, m.sender_id, m.receiver_id, m.content, m.is_read, m.created_at,
               u.username, u.first_name, u.last_name
        FROM messages m
        JOIN users u ON m.sender_id = u.id
        WHERE (m.sender_id = ? AND m.receiver_id = ?) OR (m.sender_id = ? AND m.receiver_id = ?)
        ORDER BY m.created_at DESC
        LIMIT 1`

	var message PrivateMessage
	var sender User
	err := db.QueryRow(query, userID1, userID2, userID2, userID1).Scan(
		&message.ID,
		&message.SenderID,
		&message.ReceiverID,
		&message.Content,
		&message.IsRead,
		&message.CreatedAt,
		&sender.Username,
		&sender.FirstName,
		&sender.LastName,
	)
	if err != nil {
		// Add debug log here
		// log.Printf("getLastMessage error: %v", err)
		return nil, err
	}

	sender.ID = message.SenderID
	message.Sender = &sender

	return &message, nil
}

// getUnreadCount gets the number of unread messages from a specific user
func getUnreadCount(db *sql.DB, receiverID, senderID int64) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM messages 
		WHERE receiver_id = ? AND sender_id = ? AND is_read = FALSE`,
		receiverID, senderID).Scan(&count)
	return count, err
}

// MarkMessagesAsRead marks all messages from a sender to receiver as read
func MarkMessagesAsRead(db *sql.DB, receiverID, senderID int64) error {
	_, err := db.Exec(`
		UPDATE messages 
		SET is_read = TRUE 
		WHERE receiver_id = ? AND sender_id = ? AND is_read = FALSE`,
		receiverID, senderID)
	return err
}

// GetAllUsers retrieves all users except the current user (for chat user list)
func GetAllUsers(db *sql.DB, currentUserID int64) ([]User, error) {
	query := `
		SELECT id, username, first_name, last_name, email, age, gender, created_at
		FROM users 
		WHERE id != ?
		ORDER BY username`

	rows, err := db.Query(query, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Age,
			&user.Gender,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
