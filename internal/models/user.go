package models

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

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

var (
	ErrUserExists      = errors.New("username or email already exists")
	ErrInvalidGender   = errors.New("gender must be 'male', 'female', or 'other'")
	ErrInvalidAge      = errors.New("age must be between 0 and 150")
	ErrInvalidPassword = errors.New("password must be at least 6 characters")
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
