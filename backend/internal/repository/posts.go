package repository

import (
	"database/sql"
	"fmt"
	"time"

	"real-time/backend/internal/model"
)

type PostRepository interface {
	Create(post *model.Post, categoryIDs []int64) error
	GetByID(id int64) (*model.Post, error)
	List(offset, limit int) ([]*model.Post, int, error)
	AddComment(comment *model.Comment) error
	GetComments(postID int64) ([]*model.Comment, error)
	GetCategories(postID int64) ([]*model.Category, error)
	GetByUserID(userID int64, offset, limit int) ([]*model.Post, int, error) // Updated signature
}

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(post *model.Post, categoryIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO posts (user_id, title, content, created_at, updated_at) 
         VALUES (?, ?, ?, ?, ?)`,
		post.UserID, post.Title, post.Content, time.Now(), time.Now(),
	)
	if err != nil {
		return err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	for _, catID := range categoryIDs {
		_, err = tx.Exec(
			"INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
			postID, catID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postRepository) GetByID(id int64) (*model.Post, error) {
	post := &model.Post{}
	err := r.db.QueryRow(
		`SELECT id, user_id, title, content, created_at, updated_at 
         FROM posts WHERE id = ?`, id,
	).Scan(&post.ID, &post.UserID, &post.Title,
		&post.Content, &post.CreatedAt, &post.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return post, err
}

func (r *postRepository) GetByUserID(userID int64, offset, limit int) ([]*model.Post, int, error) {
	// Count total posts for the user
	var total int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM posts WHERE user_id = ?",
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting posts: %v", err)
	}

	// Query posts with pagination
	query := `
        SELECT id, user_id, title, content, created_at, updated_at
        FROM posts
        WHERE user_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying posts: %v", err)
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
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("error scanning post: %v", err)
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating posts: %v", err)
	}

	return posts, total, nil
}

func (r *postRepository) List(offset, limit int) ([]*model.Post, int, error) {
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		`SELECT id, user_id, title, content, created_at, updated_at 
         FROM posts ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title,
			&post.Content, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, post)
	}

	return posts, total, nil
}

func (r *postRepository) AddComment(comment *model.Comment) error {
	_, err := r.db.Exec(
		`INSERT INTO comments (post_id, user_id, content, created_at) 
		 VALUES (?, ?, ?, ?)`,
		comment.PostID, comment.UserID, comment.Content, time.Now(),
	)
	return err
}

func (r *postRepository) GetComments(postID int64) ([]*model.Comment, error) {
	rows, err := r.db.Query(
		`SELECT id, post_id, user_id, content, created_at 
         FROM comments 
         WHERE post_id = ? 
         ORDER BY created_at DESC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		comment := &model.Comment{}
		if err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID,
			&comment.Content, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (r *postRepository) GetCategories(postID int64) ([]*model.Category, error) {
	rows, err := r.db.Query(
		`SELECT c.id, c.name 
         FROM categories c 
         JOIN post_categories pc ON c.id = pc.category_id 
         WHERE pc.post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		cat := &model.Category{}
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}
