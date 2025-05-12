package utils

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	emailRegex         = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
)

func ValidateEmail(email string) error {
	if !emailRegex.MatchString(strings.ToLower(email)) {
		return ErrInvalidEmail
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func HashPassword(password string) (string, error) {
	h := sha256.New()
	h.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func ComparePasswords(hashedPassword, password string) bool {
	h := sha256.New()
	h.Write([]byte(password))
	hashedInput := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(hashedInput)) == 1
}

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
