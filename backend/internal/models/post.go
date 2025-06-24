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
	Comments   []Comment  `json:"comments,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Author     *User      `json:"author,omitempty"`
	LikeCount  int        `json:"like_count"`
}

type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	ParentID  *int64    `json:"parent_id,omitempty"`
	Replies   []Comment `json:"replies,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Author    *User     `json:"author,omitempty"`
	LikeCount int       `json:"like_count"`
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

type CreateCommentRequest struct {
	Content  string `json:"content"`
	ParentID *int64 `json:"parent_id,omitempty"`
}

var (
	ErrEmptyTitle      = errors.New("title cannot be empty")
	ErrEmptyContent    = errors.New("content cannot be empty")
	ErrNoCategories    = errors.New("at least one category must be selected")
	ErrInvalidCategory = errors.New("one or more categories are invalid")
	ErrEmptyComment    = errors.New("comment content cannot be empty")
	ErrPostNotFound    = errors.New("post not found")
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
	// Get post with like count
	post := &Post{}
	err := db.QueryRow(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at,
		       (SELECT COUNT(*) FROM likes WHERE post_id = p.id) as like_count
		FROM posts p
		WHERE p.id = ?`, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.LikeCount,
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

	// Get comments (only parent comments)
	rows, err = db.Query(`
		SELECT c.id, c.post_id, c.user_id, c.content, c.parent_id, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM likes WHERE comment_id = c.id) as like_count
		FROM comments c
		WHERE c.post_id = ? AND c.parent_id IS NULL
		ORDER BY c.created_at ASC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.ParentID,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.LikeCount,
		)
		if err != nil {
			return nil, err
		}

		// Get comment author
		commentAuthor, err := GetUserByID(db, comment.UserID)
		if err != nil {
			return nil, err
		}
		comment.Author = commentAuthor

		// Get replies
		comment.Replies, err = getCommentReplies(db, comment.ID)
		if err != nil {
			return nil, err
		}

		post.Comments = append(post.Comments, comment)
	}

	// Get like count
	post.LikeCount, err = GetPostLikes(db, post.ID)
	if err != nil {
		return nil, err
	}

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

