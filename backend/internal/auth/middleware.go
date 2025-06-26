package auth

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

type contextKey string

const UserIDContextKey contextKey = "userID"

// AuthMiddleware creates a new middleware that checks for valid session
func AuthMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session cookie
			cookie, err := r.Cookie("session_token")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if session exists and is valid
			var userID int64
			var expiresAt time.Time
			err = db.QueryRow(`
				SELECT user_id, expires_at 
				FROM sessions 
				WHERE token = ?`, cookie.Value).Scan(&userID, &expiresAt)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Check if session has expired
			if time.Now().After(expiresAt) {
				// Delete expired session
				_, _ = db.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)
				http.Error(w, "Session expired", http.StatusUnauthorized)
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth is a middleware that ensures a route is only accessible to authenticated users
func RequireAuth(next http.HandlerFunc, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		middleware := AuthMiddleware(db)
		handler := middleware(http.HandlerFunc(next))
		handler.ServeHTTP(w, r)
	}
}

// GetUserID retrieves the user ID from the request context
func GetUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	return userID, ok
}
