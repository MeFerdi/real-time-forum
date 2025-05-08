package model

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be 8-72 characters")
	ErrInvalidNickname = errors.New("nickname must be 3-20 alphanumeric characters")
	emailRegex         = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
)

const (
	CategoryGeneral    = "General"
	CategoryTechnology = "Technology"
	CategorySports     = "Sports"
	CategoryMovies     = "Movies"
	CategoryMusic      = "Music"
	CategoryGaming     = "Gaming"
	CategoryTravel     = "Travel"
	CategoryFood       = "Food"
)

// User represents the application user
type User struct {
	ID           int64
	UUID         string
	Nickname     string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Age          int
	Gender       string
	CreatedAt    time.Time
	LastOnline   time.Time
	IsOnline     bool
}

// Post represents a user's post
type Post struct {
	ID        int64
	UserID    int64
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
type Category struct {
	ID   int64
	Name string
}

type CategoryDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (c *Category) ToDTO() CategoryDTO {
	return CategoryDTO{
		ID:   c.ID,
		Name: c.Name,
	}
}

// Comment represents a post comment
type Comment struct {
	ID        int64
	PostID    int64
	UserID    int64
	Content   string
	CreatedAt time.Time
}

// Session represents user authentication
type Session struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Request/Response DTOs
type UserDTO struct {
	ID         int64     `json:"id"`
	UUID       string    `json:"uuid"`
	Nickname   string    `json:"nickname"`
	Email      string    `json:"email"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Age        int       `json:"age"`
	Gender     string    `json:"gender"`
	CreatedAt  time.Time `json:"createdAt"`
	IsOnline   bool      `json:"isOnline"`
	LastOnline time.Time `json:"lastOnline"`
}

type RegisterRequest struct {
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type AuthResponse struct {
	User      UserDTO   `json:"user"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Helper methods for validation
func (r *RegisterRequest) Validate() error {
	if !emailRegex.MatchString(strings.ToLower(r.Email)) {
		return ErrInvalidEmail
	}
	if len(r.Password) < 8 || len(r.Password) > 72 {
		return ErrInvalidPassword
	}
	if len(r.Nickname) < 3 || len(r.Nickname) > 20 {
		return ErrInvalidNickname
	}
	return nil
}

// DTO conversion
func (u *User) ToDTO() UserDTO {
	return UserDTO{
		ID:         u.ID,
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

// Request DTOs
type CreatePostRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Category []string `json:"categories"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

// Response DTOs
type PostDTO struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    UserDTO   `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CommentDTO struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Author    UserDTO   `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}

type PostDetailDTO struct {
	PostDTO
	Comments []CommentDTO `json:"comments"`
}

type PostListDTO struct {
	Posts      []PostDTO `json:"posts"`
	TotalPosts int       `json:"totalPosts"`
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
}

// Conversion methods
func (p *Post) ToDTO(author UserDTO) PostDTO {
	return PostDTO{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		Author:    author,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func (c *Comment) ToDTO(author UserDTO) CommentDTO {
	return CommentDTO{
		ID:        c.ID,
		Content:   c.Content,
		Author:    author,
		CreatedAt: c.CreatedAt,
	}
}
