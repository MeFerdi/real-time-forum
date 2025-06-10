package repository

import (
	"database/sql"
	"log"
	"time"

	"real-time/backend/internal/model"
)

// PostRepository handles post-related database operations
type PostRepository interface {
	Create(post *model.Post) error
	GetByID(id int) (*model.Post, error)
	GetByUserID(userID int, offset, limit int) ([]*model.Post, int, error)
	List(offset, limit int) ([]*model.Post, int, error)
}

type postRepository struct {
	db           *sql.DB
	categoryRepo CategoryRepository
}

func NewPostRepository(db *sql.DB, categoryRepo CategoryRepository) PostRepository {
	return &postRepository{db: db, categoryRepo: categoryRepo}
}

func (r *postRepository) Create(post *model.Post) error {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO posts (user_id, title, content, image_url, created_at, updated_at) 
         VALUES (?, ?, ?, ?, ?, ?)`,
		post.UserID, post.Title, post.Content, post.ImageURL, time.Now(), time.Now(),
	)
	if err != nil {
		log.Printf("Error inserting post: %v", err)
		return err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		return err
	}

	post.ID = int(postID)
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Successfully created post with ID: %d", postID)
	return nil
}

func (r *postRepository) GetByID(id int) (*model.Post, error) {
	post := &model.Post{}
	err := r.db.QueryRow(`
        SELECT id, user_id, title, content, image_url, created_at, updated_at 
        FROM posts WHERE id = ?`, id).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImageURL,
		&post.CreatedAt, &post.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("Error fetching post %d: %v", id, err)
		return nil, err
	}

	categories, err := r.categoryRepo.GetCategories(id)
	if err != nil {
		log.Printf("Error fetching categories for post %d: %v", id, err)
		return nil, err
	}
	post.Categories = categories

	return post, nil
}

func (r *postRepository) GetByUserID(userID int, offset, limit int) ([]*model.Post, int, error) {
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		log.Printf("Error counting posts for user %d: %v", userID, err)
		return nil, 0, err
	}

	rows, err := r.db.Query(`
        SELECT id, user_id, title, content, image_url, created_at, updated_at
        FROM posts
        WHERE user_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?`, userID, limit, offset)
	if err != nil {
		log.Printf("Error querying posts for user %d: %v", userID, err)
		return nil, 0, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImageURL,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			log.Printf("Error scanning post: %v", err)
			return nil, 0, err
		}
		posts = append(posts, post)
	}

	return posts, total, nil
}

func (r *postRepository) List(offset, limit int) ([]*model.Post, int, error) {
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&total)
	if err != nil {
		log.Printf("Error counting posts: %v", err)
		return nil, 0, err
	}

	rows, err := r.db.Query(`
        SELECT id, user_id, title, content, image_url, created_at, updated_at
        FROM posts
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		log.Printf("Error querying posts: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImageURL,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			log.Printf("Error scanning post: %v", err)
			return nil, 0, err
		}
		posts = append(posts, post)
	}

	return posts, total, nil
}
