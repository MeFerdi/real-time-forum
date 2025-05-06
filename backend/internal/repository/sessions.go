package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(userID int64, timeout time.Duration) (string, time.Time, error)
	Get(token string) (int64, error)
	Delete(token string) error
	DeleteExpired() error
}

type sessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(userID int64, timeout time.Duration) (string, time.Time, error) {
	token := uuid.New().String()
	expiresAt := time.Now().Add(timeout)

	_, err := r.db.Exec(
		`INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)`,
		userID, token, expiresAt,
	)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (r *sessionRepository) Get(token string) (int64, error) {
	var userID int64
	err := r.db.QueryRow(
		`SELECT user_id FROM sessions 
        WHERE token = ? AND expires_at > datetime('now')`,
		token,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	return userID, err
}

func (r *sessionRepository) Delete(token string) error {
	result, err := r.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sessionRepository) DeleteExpired() error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at <= datetime('now')`)
	return err
}
