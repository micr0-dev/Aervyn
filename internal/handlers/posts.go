package handlers

import (
	"Aervyn/internal/middleware"
	"Aervyn/internal/models"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	if userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := map[string]interface{}{
		"PageTitle":     "Home",
		"Username":      user.Username,
		"CurrentUserID": userID,
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

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	postID := chi.URLParam(r, "postID")

	err := models.LikePost(postID, userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Get updated post
	post, err := models.GetPost(postID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Load user interactions
	err = post.LoadUserInteractions(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	renderTemplate(w, "post", post)
}

func UnlikeHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	postID := chi.URLParam(r, "postID")

	err := models.UnlikePost(postID, userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Return updated post HTML
	post, err := models.GetPost(postID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Load user interactions
	err = post.LoadUserInteractions(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	renderTemplate(w, "post", post)
}

func BoostHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	postID := chi.URLParam(r, "postID")

	err := models.BoostPost(postID, userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Return updated post HTML
	post, err := models.GetPost(postID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Load user interactions
	err = post.LoadUserInteractions(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	renderTemplate(w, "post", post)
}

func UnboostHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	postID := chi.URLParam(r, "postID")

	err := models.UnboostPost(postID, userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Return updated post HTML
	post, err := models.GetPost(postID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Load user interactions
	err = post.LoadUserInteractions(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	renderTemplate(w, "post", post)
}

func ReplyFormHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "postID")
	renderTemplate(w, "reply-form", map[string]interface{}{
		"ID": postID,
	})
}

func ReplyHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.SessionManager.GetString(r.Context(), "userID")
	postID := chi.URLParam(r, "postID")

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content cannot be empty", 400)
		return
	}

	reply, err := models.CreateReply(content, postID, userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Load necessary data
	err = reply.LoadUserInteractions(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = reply.LoadParentPost()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Render just the new reply
	renderTemplate(w, "post", reply)
}
