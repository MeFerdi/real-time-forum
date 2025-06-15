package auth

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// SessionDuration is the default duration for a session
const (
	SessionDuration = 24 * time.Hour
	CookieName      = "session_token"
)

// CreateSession creates a new session for the user
func CreateSession(db *sql.DB, userID int64, w http.ResponseWriter) error {
	// Generate session token
	token := uuid.New().String()

	// Calculate expiration time
	expiresAt := time.Now().Add(SessionDuration)

	// Delete any existing sessions for this user
	_, err := db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	// Insert new session
	_, err = db.Exec(`
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES (?, ?, ?)`,
		userID, token, expiresAt)
	if err != nil {
		return err
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

// DeleteSession removes the session and clears the cookie
func DeleteSession(db *sql.DB, r *http.Request, w http.ResponseWriter) error {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return nil // No session to delete
	}

	// Delete session from database
	_, err = db.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)
	if err != nil {
		return err
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}
