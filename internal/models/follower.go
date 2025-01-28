package models

import (
	"time"

	"github.com/google/uuid"
)

type Follower struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Actor     string    `json:"actor"`
	Accepted  bool      `json:"accepted"`
	CreatedAt time.Time `json:"createdAt"`
}

func CreateFollowRequest(userID, actor string) error {
	followID := uuid.New().String()
	_, err := db.Exec(`
        INSERT INTO followers (id, user_id, actor, accepted, created_at)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(user_id, actor) DO UPDATE SET
        accepted = FALSE
    `, followID, userID, actor, true, time.Now())

	return err
}

func GetFollowRequests(userID string) ([]Follower, error) {
	rows, err := db.Query(`
        SELECT id, user_id, actor, accepted, created_at
        FROM followers
        WHERE user_id = ? AND accepted = FALSE
        ORDER BY created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []Follower
	for rows.Next() {
		var f Follower
		err := rows.Scan(&f.ID, &f.UserID, &f.Actor, &f.Accepted, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		followers = append(followers, f)
	}
	return followers, nil
}

func AcceptFollowRequest(id string) error {
	_, err := db.Exec(`
        UPDATE followers
        SET accepted = TRUE
        WHERE id = ?
    `, id)
	return err
}

func RejectFollowRequest(id string) error {
	_, err := db.Exec(`
        DELETE FROM followers
        WHERE id = ?
    `, id)
	return err
}

// Check if user is following another user
func IsFollowing(userID, actor string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM followers 
            WHERE user_id = ? 
            AND actor = ? 
            AND accepted = true
        )
    `, userID, actor).Scan(&exists)
	return exists, err
}

// Get followers count
func GetFollowerCount(actor string) (int, error) {
	var count int
	err := db.QueryRow(`
        SELECT COUNT(*) 
        FROM followers 
        WHERE actor = ? 
        AND accepted = true
    `, actor).Scan(&count)
	return count, err
}

// Get following count
func GetFollowingCount(userID string) (int, error) {
	var count int
	err := db.QueryRow(`
        SELECT COUNT(*) 
        FROM followers 
        WHERE user_id = ? 
        AND accepted = true
    `, userID).Scan(&count)
	return count, err
}

// Unfollow a user
func Unfollow(userID, actor string) error {
	_, err := db.Exec(`
        DELETE FROM followers
        WHERE user_id = ? AND actor = ?
    `, userID, actor)
	return err
}
