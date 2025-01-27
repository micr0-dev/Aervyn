package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"Aervyn/internal/activitypub"
	"Aervyn/internal/config"
	"Aervyn/internal/models"

	"github.com/go-chi/chi/v5"
)

func WebFingerHandler(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	if resource == "" {
		http.Error(w, "Resource query parameter required", http.StatusBadRequest)
		return
	}

	// Parse acct:user@domain
	parts := strings.Split(strings.TrimPrefix(resource, "acct:"), "@")
	if len(parts) != 2 || parts[1] != config.Domain {
		http.Error(w, "Invalid resource", http.StatusBadRequest)
		return
	}

	username := parts[0]
	_, err := models.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := activitypub.NewWebFingerResponse(username)

	w.Header().Set("Content-Type", "application/jrd+json")
	json.NewEncoder(w).Encode(response)
}

func ActorHandler(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := models.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	actor := activitypub.Actor{
		Context: []string{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
		},
		ID:                config.GetActorURL(username),
		Type:              "Person",
		PreferredUsername: username,
		Name:              username, // You might want to add a display name field to your user model
		Inbox:             config.GetActorURL(username) + "/inbox",
		Outbox:            config.GetActorURL(username) + "/outbox",
		Following:         config.GetActorURL(username) + "/following",
		Followers:         config.GetActorURL(username) + "/followers",
		PublicKey: activitypub.PublicKey{
			ID:           config.GetActorURL(username) + "#main-key",
			Owner:        config.GetActorURL(username),
			PublicKeyPem: user.PublicKey,
		},
	}

	w.Header().Set("Content-Type", "application/activity+json")
	json.NewEncoder(w).Encode(actor)
}

func OutboxHandler(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := models.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get user's posts and convert them to activities
	posts, err := models.GetPostsByUser(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	activities := make([]activitypub.Activity, 0)
	for _, post := range posts {
		note := activitypub.Note{
			Type:         "Note",
			ID:           fmt.Sprintf("%s/posts/%s", config.InstanceURL, post.ID),
			Content:      post.Content,
			Published:    post.CreatedAt,
			AttributedTo: config.GetActorURL(username),
			To:           []string{activitypub.PublicAddress},
		}

		activity := activitypub.Activity{
			ID:        fmt.Sprintf("%s/activities/%s", config.InstanceURL, post.ID),
			Type:      "Create",
			Actor:     config.GetActorURL(username),
			Object:    note,
			Published: post.CreatedAt,
			To:        []string{activitypub.PublicAddress},
		}

		activities = append(activities, activity)
	}

	collection := activitypub.OrderedCollection{
		Context:      []string{"https://www.w3.org/ns/activitystreams"},
		Type:         "OrderedCollection",
		ID:           fmt.Sprintf("%s/users/%s/outbox", config.InstanceURL, username),
		TotalItems:   len(activities),
		OrderedItems: activities,
	}

	w.Header().Set("Content-Type", "application/activity+json")
	json.NewEncoder(w).Encode(collection)
}

func InboxHandler(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	// Only accept POST requests
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify signature
	if err := activitypub.VerifySignature(r); err != nil {
		log.Printf("Signature verification failed: %v", err)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// Parse activity
	var activity activitypub.Activity
	if err := json.Unmarshal(body, &activity); err != nil {
		http.Error(w, "Invalid activity", http.StatusBadRequest)
		return
	}

	// Get user
	user, err := models.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Store activity
	err = models.StoreInboxActivity(models.Activity{
		ID:        activity.ID,
		UserID:    user.ID,
		Type:      activity.Type,
		Actor:     activity.Actor,
		ObjectID:  fmt.Sprintf("%v", activity.Object), // Simple conversion for now
		RawData:   string(body),
		CreatedAt: activity.Published,
	})
	if err != nil {
		http.Error(w, "Error storing activity", http.StatusInternalServerError)
		return
	}

	// Acknowledge receipt
	w.WriteHeader(http.StatusAccepted)
}
