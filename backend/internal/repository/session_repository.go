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

// Create inserts a new session into the database.
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

// Get retrieves the user ID for a valid session token, or 0 if not found/expired.
func (r *sessionRepository) Get(token string) (int, error) {
	var userID int
	var expiresAt time.Time
	err := r.db.QueryRow(
		`SELECT user_id, expires_at FROM sessions WHERE token = ?`,
		token,
	).Scan(&userID, &expiresAt)
	if err == sql.ErrNoRows {
		log.Printf("No session found for token %s", token)
		return 0, nil
	}
	if err != nil {
		log.Printf("Error fetching session for token %s: %v", token, err)
		return 0, err
	}
	if expiresAt.Before(time.Now()) {
		log.Printf("Session expired for token %s, expires_at: %v", token, expiresAt)
		return 0, nil
	}
	log.Printf("Session found for token %s, userID: %d, expires_at: %v", token, userID, expiresAt)
	return userID, nil
}

// Delete removes a session by token.
func (r *sessionRepository) Delete(token string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	if err != nil {
		log.Printf("Error deleting session for token %s: %v", token, err)
		return err
	}
	return nil
}
