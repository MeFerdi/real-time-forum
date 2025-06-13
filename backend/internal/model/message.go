package model

import (
	"time"
)

// PrivateMessage represents a direct message between users
type PrivateMessage struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	SenderID   int64     `json:"sender_id" gorm:"not null;index"`
	ReceiverID int64     `json:"receiver_id" gorm:"not null;index"`
	Content    string    `json:"content" gorm:"not null;type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	IsRead     bool      `json:"is_read" gorm:"default:false"`

	// Relationships
	Sender   *User `gorm:"foreignKey:SenderID"`
	Receiver *User `gorm:"foreignKey:ReceiverID"`
}
