package handlers

import (
	"Aervyn/internal/middleware"
	"Aervyn/internal/models"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	if userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get posts
	posts, err := models.GetPosts()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Get username for display
	user, err := models.GetUserByID(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := map[string]interface{}{
		"PageTitle": "Home",
		"Username":  user.Username,
		"Posts":     posts,
	}

	renderTemplate(w, "layout.html", data)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content cannot be empty", 400)
		return
	}

	post, err := models.CreatePost(content, userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Render just the new post
	renderTemplate(w, "post", post)
}
