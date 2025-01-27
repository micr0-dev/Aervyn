package models

import (
	"fmt"
	"log"
	"time"
)

type Activity struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Type      string    `json:"type"`
	Actor     string    `json:"actor"`
	ObjectID  string    `json:"object,omitempty"`
	RawData   string    `json:"-"`
	CreatedAt time.Time `json:"published"`
	Processed bool      `json:"-"`
}

func StoreInboxActivity(activity *Activity) error {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM inbox_activities WHERE id = ?)",
		activity.ID,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	_, err = db.Exec(`
        INSERT INTO inbox_activities 
        (id, user_id, activity_type, actor, object_id, raw_data, created_at, processed)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `,
		activity.ID,
		activity.UserID,
		activity.Type,
		activity.Actor,
		activity.ObjectID,
		activity.RawData,
		activity.CreatedAt,
		false,
	)
	return err
}

func (a *Activity) ProcessActivity() error {
	log.Printf("Processing activity type: %s", a.Type)

	switch a.Type {
	case "Follow":
		return CreateFollowRequest(a.UserID, a.Actor)
	case "Like":
		// TODO: Implement like handling
		return nil
	case "Announce": // Boost
		// TODO: Implement boost handling
		return nil
	case "Undo":
		// TODO: Implement undo handling
		return nil
	default:
		return fmt.Errorf("unknown activity type: %s", a.Type)
	}
}
