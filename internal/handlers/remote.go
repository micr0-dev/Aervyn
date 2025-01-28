package handlers

import (
	"Aervyn/internal/models"
	"log"
	"net/http"
	"strings"
)

func LookupHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	log.Printf("Looking up: %s", query)

	// Parse username@instance
	parts := strings.Split(query, "@")
	if len(parts) != 2 {
		http.Error(w, "Invalid format. Use username@instance", 400)
		return
	}

	username, domain := parts[0], parts[1]

	// Do WebFinger lookup
	profileURL, err := models.WebFingerLookup(username, domain)
	if err != nil {
		log.Printf("WebFinger lookup failed: %v", err)
		http.Error(w, "Failed to find user", 404)
		return
	}

	// Fetch profile
	profile, err := models.FetchRemoteProfile(profileURL)
	if err != nil {
		log.Printf("Profile fetch failed: %v", err)
		http.Error(w, "Failed to fetch profile", 500)
		return
	}

	// Render profile template
	renderTemplate(w, "profile", profile)
}
