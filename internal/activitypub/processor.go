package activitypub

import "fmt"

func processActivity(activity Activity) error {
	switch activity.Type {
	case "Follow":
		return handleFollow(activity)
	case "Like":
		return handleLike(activity)
	case "Announce": // Boost
		return handleAnnounce(activity)
	case "Create":
		return handleCreate(activity)
	case "Undo":
		return handleUndo(activity)
	default:
		return fmt.Errorf("unknown activity type: %s", activity.Type)
	}
}

func handleFollow(activity Activity) error {
	// Store follow request
	// Send Accept activity
	// Update followers list
	return nil
}

func handleLike(activity Activity) error {
	// Store like
	// Update like count
	// Send notification
	return nil
}

func handleAnnounce(activity Activity) error {
	// Store boost
	// Send notification
	return nil
}

func handleCreate(activity Activity) error {
	// Store post
	// Send notification
	return nil
}

func handleUndo(activity Activity) error {
	// Undo previous activity
	return nil
}
