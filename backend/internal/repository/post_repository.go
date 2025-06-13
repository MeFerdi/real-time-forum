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
	GetLikedPosts(userID int, offset, limit int) ([]*model.Post, int, error)
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
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	result, err := tx.Exec(
		`INSERT INTO posts (user_id, title, content, image_url, created_at, updated_at) 
         VALUES (?, ?, ?, ?, ?, ?)`,
		post.UserID, post.Title, post.Content, post.ImageURL, time.Now(), time.Now(),
	)
	if err != nil {
		tx.Rollback()
		log.Printf("Error inserting post: %v", err)
		return err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
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
		// Attach categories for each post
		categories, err := r.categoryRepo.GetCategories(post.ID)
		if err == nil {
			post.Categories = categories
		}
		posts = append(posts, post)
	}

	return posts, total, nil
}
func (r *postRepository) GetLikedPosts(userID int, offset, limit int) ([]*model.Post, int, error) {
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.image_url, p.created_at, COUNT(*) as total
		FROM posts p
		JOIN reactions r ON p.id = r.post_id
		WHERE r.user_id = ? AND r.type = 'like'
		GROUP BY p.id
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []*model.Post
	var total int
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.ImageURL, &p.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		posts = append(posts, &p)
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
		// Attach categories for each post
		categories, err := r.categoryRepo.GetCategories(post.ID)
		if err == nil {
			post.Categories = categories
		}
		posts = append(posts, post)
	}

	return posts, total, nil
}
