package middleware

import (
	"context"
	"net/http"
	"real-time/backend/internal/repository"
	"strings"
)

type contextKey string

const userContextKey contextKey = "userID"

// SessionAuthMiddleware is a middleware to handle session authentication
func SessionAuthMiddleware(sessionRepo repository.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			userID, err := sessionRepo.Get(token)
			if err != nil || userID == 0 {
				http.Error(w, "Invalid session", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext retrieves the userID from the context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userContextKey).(int)
	return userID, ok
}
