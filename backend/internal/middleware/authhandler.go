package middleware

import (
	"context"
	"log"
	"net/http"
	"real-time/backend/internal/repository"
	"strings"
)

type contextKey string

const userContextKey contextKey = "userID"

func SessionAuthMiddleware(sessionRepo repository.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			log.Printf("Request %s %s, Authorization: %s", r.Method, r.URL.Path, authHeader)
			if authHeader == "" {
				if cookie, err := r.Cookie("auth_token"); err == nil {
					log.Printf("Found auth_token cookie: %s", cookie.Value)
					authHeader = "Bearer " + cookie.Value
				} else {
					log.Println("Authorization header and cookie missing")
					http.Error(w, "Authorization header or cookie missing", http.StatusUnauthorized)
					return
				}
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				log.Println("Invalid token format")
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}
			log.Printf("Token: %s", token)

			userID, err := sessionRepo.Get(token)
			if err != nil {
				log.Printf("Session lookup error for token %s: %v", token, err)
				http.Error(w, "Invalid session", http.StatusUnauthorized)
				return
			}
			if userID == 0 {
				log.Printf("No valid session for token %s", token)
				http.Error(w, "Invalid session", http.StatusUnauthorized)
				return
			}
			log.Printf("Authenticated userID: %d", userID)

			ctx := context.WithValue(r.Context(), userContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userContextKey).(int)
	return userID, ok
}
