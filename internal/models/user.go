package models

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a forum user
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never send in JSON responses
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Age          int       `json:"age"`
	Gender       string    `json:"gender"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
}

type LoginRequest struct {
	Login    string `json:"login"` // can be email or username
	Password string `json:"password"`
}

var (
	ErrUserExists         = errors.New("username or email already exists")
	ErrInvalidGender      = errors.New("gender must be 'male', 'female', or 'other'")
	ErrInvalidAge         = errors.New("age must be between 0 and 150")
	ErrInvalidPassword    = errors.New("password must be at least 6 characters")
	ErrInvalidCredentials = errors.New("invalid login credentials")
)

// CreateUser creates a new user in the database
func CreateUser(db *sql.DB, req RegisterRequest) (*User, error) {
	// Validate request
	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Check if user exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)",
		req.Username, req.Email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Insert user
	result, err := db.Exec(`
		INSERT INTO users (username, email, password_hash, first_name, last_name, age, gender)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.Username, req.Email, hashedPassword, req.FirstName, req.LastName, req.Age, req.Gender)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Return created user
	user := &User{
		ID:        id,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Age:       req.Age,
		Gender:    req.Gender,
	}

	return user, nil
}

func validateRegisterRequest(req RegisterRequest) error {
	if req.Age < 0 || req.Age > 150 {
		return ErrInvalidAge
	}

	if len(req.Password) < 6 {
		return ErrInvalidPassword
	}

	gender := req.Gender
	if gender != "male" && gender != "female" && gender != "other" {
		return ErrInvalidGender
	}

	return nil
}

// GetUserByLogin retrieves a user by email or username
func GetUserByLogin(db *sql.DB, login string) (*User, error) {
	query := `SELECT id, username, email, password_hash, first_name, last_name, age, gender, created_at 
			  FROM users WHERE email = ? OR username = ?`

	var user User
	err := db.QueryRow(query, login, login).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Age,
		&user.Gender,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidatePassword checks if the provided password is correct
func (u *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// CreateSession creates a new session for the user
func CreateSession(db *sql.DB, userID int64, token string) error {
	expiresAt := time.Now().Add(time.Hour * 24)
	_, err := db.Exec(`
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES (?, ?, ?)`,
		userID, token, expiresAt)
	return err
}

// GetUserBySessionToken retrieves a user by their session token
func GetUserBySessionToken(db *sql.DB, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.first_name, u.last_name, u.age, u.gender, u.created_at
		FROM users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.token = ? AND s.expires_at > ?`

	var user User
	err := db.QueryRow(query, token, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Age,
		&user.Gender,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("invalid or expired session")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
