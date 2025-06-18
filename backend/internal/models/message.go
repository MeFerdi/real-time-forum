package models

import (
	"database/sql"
	"time"
)

// Message represents a private message between users
type Message struct {
	ID         int64     `json:"id"`
	SenderID   int64     `json:"sender_id"`
	ReceiverID int64     `json:"receiver_id"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
	Sender     *User     `json:"sender,omitempty"`
	Receiver   *User     `json:"receiver,omitempty"`
}

// GetMessagesByUsers retrieves messages between two users with pagination
func GetMessagesByUsers(db *sql.DB, user1ID, user2ID int64, offset int) ([]Message, error) {
	rows, err := db.Query(`
        SELECT id, sender_id, receiver_id, content, is_read, created_at
        FROM messages 
        WHERE (sender_id = ? AND receiver_id = ?) 
           OR (sender_id = ? AND receiver_id = ?)
        ORDER BY created_at DESC
        LIMIT 10 OFFSET ?`,
		user1ID, user2ID, user2ID, user1ID, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.IsRead,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get sender details
		msg.Sender, err = GetUserByID(db, msg.SenderID)
		if err != nil {
			return nil, err
		}

		// Get receiver details
		msg.Receiver, err = GetUserByID(db, msg.ReceiverID)
		if err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// CreateMessage saves a new message to the database
func CreateMessage(db *sql.DB, senderID, receiverID int64, content string) (*Message, error) {
	result, err := db.Exec(`
        INSERT INTO messages (sender_id, receiver_id, content, is_read, created_at)
        VALUES (?, ?, ?, false, ?)`,
		senderID, receiverID, content, time.Now())
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetMessageByID(db, id)
}

// GetMessageByID retrieves a single message by its ID
func GetMessageByID(db *sql.DB, id int64) (*Message, error) {
	msg := &Message{}
	err := db.QueryRow(`
        SELECT id, sender_id, receiver_id, content, is_read, created_at
        FROM messages
        WHERE id = ?`, id).Scan(
		&msg.ID,
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.Content,
		&msg.IsRead,
		&msg.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Get sender details
	msg.Sender, err = GetUserByID(db, msg.SenderID)
	if err != nil {
		return nil, err
	}

	// Get receiver details
	msg.Receiver, err = GetUserByID(db, msg.ReceiverID)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// MarkMessageAsRead marks a message as read
func MarkMessageAsRead(db *sql.DB, messageID int64) error {
	_, err := db.Exec(`
        UPDATE messages 
        SET is_read = true 
        WHERE id = ?`, messageID)
	return err
}

// GetUnreadMessageCount returns count of unread messages for a user
func GetUnreadMessageCount(db *sql.DB, userID int64) (int, error) {
	var count int
	err := db.QueryRow(`
        SELECT COUNT(*) 
        FROM messages 
        WHERE receiver_id = ? AND is_read = false`,
		userID).Scan(&count)
	return count, err
}

// GetRecentChats returns users with whom the current user has exchanged messages
func GetRecentChats(db *sql.DB, userID int64) ([]User, error) {
	rows, err := db.Query(`
        SELECT DISTINCT 
            CASE 
                WHEN sender_id = ? THEN receiver_id
                ELSE sender_id
            END as other_user_id
        FROM messages
        WHERE sender_id = ? OR receiver_id = ?
        ORDER BY (
            SELECT created_at 
            FROM messages m2 
            WHERE (m2.sender_id = messages.sender_id AND m2.receiver_id = messages.receiver_id)
               OR (m2.sender_id = messages.receiver_id AND m2.sender_id = messages.sender_id)
            ORDER BY created_at DESC 
            LIMIT 1
        ) DESC`,
		userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var otherUserID int64
		if err := rows.Scan(&otherUserID); err != nil {
			return nil, err
		}

		user, err := GetUserByID(db, otherUserID)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	return users, nil
}
