package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

var db *sql.DB

func InitDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./posts.db")
	if err != nil {
		return err
	}

	// Create posts table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS posts (
            id TEXT PRIMARY KEY,
            content TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	return err
}

type Post struct {
	ID        string
	Content   string
	CreatedAt time.Time
}

func CreatePost(content string) (*Post, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.Exec(
		"INSERT INTO posts (id, content, created_at) VALUES (?, ?, ?)",
		id, content, now,
	)
	if err != nil {
		return nil, err
	}

	return &Post{
		ID:        id,
		Content:   content,
		CreatedAt: now,
	}, nil
}

func GetPosts() ([]Post, error) {
	rows, err := db.Query("SELECT id, content, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Content, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}
