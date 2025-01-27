package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        string
	UserID    string
	Username  string
	Content   string
	CreatedAt time.Time
}

func GetPosts() ([]Post, error) {
	rows, err := db.Query(`
        SELECT 
            p.id, 
            p.user_id, 
            u.username, 
            p.content, 
            p.created_at 
        FROM posts p 
        JOIN users u ON p.user_id = u.id 
        ORDER BY p.created_at DESC
        LIMIT 50
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Content, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func GetPostsByUser(userID string) ([]Post, error) {
	rows, err := db.Query(`
        SELECT 
            p.id, 
            p.user_id, 
            u.username, 
            p.content, 
            p.created_at 
        FROM posts p 
        JOIN users u ON p.user_id = u.id 
        WHERE p.user_id = ?
        ORDER BY p.created_at DESC
        LIMIT 20
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Content, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func CreatePost(content string, userID string) (*Post, error) {
	id := uuid.New().String()
	now := time.Now()

	// First create the post
	_, err := db.Exec(
		"INSERT INTO posts (id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
		id, userID, content, now,
	)
	if err != nil {
		return nil, err
	}

	// Then get the username for the response
	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		return nil, err
	}

	return &Post{
		ID:        id,
		UserID:    userID,
		Username:  username,
		Content:   content,
		CreatedAt: now,
	}, nil
}
