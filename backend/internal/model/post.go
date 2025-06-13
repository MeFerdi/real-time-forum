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

// PostList is a paginated list of posts
type PostList struct {
	Posts      []Post `json:"posts"`
	TotalPosts int64  `json:"totalPosts"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}

// CreatePostRequest is the request body for creating a post
type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
