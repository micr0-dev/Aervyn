package handlers

import (
	"Aervyn/internal/middleware"
	"Aervyn/internal/models"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := map[string]interface{}{
			"PageTitle": "Login",
		}
		renderTemplate(w, "layout.html", data)
		return
	}

	// Handle POST
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := models.GetUserByUsername(username)
	if err != nil || !user.CheckPassword(password) {
		renderTemplate(w, "layout.html", map[string]interface{}{
			"PageTitle": "Login",
			"Error":     "Invalid username or password",
		})
		return
	}

	middleware.SessionManager.Put(r.Context(), "userID", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := map[string]interface{}{
			"PageTitle": "Register",
		}
		renderTemplate(w, "layout.html", data)
		return
	}

	// Handle POST
	username := r.FormValue("username")
	password := r.FormValue("password")

	_, err := models.CreateUser(username, password)
	if err != nil {
		renderTemplate(w, "layout.html", map[string]interface{}{
			"PageTitle": "Register",
			"Error":     "Error creating user",
		})
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	middleware.SessionManager.Destroy(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
