package models

import (
	"database/sql"
	"errors"
	"time"
)

type Post struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	Categories []Category `json:"categories"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Author     *User      `json:"author,omitempty"`
}

type Category struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreatePostRequest struct {
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	CategoryIDs []int64 `json:"category_ids"`
}

var (
	ErrEmptyTitle      = errors.New("title cannot be empty")
	ErrEmptyContent    = errors.New("content cannot be empty")
	ErrNoCategories    = errors.New("at least one category must be selected")
	ErrInvalidCategory = errors.New("one or more categories are invalid")
)

// CreatePost creates a new post and links it with the specified categories
func CreatePost(db *sql.DB, userID int64, req CreatePostRequest) (*Post, error) {
	if err := validateCreatePostRequest(req); err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert post
	result, err := tx.Exec(`
		INSERT INTO posts (user_id, title, content)
		VALUES (?, ?, ?)`,
		userID, req.Title, req.Content)
	if err != nil {
		return nil, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Link categories
	for _, categoryID := range req.CategoryIDs {
		_, err := tx.Exec(`
			INSERT INTO post_categories (post_id, category_id)
			VALUES (?, ?)`,
			postID, categoryID)
		if err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return the created post with categories
	return GetPostByID(db, postID)
}

// GetPostByID retrieves a post by its ID, including categories and author
func GetPostByID(db *sql.DB, id int64) (*Post, error) {
	// Get post
	post := &Post{}
	err := db.QueryRow(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		WHERE p.id = ?`, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Get categories
	rows, err := db.Query(`
		SELECT c.id, c.name, c.description, c.created_at
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cat Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)
		if err != nil {
			return nil, err
		}
		post.Categories = append(post.Categories, cat)
	}

	// Get author
	author, err := GetUserByID(db, post.UserID)
	if err != nil {
		return nil, err
	}
	post.Author = author

	return post, nil
}

// ListCategories returns all available categories
func ListCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query(`
		SELECT id, name, description, created_at 
		FROM categories 
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func validateCreatePostRequest(req CreatePostRequest) error {
	if req.Title == "" {
		return ErrEmptyTitle
	}
	if req.Content == "" {
		return ErrEmptyContent
	}
	if len(req.CategoryIDs) == 0 {
		return ErrNoCategories
	}
	return nil
}
