package handlers

import (
	"Aervyn/internal/middleware"
	"Aervyn/internal/models"
	"Aervyn/internal/utils"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func FollowHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetUsername := chi.URLParam(r, "username")
	// Remove @ prefix if present
	targetUsername = strings.TrimPrefix(targetUsername, "@")
	targetUsername, err := utils.ValidateAndNormalizeUsername(targetUsername)
	if err != nil {
		log.Printf("Invalid username format: %v", err)
		http.Error(w, "Invalid username format", http.StatusBadRequest)
		return
	}

	isRemote := strings.Contains(targetUsername, "@")

	var actorURI string
	if isRemote {
		// Handle remote follow
		parts := strings.Split(targetUsername, "@")
		if len(parts) != 2 {
			http.Error(w, "Invalid username format", http.StatusBadRequest)
			return
		}
		username, domain := parts[0], parts[1]

		// Get ActivityPub actor URI through WebFinger
		var err error
		actorURI, err = models.WebFingerLookup(username, domain)
		if err != nil {
			log.Printf("WebFinger lookup failed: %v", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
	} else {
		// Handle local follow
		profile, err := models.GetProfileByUsername(targetUsername)
		if err != nil {
			log.Printf("Failed to get local profile: %v", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		actorURI = profile.ID
	}

	log.Printf("Creating follow request from %s to actor %s", userID, actorURI)
	err = models.CreateFollowRequest(userID, actorURI)
	if err != nil {
		log.Printf("Failed to create follow request: %v", err)
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		return
	}

	// Return updated follow button with correct format
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
        <button class="unfollow-btn" 
                hx-delete="/follow/@` + targetUsername + `" 
                hx-target="this" 
                hx-swap="outerHTML">
            Unfollow
        </button>
    `))
}

func UnfollowHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetUsername := chi.URLParam(r, "username")
	isRemote := strings.Contains(targetUsername, "@")

	var actorURI string
	if isRemote {
		parts := strings.Split(targetUsername, "@")
		if len(parts) != 2 {
			http.Error(w, "Invalid username format", http.StatusBadRequest)
			return
		}
		username, domain := parts[0], parts[1]
		var err error
		actorURI, err = models.WebFingerLookup(username, domain)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
	} else {
		profile, err := models.GetProfileByUsername(targetUsername)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		actorURI = profile.ID
	}

	err := models.Unfollow(userID, actorURI)
	if err != nil {
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}

	// Return updated follow button
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
        <button class="follow-btn" 
                hx-post="/follow/` + targetUsername + `" 
                hx-target="this" 
                hx-swap="outerHTML">
            Follow
        </button>
    `))
}
