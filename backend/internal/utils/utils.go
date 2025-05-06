package utils

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var Validate = validator.New()

// DecodeJSONBody decodes a JSON request body into a struct
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

// RespondWithError sends an error response
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithValidationError handles validation errors
func RespondWithValidationError(w http.ResponseWriter, err error) {
	if _, ok := err.(validator.ValidationErrors); ok {
		RespondWithError(w, http.StatusBadRequest, "Validation failed")
		return
	}
	RespondWithError(w, http.StatusBadRequest, err.Error())
}

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// ComparePasswords compares a hashed password with its plaintext version
func ComparePasswords(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// SetAuthCookie sets the authentication cookie
func SetAuthCookie(w http.ResponseWriter, token string, expiresAt time.Time, isProduction bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearAuthCookie removes the authentication cookie
func ClearAuthCookie(w http.ResponseWriter, isProduction bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
	})
}

// GetAuthToken extracts the auth token from request
func GetAuthToken(r *http.Request) string {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		auth := r.Header.Get("Authorization")
		if len(auth) > 7 && auth[:7] == "Bearer " {
			return auth[7:]
		}
		return ""
	}
	return cookie.Value
}
