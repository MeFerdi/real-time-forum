package model

// Category represents a post category
type Category struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}

// PostCategory represents the many-to-many relationship between posts and categories
type PostCategory struct {
	PostID     int64 `json:"post_id" gorm:"primaryKey"`
	CategoryID int64 `json:"category_id" gorm:"primaryKey"`
}
