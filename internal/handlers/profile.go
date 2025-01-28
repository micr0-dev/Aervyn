package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"Aervyn/internal/middleware"
	"Aervyn/internal/models"
	"Aervyn/internal/utils"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Get the profile identifier from the URL (e.g., "micr0" or "gargron@mastodon.social")
	identifier := chi.URLParam(r, "identifier")
	identifier, err := utils.ValidateAndNormalizeUsername(identifier)
	if err != nil {
		log.Printf("Invalid username format: %v", err)
		http.Error(w, "Invalid username format", http.StatusBadRequest)
		return
	}
	log.Printf("Looking up profile: %s", identifier)

	var profile *models.Profile

	currentUserID := middleware.SessionManager.GetString(r.Context(), "userID")

	// Check if it's a remote profile (contains @)
	if strings.Contains(identifier, "@") {
		// Remote profile
		parts := strings.Split(identifier, "@")
		if len(parts) != 2 {
			http.Error(w, "Invalid profile format", http.StatusBadRequest)
			return
		}

		username, domain := parts[0], parts[1]

		// Do WebFinger lookup
		profileURL, err := models.WebFingerLookup(username, domain)
		if err != nil {
			log.Printf("WebFinger lookup failed: %v", err)
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}

		profile, err = models.FetchRemoteProfile(profileURL)
		if err != nil {
			log.Printf("Failed to fetch remote profile: %v", err)
			http.Error(w, "Failed to fetch profile", http.StatusInternalServerError)
			return
		}
	} else {
		// Local profile
		profile, err = models.GetProfileByUsername(identifier)
		if err != nil {
			log.Printf("Failed to fetch local profile: %v", err)
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
	}

	// Get follower and following count
	followerCount, err := models.GetFollowerCount(profile.ID)
	if err != nil {
		log.Printf("Failed to get follower count: %v", err)
		followerCount = 0
	}

	followingCount, err := models.GetFollowingCount(profile.ID)
	if err != nil {
		log.Printf("Failed to get following count: %v", err)
		followingCount = 0
	}

	var isFollowing bool
	if currentUserID != "" {
		isFollowing, err = models.IsFollowing(currentUserID, profile.ID)
		if err != nil {
			log.Printf("Failed to check following status: %v", err)
			isFollowing = false
		}
	}
	// Get posts for the profile
	posts, err := models.GetPostsForProfile(profile)
	if err != nil {
		log.Printf("Failed to fetch posts: %v", err)
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Profile":        profile,
		"Posts":          posts,
		"PageTitle":      fmt.Sprintf("@%s", profile.Username),
		"CurrentUserID":  currentUserID,
		"FollowerCount":  followerCount,
		"FollowingCount": followingCount,
		"IsFollowing":    isFollowing,
	}

	log.Printf("Rendering profile page for: %s", profile.Username)
	renderTemplate(w, "layout.html", data)
}

func ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := models.GetProfileByID(userID)
	if err != nil {
		log.Printf("Failed to get profile: %v", err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Profile":       profile,
		"CurrentUserID": userID,
	}

	renderTemplate(w, "profile_edit", data)
}

func ProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	displayName := r.FormValue("displayName")
	bio := r.FormValue("bio")

	user := &models.User{ID: userID}
	if err := user.UpdateProfile(displayName, bio); err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	profile, err := models.GetProfileByID(userID)
	if err != nil {
		log.Printf("Failed to get updated profile: %v", err)
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/@%s", profile.Username), http.StatusSeeOther)
}
