package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Profile struct {
	ID          string    `json:"id"`
	Username    string    `json:"preferredUsername"`
	Domain      string    `json:"domain,omitempty"` // empty for local users
	DisplayName string    `json:"name,omitempty"`
	Bio         string    `json:"summary,omitempty"`
	PublicKey   string    `json:"publicKey,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	IsLocal     bool      `json:"-"`
	OutboxURL   string    `json:"outbox,omitempty"`
}

func GetProfileByUsername(username string) (*Profile, error) {
	var (
		profile     Profile
		displayName sql.NullString
		bio         sql.NullString
		publicKey   sql.NullString
	)

	err := db.QueryRow(`
        SELECT 
            id, 
            username,
            display_name,
            bio, 
            created_at,
            public_key
        FROM users 
        WHERE username = ?
    `, username).Scan(
		&profile.ID,
		&profile.Username,
		&displayName,
		&bio,
		&profile.CreatedAt,
		&publicKey,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch local profile: %w", err)
	}

	// Convert NULL strings to empty strings
	if displayName.Valid {
		profile.DisplayName = displayName.String
	}
	if bio.Valid {
		profile.Bio = bio.String
	}
	if publicKey.Valid {
		profile.PublicKey = publicKey.String
	}

	profile.IsLocal = true
	return &profile, nil
}

func GetPostsForProfile(profile *Profile) ([]Post, error) {
	if profile.IsLocal {
		// Fetch local posts
		return GetPostsByUserID(profile.ID)
	} else {
		// Fetch or cache remote posts
		return FetchRemotePosts(profile)
	}
}

func FetchRemotePosts(profile *Profile) ([]Post, error) {
	log.Printf("Fetching posts for remote profile: %s", profile.ID)

	// Fetch outbox
	outboxURL := profile.ID + "/outbox"
	log.Printf("Fetching outbox from: %s", outboxURL)

	req, err := http.NewRequest("GET", outboxURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/activity+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch outbox: %w", err)
	}
	defer resp.Body.Close()

	// Parse the initial outbox response
	var outboxResp struct {
		First string `json:"first"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&outboxResp); err != nil {
		return nil, fmt.Errorf("failed to parse outbox: %w", err)
	}

	// Fetch the first page
	log.Printf("Fetching first page: %s", outboxResp.First)
	req, err = http.NewRequest("GET", outboxResp.First, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for first page: %w", err)
	}
	req.Header.Set("Accept", "application/activity+json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch first page: %w", err)
	}
	defer resp.Body.Close()

	// Read and log the response for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	log.Printf("First page response: %s", string(body))

	// Parse the page with a more flexible structure
	var page struct {
		OrderedItems []json.RawMessage `json:"orderedItems"`
	}
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("failed to parse page: %w", err)
	}

	var posts []Post
	for _, itemRaw := range page.OrderedItems {
		// First try to parse as an activity
		var activity struct {
			Type   string          `json:"type"`
			Object json.RawMessage `json:"object"`
		}
		if err := json.Unmarshal(itemRaw, &activity); err != nil {
			log.Printf("Failed to parse activity: %v", err)
			continue
		}

		var postContent struct {
			Type      string    `json:"type"`
			ID        string    `json:"id"`
			Content   string    `json:"content"`
			Published time.Time `json:"published"`
		}

		// If it's a Create activity, parse the object
		if activity.Type == "Create" {
			if err := json.Unmarshal(activity.Object, &postContent); err != nil {
				log.Printf("Failed to parse post content: %v", err)
				continue
			}
		} else {
			// Try parsing the item directly as a post
			if err := json.Unmarshal(itemRaw, &postContent); err != nil {
				log.Printf("Failed to parse as direct post: %v", err)
				continue
			}
		}

		// Only process Notes
		if postContent.Type != "Note" {
			continue
		}

		post := Post{
			ID:       postContent.ID,
			Content:  postContent.Content,
			AuthorID: profile.ID,
			Author: Profile{
				ID:          profile.ID,
				Username:    profile.Username,
				Domain:      profile.Domain,
				DisplayName: profile.DisplayName,
				IsLocal:     false,
			},
			CreatedAt: postContent.Published,
			IsLocal:   false,
			URL:       postContent.ID,
		}
		posts = append(posts, post)
	}

	log.Printf("Found %d posts for %s", len(posts), profile.Username)
	return posts, nil
}

func GetProfileByID(userID string) (*Profile, error) {
	var profile Profile
	var displayName, bio sql.NullString

	err := db.QueryRow(`
        SELECT 
            id, 
            username,
            display_name,
            bio, 
            created_at
        FROM users 
        WHERE id = ?
    `, userID).Scan(
		&profile.ID,
		&profile.Username,
		&displayName,
		&bio,
		&profile.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if displayName.Valid {
		profile.DisplayName = displayName.String
	}
	if bio.Valid {
		profile.Bio = bio.String
	}

	profile.IsLocal = true
	return &profile, nil
}
