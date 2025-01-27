package models

import (
	"database/sql"
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
        user_id TEXT NOT NULL,
        content TEXT NOT NULL,
		reply_to TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(user_id) REFERENCES users(id)
    )
`)

	// Create users table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		public_key TEXT,
		private_key TEXT
	)
`)
	if err != nil {
		return err
	}

	// Likes table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS likes (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL,
            post_id TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(user_id) REFERENCES users(id),
            FOREIGN KEY(post_id) REFERENCES posts(id),
            UNIQUE(user_id, post_id)
        )
    `)
	if err != nil {
		return err
	}

	// Boosts/Reposts table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS boosts (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL,
            post_id TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(user_id) REFERENCES users(id),
            FOREIGN KEY(post_id) REFERENCES posts(id),
            UNIQUE(user_id, post_id)
        )
    `)
	if err != nil {
		return err
	}

	// Create inbox_activities table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS inbox_activities (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL,
            activity_type TEXT NOT NULL,
            actor TEXT NOT NULL,
            object_id TEXT,
            raw_data TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            processed BOOLEAN DEFAULT FALSE,
            FOREIGN KEY(user_id) REFERENCES users(id)
        )
    `)
	if err != nil {
		return err
	}

	// Create followers table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS followers (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		actor TEXT NOT NULL,
		accepted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id),
		UNIQUE(user_id, actor)
	)
`)
	if err != nil {
		return err
	}

	return err
}
