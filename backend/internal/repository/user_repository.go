package repository

import (
	"database/sql"
	"log"
	"time"

	"real-time/backend/internal/model"
)

// UserRepository handles user-related database operations
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id int) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByNickname(nickname string) (*model.User, error)
	CreatePrivateMessage(message model.PrivateMessage) error
	GetPrivateMessages(senderID, receiverID int) ([]model.PrivateMessage, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (uuid, nickname, email, password_hash, first_name, last_name, age, gender, created_at, last_online, is_online)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.UUID, user.Nickname, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Age, user.Gender,
		time.Now(), time.Now(), user.IsOnline,
	)
	if err != nil {
		log.Printf("Error creating user %s: %v", user.Email, err)
		return err
	}
	return nil
}

func (r *userRepository) GetByID(id int) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(`
		SELECT id, uuid, nickname, email, password_hash, first_name, last_name, age, gender, created_at, last_online, is_online
		FROM users WHERE id = ?`, id).Scan(
		&user.ID, &user.UUID, &user.Nickname, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.Age, &user.Gender, &user.CreatedAt, &user.LastOnline, &user.IsOnline,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("Error fetching user %d: %v", id, err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(`
		SELECT id, uuid, nickname, email, password_hash, first_name, last_name, age, gender, created_at, last_online, is_online
		FROM users WHERE email = ?`, email).Scan(
		&user.ID, &user.UUID, &user.Nickname, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.Age, &user.Gender, &user.CreatedAt, &user.LastOnline, &user.IsOnline,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("Error fetching user by email %s: %v", email, err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByNickname(nickname string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(`
		SELECT id, uuid, nickname, email, password_hash, first_name, last_name, age, gender, created_at, last_online, is_online
		FROM users WHERE nickname = ?`, nickname).Scan(
		&user.ID, &user.UUID, &user.Nickname, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.Age, &user.Gender, &user.CreatedAt, &user.LastOnline, &user.IsOnline,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("Error fetching user by nickname %s: %v", nickname, err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreatePrivateMessage(message model.PrivateMessage) error {
	_, err := r.db.Exec(
		`INSERT INTO private_messages (sender_id, receiver_id, content, created_at, is_read)
		 VALUES (?, ?, ?, ?, ?)`,
		message.SenderID, message.ReceiverID, message.Content, time.Now(), message.IsRead,
	)
	if err != nil {
		log.Printf("Error creating private message from %d to %d: %v", message.SenderID, message.ReceiverID, err)
		return err
	}
	return nil
}

func (r *userRepository) GetPrivateMessages(senderID, receiverID int) ([]model.PrivateMessage, error) {
	rows, err := r.db.Query(`
		SELECT id, sender_id, receiver_id, content, created_at, is_read
		FROM private_messages
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at ASC`, senderID, receiverID, receiverID, senderID)
	if err != nil {
		log.Printf("Error fetching messages between %d and %d: %v", senderID, receiverID, err)
		return nil, err
	}
	defer rows.Close()

	var messages []model.PrivateMessage
	for rows.Next() {
		var msg model.PrivateMessage
		if err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.CreatedAt, &msg.IsRead,
		); err != nil {
			log.Printf("Error scanning message: %v", err)
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
