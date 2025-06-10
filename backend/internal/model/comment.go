package model

import (
	"time"
)

// Comment represents a comment on a post
type Comment struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	PostID    int       `json:"postId" gorm:"not null"`
	UserID    int       `json:"userId" gorm:"not null"`
	Content   string    `json:"content" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`

	// Relationships
	User User `gorm:"foreignKey:UserID"`
	Post Post `gorm:"foreignKey:PostID"`
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
