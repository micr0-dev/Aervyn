package models

import (
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

func StoreInboxActivity(activity Activity) error {
	_, err := db.Exec(`
        INSERT INTO inbox_activities 
        (id, user_id, activity_type, actor, object_id, raw_data, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `,
		activity.ID,
		activity.UserID,
		activity.Type,
		activity.Actor,
		activity.ObjectID,
		activity.RawData,
		activity.CreatedAt,
	)
	return err
}