// ListPosts returns all posts with their categories and authors
func ListPosts(db *sql.DB) ([]Post, error) {
	// Get all posts
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		ORDER BY p.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
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

		// Get categories for this post
		catRows, err := db.Query(`
			SELECT c.id, c.name, c.description, c.created_at
			FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?`, post.ID)
		if err != nil {
			return nil, err
		}
		defer catRows.Close()

		for catRows.Next() {
			var cat Category
			err := catRows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)
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

		// Get like count
		post.LikeCount, err = GetPostLikes(db, post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// ListPostsByCategory returns all posts in a specific category
func ListPostsByCategory(db *sql.DB, categoryID int64) ([]Post, error) {
	// Get posts with a specific category
	rows, err := db.Query(`
		SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?
		ORDER BY p.created_at DESC`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
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

		// Get categories for this post
		catRows, err := db.Query(`
			SELECT c.id, c.name, c.description, c.created_at
			FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?`, post.ID)
		if err != nil {
			return nil, err
		}
		defer catRows.Close()

		for catRows.Next() {
			var cat Category
			err := catRows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)
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

		// Get like count
		post.LikeCount, err = GetPostLikes(db, post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// ListPostsByUserID returns all posts created by a specific user
func ListPostsByUserID(db *sql.DB, userID int64) ([]Post, error) {
	// Get posts by user
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at
		FROM posts p
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
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

		// Get categories for this post
		catRows, err := db.Query(`
			SELECT c.id, c.name, c.description, c.created_at
			FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?`, post.ID)
		if err != nil {
			return nil, err
		}
		defer catRows.Close()

		for catRows.Next() {
			var cat Category
			err := catRows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)
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

		// Get like count
		post.LikeCount, err = GetPostLikes(db, post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// CreateComment adds a new comment to a post
func CreateComment(db *sql.DB, postID, userID int64, req CreateCommentRequest) (*Comment, error) {
	if req.Content == "" {
		return nil, ErrEmptyComment
	}

	// Check if post exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", postID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPostNotFound
	}

	// If parent_id is provided, check if it exists and belongs to the same post
	if req.ParentID != nil {
		err := db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM comments 
				WHERE id = ? AND post_id = ?
			)`, req.ParentID, postID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("parent comment not found or doesn't belong to this post")
		}
	}

	result, err := db.Exec(`
		INSERT INTO comments (post_id, user_id, content, parent_id)
		VALUES (?, ?, ?, ?)`,
		postID, userID, req.Content, req.ParentID)
	if err != nil {
		return nil, err
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetCommentByID(db, commentID)
}

// GetCommentByID retrieves a comment by its ID
func GetCommentByID(db *sql.DB, id int64) (*Comment, error) {
	comment := &Comment{}
	err := db.QueryRow(`
		SELECT c.id, c.post_id, c.user_id, c.content, c.parent_id, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM likes WHERE comment_id = c.id) as like_count
		FROM comments c
		WHERE c.id = ?`, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.ParentID,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&comment.LikeCount,
	)
	if err != nil {
		return nil, err
	}

	// Get author
	author, err := GetUserByID(db, comment.UserID)
	if err != nil {
		return nil, err
	}
	comment.Author = author

	// Get replies if this is a parent comment
	rows, err := db.Query(`
		SELECT id, post_id, user_id, content, parent_id, created_at, updated_at
		FROM comments
		WHERE parent_id = ?
		ORDER BY created_at ASC`, comment.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reply Comment
		err := rows.Scan(
			&reply.ID,
			&reply.PostID,
			&reply.UserID,
			&reply.Content,
			&reply.ParentID,
			&reply.CreatedAt,
			&reply.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get reply author
		replyAuthor, err := GetUserByID(db, reply.UserID)
		if err != nil {
			return nil, err
		}
		reply.Author = replyAuthor

		comment.Replies = append(comment.Replies, reply)
	}

	// Get like count
	comment.LikeCount, err = GetCommentLikes(db, comment.ID)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// getCommentReplies returns all replies for a given comment
func getCommentReplies(db *sql.DB, parentID int64) ([]Comment, error) {
	rows, err := db.Query(`
		SELECT c.id, c.post_id, c.user_id, c.content, c.parent_id, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM likes WHERE comment_id = c.id) as like_count
		FROM comments c
		WHERE c.parent_id = ?
		ORDER BY c.created_at ASC`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []Comment
	for rows.Next() {
		var reply Comment
		err := rows.Scan(
			&reply.ID,
			&reply.PostID,
			&reply.UserID,
			&reply.Content,
			&reply.ParentID,
			&reply.CreatedAt,
			&reply.UpdatedAt,
			&reply.LikeCount,
		)
		if err != nil {
			return nil, err
		}

		// Get reply author
		replyAuthor, err := GetUserByID(db, reply.UserID)
		if err != nil {
			return nil, err
		}
		reply.Author = replyAuthor

		replies = append(replies, reply)
	}

	return replies, nil
}

// LikePost records a like for a post by a user
func LikePost(db *sql.DB, postID, userID int64) error {
	// Check if the post exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", postID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("post not found")
	}

	// Check if the user has already liked this post
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE post_id = ? AND user_id = ?)",
		postID, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		// If like exists, remove it (unlike)
		_, err := db.Exec("DELETE FROM likes WHERE post_id = ? AND user_id = ?", postID, userID)
		return err
	}

	// Add new like
	_, err = db.Exec("INSERT INTO likes (post_id, user_id) VALUES (?, ?)", postID, userID)
	return err
}

// LikeComment records a like for a comment by a user
func LikeComment(db *sql.DB, commentID, userID int64) error {
	// Check if the comment exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE id = ?)", commentID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("comment not found")
	}

	// Check if the user has already liked this comment
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE comment_id = ? AND user_id = ?)",
		commentID, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		// If like exists, remove it (unlike)
		_, err := db.Exec("DELETE FROM likes WHERE comment_id = ? AND user_id = ?", commentID, userID)
		return err
	}

	// Add new like
	_, err = db.Exec("INSERT INTO likes (comment_id, user_id) VALUES (?, ?)", commentID, userID)
	return err
}

// GetPostLikes returns the number of likes for a post
func GetPostLikes(db *sql.DB, postID int64) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ?", postID).Scan(&count)
	return count, err
}

// GetCommentLikes returns the number of likes for a comment
func GetCommentLikes(db *sql.DB, commentID int64) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE comment_id = ?", commentID).Scan(&count)
	return count, err
}

// HasUserLikedPost checks if a user has liked a post
func HasUserLikedPost(db *sql.DB, postID, userID int64) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE post_id = ? AND user_id = ?)",
		postID, userID).Scan(&exists)
	return exists, err
}

// HasUserLikedComment checks if a user has liked a comment
func HasUserLikedComment(db *sql.DB, commentID, userID int64) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE comment_id = ? AND user_id = ?)",
		commentID, userID).Scan(&exists)
	return exists, err
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
