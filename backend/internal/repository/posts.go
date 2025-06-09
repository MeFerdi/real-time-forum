package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"real-time/backend/internal/model"
)

type PostRepository interface {
	Create(post *model.Post, categoryIDs []int64) error
	GetByID(id int) (*model.Post, error)
	GetByUserID(userID int, offset, limit int) ([]*model.Post, int, error)
	List(offset, limit int) ([]*model.Post, int, error)
	GetComments(postID int) ([]*model.Comment, error)
	GetCategories(postID int) ([]string, error)
	AddComment(comment *model.Comment) error
	GetCommentByID(commentID int) (*model.Comment, error)
	UpdateComment(comment *model.Comment) error
	DeleteComment(commentID int) error
	AddReaction(postID, userID int, reactionType string) error
	RemoveReaction(postID, userID int) error
	GetPostReactions(postID int) (likes int, dislikes int, error error)
	GetUserReaction(postID, userID int) (string, error)
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

	log.Printf("Post inserted with ID: %d", postID) // Debugging log

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

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	post.ID = int(postID)
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
		return nil, err
	}

	// Get categories
	categories, err := r.GetCategories(id)
	if err != nil {
		return nil, err
	}
	post.Categories = categories

	return post, nil
}

func (r *postRepository) GetByUserID(userID int, offset, limit int) ([]*model.Post, int, error) {
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
        SELECT id, user_id, title, content, image_url, created_at, updated_at
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
			&post.ImageURL,
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
	log.Printf("Listing posts with offset: %d, limit: %d", offset, limit)

	// Get total count
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&total)
	if err != nil {
		log.Printf("Error counting posts: %v", err)
		return nil, 0, err
	}
	log.Printf("Total posts in database: %d", total)

	// Get posts
	rows, err := r.db.Query(`
        SELECT p.id, p.user_id, p.title, p.content, p.image_url, p.created_at, p.updated_at
        FROM posts p
        ORDER BY p.created_at DESC
        LIMIT ? OFFSET ?
    `, limit, offset)
	if err != nil {
		log.Printf("Error querying posts: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImageURL,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning post row: %v", err)
			return nil, 0, err
		}
		posts = append(posts, post)
	}

	log.Printf("Retrieved %d posts", len(posts))
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

func (r *postRepository) GetComments(postID int) ([]*model.Comment, error) {
	rows, err := r.db.Query(`
        SELECT id, post_id, user_id, content, created_at 
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
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (r *postRepository) GetCategories(postID int) ([]string, error) {
	rows, err := r.db.Query(`
        SELECT c.name 
        FROM categories c 
        JOIN post_categories pc ON c.id = pc.category_id 
        WHERE pc.post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *postRepository) GetCommentByID(commentID int) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.QueryRow(`
		SELECT id, post_id, user_id, content, created_at 
		FROM comments 
		WHERE id = ?`,
		commentID,
	).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *postRepository) UpdateComment(comment *model.Comment) error {
	_, err := r.db.Exec(
		"UPDATE comments SET content = ? WHERE id = ?",
		comment.Content, comment.ID,
	)
	return err
}

func (r *postRepository) DeleteComment(commentID int) error {
	_, err := r.db.Exec("DELETE FROM comments WHERE id = ?", commentID)
	return err
}

func (r *postRepository) AddReaction(postID, userID int, reactionType string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete any existing reaction first
	_, err = tx.Exec("DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID)
	if err != nil {
		return err
	}

	// Add new reaction
	_, err = tx.Exec(
		"INSERT INTO post_reactions (post_id, user_id, reaction_type, created_at) VALUES (?, ?, ?, ?)",
		postID, userID, reactionType, time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postRepository) RemoveReaction(postID, userID int) error {
	_, err := r.db.Exec(
		"DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?",
		postID, userID,
	)
	return err
}

func (r *postRepository) GetPostReactions(postID int) (likes int, dislikes int, err error) {
	err = r.db.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN reaction_type = 'like' THEN 1 ELSE 0 END), 0) as likes,
			COALESCE(SUM(CASE WHEN reaction_type = 'dislike' THEN 1 ELSE 0 END), 0) as dislikes
		FROM post_reactions 
		WHERE post_id = ?`,
		postID,
	).Scan(&likes, &dislikes)
	return
}

func (r *postRepository) GetUserReaction(postID, userID int) (string, error) {
	var reactionType string
	err := r.db.QueryRow(
		"SELECT reaction_type FROM post_reactions WHERE post_id = ? AND user_id = ?",
		postID, userID,
	).Scan(&reactionType)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return reactionType, nil
}
