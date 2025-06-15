package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// SessionDuration is the default duration for a session
const SessionDuration = 24 * time.Hour

// GenerateToken generates a random session token
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
