package repository

import (
	"database/sql"
	"log"
	"time"
)

// ReactionRepository handles reaction-related database operations
type ReactionRepository interface {
	AddReaction(postID, userID int, reactionType string) error
	RemoveReaction(postID, userID int) error
	GetPostReactions(postID int) (likes int, dislikes int, err error)
	GetUserReaction(postID, userID int) (string, error)
}

type reactionRepository struct {
	db *sql.DB
}

func NewReactionRepository(db *sql.DB) ReactionRepository {
	return &reactionRepository{db: db}
}

// AddReaction adds or updates a user's reaction to a post.
func (r *reactionRepository) AddReaction(postID, userID int, reactionType string) error {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID)
	if err != nil {
		log.Printf("Error deleting existing reaction for post %d, user %d: %v", postID, userID, err)
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO post_reactions (post_id, user_id, reaction_type, created_at) VALUES (?, ?, ?, ?)",
		postID, userID, reactionType, time.Now(),
	)
	if err != nil {
		log.Printf("Error adding reaction for post %d, user %d: %v", postID, userID, err)
		return err
	}

	return tx.Commit()
}

// RemoveReaction removes a user's reaction from a post.
func (r *reactionRepository) RemoveReaction(postID, userID int) error {
	_, err := r.db.Exec(
		"DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?",
		postID, userID,
	)
	if err != nil {
		log.Printf("Error removing reaction for post %d, user %d: %v", postID, userID, err)
		return err
	}
	return nil
}

// GetPostReactions returns the number of likes and dislikes for a post.
func (r *reactionRepository) GetPostReactions(postID int) (likes int, dislikes int, err error) {
	err = r.db.QueryRow(`
        SELECT 
            COALESCE(SUM(CASE WHEN reaction_type = 'like' THEN 1 ELSE 0 END), 0) as likes,
            COALESCE(SUM(CASE WHEN reaction_type = 'dislike' THEN 1 ELSE 0 END), 0) as dislikes
        FROM post_reactions 
        WHERE post_id = ?`, postID).Scan(&likes, &dislikes)
	if err != nil {
		log.Printf("Error fetching reactions for post %d: %v", postID, err)
	}
	return
}

// GetUserReaction returns the reaction type for a user on a post.
func (r *reactionRepository) GetUserReaction(postID, userID int) (string, error) {
	var reactionType string
	err := r.db.QueryRow(
		"SELECT reaction_type FROM post_reactions WHERE post_id = ? AND user_id = ?",
		postID, userID,
	).Scan(&reactionType)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		log.Printf("Error fetching user reaction for post %d, user %d: %v", postID, userID, err)
		return "", err
	}
	return reactionType, nil
}
