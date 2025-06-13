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
