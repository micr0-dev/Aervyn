package handlers

import (
	"Aervyn/internal/models"
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseGlob("web/templates/*.html"))

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := models.GetPosts()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	templates.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"Posts": posts,
	})
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content cannot be empty", 400)
		return
	}

	post, err := models.CreatePost(content)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Return just the new post HTML
	templates.ExecuteTemplate(w, "post.html", post)
}
