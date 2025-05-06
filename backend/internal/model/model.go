package domain

import (
	"time"
)

type User struct {
	ID           int       `json:"id"`
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
	Posts            []Post           `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Comments         []Comment        `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	SentMessages     []PrivateMessage `json:"-" gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE;"`
	ReceivedMessages []PrivateMessage `json:"-" gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE;"`
	Sessions         []Session        `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Likes            []Like           `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Following        []*User          `json:"-" gorm:"many2many:user_followers;joinForeignKey:FollowerID;joinReferences:FollowingID;constraint:OnDelete:CASCADE;"`
	Followers        []*User          `json:"-" gorm:"many2many:user_followers;joinForeignKey:FollowingID;joinReferences:FollowerID;constraint:OnDelete:CASCADE;"`
}

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
	IsOnline   bool      `json:"is_online"`
	LastOnline time.Time `json:"last_online"`
}

type RegisterRequest struct {
	Nickname  string `json:"nickname" validate:"required,alphanum,min=3,max=20"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	FirstName string `json:"first_name" validate:"required,alpha,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,alpha,min=2,max=50"`
	Age       int    `json:"age" validate:"required,min=13,max=120"`
	Gender    string `json:"gender" validate:"required,oneof=male female other"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type AuthResponse struct {
	User      UserDTO   `json:"user"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

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
		IsOnline:   u.IsOnline,
		LastOnline: u.LastOnline,
	}
}

// Session represents a user's active session
type Session struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"unique;not null;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Post represents a user-created post
type Post struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"not null;index"`
	Title     string    `json:"title" gorm:"not null;size:255"`
	Content   string    `json:"content" gorm:"not null;type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User       *User      `json:"user,omitempty" gorm:"-"`
	Comments   []Comment  `json:"comments,omitempty" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE;"`
	Categories []Category `json:"categories,omitempty" gorm:"many2many:post_categories;constraint:OnDelete:CASCADE;"`
	Likes      []Like     `json:"likes,omitempty" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE;"`
}

// Comment represents a comment on a post
type Comment struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	PostID    int       `json:"post_id" gorm:"not null;index"`
	UserID    int       `json:"user_id" gorm:"not null;index"`
	Content   string    `json:"content" gorm:"not null;type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Post *Post `json:"post,omitempty" gorm:"-"`
	User *User `json:"user,omitempty" gorm:"-"`
}

// PrivateMessage represents a direct message between users
type PrivateMessage struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	SenderID   int       `json:"sender_id" gorm:"not null;index"`
	ReceiverID int       `json:"receiver_id" gorm:"not null;index"`
	Content    string    `json:"content" gorm:"not null;type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	IsRead     bool      `json:"is_read" gorm:"default:false"`

	// Relationships
	Sender   *User `json:"sender,omitempty" gorm:"-"`
	Receiver *User `json:"receiver,omitempty" gorm:"-"`
}

// Category represents a post category
type Category struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}

// Like represents a user liking a post
type Like struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"not null;index"`
	PostID    int       `json:"post_id" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"-"`
	Post *Post `json:"post,omitempty" gorm:"-"`
}

// UserFollower represents the many-to-many relationship between users
type UserFollower struct {
	FollowerID  int       `json:"follower_id" gorm:"primaryKey"`
	FollowingID int       `json:"following_id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// PostCategory represents the many-to-many relationship between posts and categories
type PostCategory struct {
	PostID     int `json:"post_id" gorm:"primaryKey"`
	CategoryID int `json:"category_id" gorm:"primaryKey"`
}
