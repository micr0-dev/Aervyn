package models

import (
	"time"
)

type User struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	DisplayName  string    `db:"display_name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Summary      string    `db:"summary"`
	PrivateKey   string    `db:"private_key"` // For ActivityPub signing
	PublicKey    string    `db:"public_key"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Post struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	Content     string    `db:"content"`
	InReplyToID *int64    `db:"in_reply_to_id"` // Optional, for replies
	ActivityID  string    `db:"activity_id"`    // ActivityPub ID
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Follow struct {
	ID          int64     `db:"id"`
	FollowerID  int64     `db:"follower_id"`
	FollowingID int64     `db:"following_id"`
	State       string    `db:"state"` // pending, accepted, rejected
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type RemoteInstance struct {
	ID        int64     `db:"id"`
	Domain    string    `db:"domain"`
	LastSeen  time.Time `db:"last_seen"`
	CreatedAt time.Time `db:"created_at"`
}
