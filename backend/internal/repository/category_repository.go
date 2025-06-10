package repository

import (
	"database/sql"
	"log"
)

// CategoryRepository handles category-related database operations
type CategoryRepository interface {
	AddPostCategories(postID int, categoryIDs []int64) error
	GetCategories(postID int) ([]string, error)
}

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) AddPostCategories(postID int, categoryIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	for _, categoryID := range categoryIDs {
		_, err := tx.Exec(
			`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`,
			postID, categoryID,
		)
		if err != nil {
			log.Printf("Error inserting category %d for post %d: %v", categoryID, postID, err)
			return err
		}
	}

	return tx.Commit()
}

func (r *categoryRepository) GetCategories(postID int) ([]string, error) {
	rows, err := r.db.Query(`
        SELECT c.name 
        FROM categories c 
        JOIN post_categories pc ON c.id = pc.category_id 
        WHERE pc.post_id = ?`, postID)
	if err != nil {
		log.Printf("Error fetching categories for post %d: %v", postID, err)
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			log.Printf("Error scanning category for post %d: %v", postID, err)
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}
