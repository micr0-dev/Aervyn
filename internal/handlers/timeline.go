package handlers

import (
	"Aervyn/internal/middleware"
	"log"
	"net/http"

	"Aervyn/internal/models"
)

func FollowingTimelineHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	posts, err := models.GetFollowingTimeline(userID)
	if err != nil {
		log.Printf("Failed to get following timeline: %v", err)
		http.Error(w, "Failed to load timeline", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Posts":         posts,
		"CurrentUserID": userID,
	}

	renderTemplate(w, "timeline", data)
}

func LocalTimelineHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")

	posts, err := models.GetLocalTimeline()
	if err != nil {
		log.Printf("Failed to get local timeline: %v", err)
		http.Error(w, "Failed to load timeline", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Posts":         posts,
		"CurrentUserID": userID,
	}

	renderTemplate(w, "timeline", data)
}
