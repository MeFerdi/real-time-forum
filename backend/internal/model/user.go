package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	UUID         string    `json:"uuid" gorm:"unique;not null"`
	Nickname     string    `json:"nickname" gorm:"unique;not null"`
	Email        string    `json:"email" gorm:"unique;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	FirstName    string    `json:"first_name" gorm:"not null"`
	LastName     string    `json:"last_name" gorm:"not null"`
	Age          int       `json:"age" gorm:"not null"`
	Gender       string    `json:"gender" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	LastOnline   time.Time `json:"last_online"`
	IsOnline     bool      `json:"is_online" gorm:"default:false"`

	// Relationships
	Posts            []Post           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Comments         []Comment        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	SentMessages     []PrivateMessage `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	ReceivedMessages []PrivateMessage `gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE"`
	Sessions         []Session        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Following        []*User          `gorm:"many2many:user_followers;joinForeignKey:FollowerID;joinReferences:FollowingID;constraint:OnDelete:CASCADE"`
	Followers        []*User          `gorm:"many2many:user_followers;joinForeignKey:FollowingID;joinReferences:FollowerID;constraint:OnDelete:CASCADE"`
}

// UserDTO is the data transfer object for User
type UserDTO struct {
	ID         int       `json:"id"`
	UUID       string    `json:"uuid"`
	Nickname   string    `json:"nickname"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Age        int       `json:"age"`
	Gender     string    `json:"gender"`
	CreatedAt  time.Time `json:"created_at"`
	LastOnline time.Time `json:"last_online"`
	IsOnline   bool      `json:"is_online"`
}

// ToDTO converts User to UserDTO
func (u *User) ToDTO() UserDTO {
	return UserDTO{
		ID:         u.ID,
		UUID:       u.UUID,
		Nickname:   u.Nickname,
		Email:      u.Email,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Age:        u.Age,
		Gender:     u.Gender,
		CreatedAt:  u.CreatedAt,
		LastOnline: u.LastOnline,
		IsOnline:   u.IsOnline,
	}
}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Identifier string `json:"identifier"` // Email or nickname
	Password   string `json:"password"`
}

// AuthResponse is the response for register and login
type AuthResponse struct {
	User      UserDTO   `json:"user"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
