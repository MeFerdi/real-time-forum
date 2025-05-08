package repository

import (
	"database/sql"
	"errors"
	domain "real-time/backend/internal/model"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("user already exists")
)

type UserRepository interface {
	EmailOrNicknameExists(email, nickname string) (bool, error)
	FindByIdentifier(identifier string) (*domain.User, error)
	Create(user domain.User) (*domain.User, error)
	UpdateLastOnline(userID int) error
	SetOnlineStatus(userID int, online bool) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) EmailOrNicknameExists(email, nickname string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE email = ? OR nickname = ?`,
		email, nickname,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) FindByIdentifier(identifier string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(
		`SELECT id, uuid, nickname, email, password_hash, first_name, last_name, 
		age, gender, created_at, last_online, is_online
		FROM users WHERE email = ? OR nickname = ?`,
		identifier, identifier,
	).Scan(
		&user.ID, &user.UUID, &user.Nickname, &user.Email,
		&user.PasswordHash, &user.FirstName, &user.LastName,
		&user.Age, &user.Gender, &user.CreatedAt,
		&user.LastOnline, &user.IsOnline,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) Create(user domain.User) (*domain.User, error) {
	exists, err := r.EmailOrNicknameExists(user.Email, user.Nickname)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	res, err := r.db.Exec(
		`INSERT INTO users 
		(uuid, nickname, email, password_hash, first_name, last_name, age, gender, is_online, last_online) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.UUID, user.Nickname, user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.Age, user.Gender,
		user.IsOnline, user.LastOnline,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	createdUser := user
	createdUser.ID = id
	return &createdUser, nil
}

func (r *userRepository) UpdateLastOnline(userID int) error {
	_, err := r.db.Exec(
		`UPDATE users SET last_online = ? WHERE id = ?`,
		time.Now(), userID,
	)
	return err
}

func (r *userRepository) SetOnlineStatus(userID int, online bool) error {
	_, err := r.db.Exec(
		`UPDATE users SET is_online = ?, last_online = ? WHERE id = ?`,
		online, time.Now(), userID,
	)
	return err
}
