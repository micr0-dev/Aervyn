package database

import (
	"context"
	"database/sql"

	"github.com/micr0-dev/Aervyn/internal/models"
)

// User operations
func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (username, display_name, email, password_hash, summary)
        VALUES (?, ?, ?, ?, ?)
        RETURNING id, created_at, updated_at
    `

	return db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.DisplayName,
		user.Email,
		user.PasswordHash,
		user.Summary,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (db *DB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, username, display_name, email, password_hash, summary,
               private_key, public_key, created_at, updated_at
        FROM users
        WHERE username = ?
    `

	err := db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.Email,
		&user.PasswordHash,
		&user.Summary,
		&user.PrivateKey,
		&user.PublicKey,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// Post operations
func (db *DB) CreatePost(ctx context.Context, post *models.Post) error {
	query := `
        INSERT INTO posts (user_id, content, in_reply_to_id, activity_id)
        VALUES (?, ?, ?, ?)
        RETURNING id, created_at, updated_at
    `

	return db.QueryRowContext(
		ctx,
		query,
		post.UserID,
		post.Content,
		post.InReplyToID,
		post.ActivityID,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (db *DB) GetUserPosts(ctx context.Context, userID int64, limit, offset int) ([]models.Post, error) {
	query := `
        SELECT id, user_id, content, in_reply_to_id, activity_id, created_at, updated_at
        FROM posts
        WHERE user_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&post.InReplyToID,
			&post.ActivityID,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}
