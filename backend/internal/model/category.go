package model

// Category represents a post category
type Category struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}

// CategoryDTO is the data transfer object for Category
type CategoryDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
