package activitypub

import (
	"Aervyn/internal/config"
	"fmt"

	"github.com/google/uuid"
)

func createActivity(activityType string, actor string, object interface{}) Activity {
	return Activity{
		Context: "https://www.w3.org/ns/activitystreams",
		Type:    activityType,
		Actor:   actor,
		Object:  object,
		ID:      fmt.Sprintf("%s/activities/%s", config.InstanceURL, uuid.New().String()),
	}
}

func sendActivity(activity Activity, inbox string) error {
	// Sign request
	// Send HTTP POST
	// Handle response
	return nil
}
