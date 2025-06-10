package repository

import (
	"database/sql"
	"log"
	"time"

	"real-time/backend/internal/model"
)

// CommentRepository handles comment-related database operations
type CommentRepository interface {
	AddComment(comment *model.Comment) error
	GetComments(postID int) ([]*model.Comment, error)
	GetCommentByID(commentID int) (*model.Comment, error)
	UpdateComment(comment *model.Comment) error
	DeleteComment(commentID int) error
}

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) AddComment(comment *model.Comment) error {
	result, err := r.db.Exec(
		`INSERT INTO comments (post_id, user_id, content, created_at) 
		 VALUES (?, ?, ?, ?)`,
		comment.PostID, comment.UserID, comment.Content, time.Now(),
	)
	if err != nil {
		log.Printf("Error adding comment for post %d: %v", comment.PostID, err)
		return err
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID for comment: %v", err)
		return err
	}
	comment.ID = int(commentID)
	return nil
}

func (r *commentRepository) GetComments(postID int) ([]*model.Comment, error) {
	rows, err := r.db.Query(`
        SELECT id, post_id, user_id, content, created_at 
        FROM comments 
        WHERE post_id = ? 
        ORDER BY created_at DESC`, postID)
	if err != nil {
		log.Printf("Error fetching comments for post %d: %v", postID, err)
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		comment := &model.Comment{}
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
		); err != nil {
			log.Printf("Error scanning comment: %v", err)
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (r *commentRepository) GetCommentByID(commentID int) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.QueryRow(`
		SELECT id, post_id, user_id, content, created_at 
		FROM comments 
		WHERE id = ?`, commentID).Scan(
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
		log.Printf("Error fetching comment %d: %v", commentID, err)
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) UpdateComment(comment *model.Comment) error {
	_, err := r.db.Exec(
		"UPDATE comments SET content = ? WHERE id = ?",
		comment.Content, comment.ID,
	)
	if err != nil {
		log.Printf("Error updating comment %d: %v", comment.ID, err)
		return err
	}
	return nil
}

func (r *commentRepository) DeleteComment(commentID int) error {
	_, err := r.db.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		log.Printf("Error deleting comment %d: %v", commentID, err)
		return err
	}
	return nil
}
