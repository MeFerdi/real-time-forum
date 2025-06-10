package model

import (
	"time"
)

// Post represents a forum post
type Post struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	UserID       int       `json:"userId" gorm:"not null"`
	Title        string    `json:"title"`
	Content      string    `json:"content" gorm:"not null"`
	ImageURL     string    `json:"imageUrl,omitempty"`
	Categories   []string  `json:"categories" gorm:"-"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	LikeCount    int       `json:"likeCount" gorm:"-"`
	DislikeCount int       `json:"dislikeCount" gorm:"-"`
	UserReaction string    `json:"userReaction,omitempty" gorm:"-"`

	// Relationships
	User     User      `gorm:"foreignKey:UserID"`
	Comments []Comment `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
}

// PostDTO is the data transfer object for Post
type PostDTO struct {
	ID           int          `json:"id"`
	UserID       int          `json:"userId"`
	User         UserDTO      `json:"user"`
	Title        string       `json:"title"`
	Content      string       `json:"content"`
	ImageURL     string       `json:"imageUrl,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	Comments     []CommentDTO `json:"comments,omitempty"`
	Categories   []string     `json:"categories"`
	LikeCount    int          `json:"likeCount"`
	DislikeCount int          `json:"dislikeCount"`
	UserReaction string       `json:"userReaction,omitempty"`
}

// ToDTO converts Post to PostDTO
func (p *Post) ToDTO(user UserDTO) PostDTO {
	return PostDTO{
		ID:           p.ID,
		UserID:       p.UserID,
		User:         user,
		Title:        p.Title,
		Content:      p.Content,
		ImageURL:     p.ImageURL,
		Categories:   p.Categories,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		Comments:     make([]CommentDTO, 0),
		LikeCount:    p.LikeCount,
		DislikeCount: p.DislikeCount,
		UserReaction: p.UserReaction,
	}
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
