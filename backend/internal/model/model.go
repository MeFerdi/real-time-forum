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
	Posts            []Post           `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Comments         []Comment        `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	SentMessages     []PrivateMessage `json:"-" gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	ReceivedMessages []PrivateMessage `json:"-" gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE"`
	Sessions         []Session        `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Following        []*User          `json:"-" gorm:"many2many:user_followers;joinForeignKey:FollowerID;joinReferences:FollowingID;constraint:OnDelete:CASCADE"`
	Followers        []*User          `json:"-" gorm:"many2many:user_followers;joinForeignKey:FollowingID;joinReferences:FollowerID;constraint:OnDelete:CASCADE"`
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

// Session represents a user's active session
type Session struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"unique;not null;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Post represents a forum post
type Post struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	UserID     int       `json:"userId" gorm:"not null"`
	Title      string    `json:"title"`
	Content    string    `json:"content" gorm:"not null"`
	ImageURL   string    `json:"imageUrl,omitempty"`
	Categories []string  `json:"categories" gorm:"-"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// Relationships
	User     User      `json:"-" gorm:"foreignKey:UserID"`
	Comments []Comment `json:"-" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
}

// ToDTO converts Post to PostDTO
func (p *Post) ToDTO(user UserDTO) PostDTO {
	return PostDTO{
		ID:         p.ID,
		UserID:     p.UserID,
		User:       user,
		Title:      p.Title,
		Content:    p.Content,
		ImageURL:   p.ImageURL,
		Categories: p.Categories,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

// Comment represents a comment on a post
type Comment struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	PostID    int       `json:"postId" gorm:"not null"`
	UserID    int       `json:"userId" gorm:"not null"`
	Content   string    `json:"content" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID"`
	Post Post `json:"-" gorm:"foreignKey:PostID"`
}

// CommentDTO represents the data transfer object for a comment
// type CommentDTO struct {
// 	ID        int       `json:"id"`
// 	PostID    int       `json:"postId"`
// 	UserID    int       `json:"userId"`
// 	User      UserDTO   `json:"user"`
// 	Content   string    `json:"content"`
// 	CreatedAt time.Time `json:"createdAt"`
// }

// ToDTO converts Comment to CommentDTO
func (c *Comment) ToDTO(user UserDTO) CommentDTO {
	return CommentDTO{
		ID:        c.ID,
		PostID:    c.PostID,
		UserID:    c.UserID,
		User:      user,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
	}
}

// PrivateMessage represents a direct message between users
type PrivateMessage struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	SenderID   int64     `json:"sender_id" gorm:"not null;index"`
	ReceiverID int64     `json:"receiver_id" gorm:"not null;index"`
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

// ToDTO converts Category to CategoryDTO
func (c *Category) ToDTO() CategoryDTO {
	return CategoryDTO{
		ID:   c.ID,
		Name: c.Name,
	}
}

// PostCategory represents the many-to-many relationship between posts and categories
type PostCategory struct {
	PostID     int64 `json:"post_id" gorm:"primaryKey"`
	CategoryID int64 `json:"category_id" gorm:"primaryKey"`
}

// DTOs for API responses

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

// PostDTO is the data transfer object for Post
type PostDTO struct {
	ID         int          `json:"id"`
	UserID     int          `json:"userId"`
	User       UserDTO      `json:"user"`
	Title      string       `json:"title"`
	Content    string       `json:"content"`
	ImageURL   string       `json:"imageUrl,omitempty"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	Comments   []CommentDTO `json:"comments,omitempty"`
	Categories []string     `json:"categories"`
}

// CommentDTO is the data transfer object for Comment
type CommentDTO struct {
	ID        int       `json:"id"`
	PostID    int       `json:"postId"`
	UserID    int       `json:"userId"`
	User      UserDTO   `json:"user"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// CategoryDTO is the data transfer object for Category
type CategoryDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// PostListDTO is the data transfer object for a list of posts
type PostListDTO struct {
	Posts      []PostDTO `json:"posts"`
	TotalPosts int64     `json:"totalPosts"`
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
}

// CreatePostRequest is the request body for creating a post
type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
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
