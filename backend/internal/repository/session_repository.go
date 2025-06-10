package repository

import (
	"database/sql"
	"log"
	"real-time/backend/internal/model"
	"time"
)

// SessionRepository handles session-related database operations
type SessionRepository interface {
	Create(session *model.Session) error
	Get(token string) (int, error)
	Delete(token string) error
}

type sessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(session *model.Session) error {
	_, err := r.db.Exec(
		`INSERT INTO sessions (user_id, token, expires_at, created_at)
		 VALUES (?, ?, ?, ?)`,
		session.UserID, session.Token, session.ExpiresAt, session.CreatedAt,
	)
	if err != nil {
		log.Printf("Error creating session for user %d: %v", session.UserID, err)
		return err
	}
	return nil
}

func (r *sessionRepository) Get(token string) (int, error) {
	var userID int
	err := r.db.QueryRow(
		`SELECT user_id FROM sessions WHERE token = ? AND expires_at > ?`,
		token, time.Now(),
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		log.Printf("Error fetching session for token %s: %v", token, err)
		return 0, err
	}
	return userID, nil
}

func (r *sessionRepository) Delete(token string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	if err != nil {
		log.Printf("Error deleting session for token %s: %v", token, err)
		return err
	}
	return nil
}
