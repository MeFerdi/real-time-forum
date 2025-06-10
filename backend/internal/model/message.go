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

// MessageDTO is the data transfer object for PrivateMessage
type MessageDTO struct {
	ID         int64  `json:"id"`
	SenderID   int64  `json:"sender_id"`
	ReceiverID int64  `json:"receiver_id"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	IsRead     bool   `json:"is_read"`
}

// ToDTO converts PrivateMessage to MessageDTO
func (m *PrivateMessage) ToDTO() MessageDTO {
	return MessageDTO{
		ID:         int64(m.ID),
		SenderID:   m.SenderID,
		ReceiverID: m.ReceiverID,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		IsRead:     m.IsRead,
	}
}
