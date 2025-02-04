// cmd/main.go
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	"Aervyn/internal/handlers"
	"Aervyn/internal/middleware"
	"Aervyn/internal/models"
)

func main() {
	// Initialize database
	if err := models.InitDB(); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.SessionManager.LoadAndSave)

	// File server for static files
	fileServer := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/login", handlers.LoginHandler)
		r.Post("/login", handlers.LoginHandler)
		r.Get("/register", handlers.RegisterHandler)
		r.Post("/register", handlers.RegisterHandler)
		r.Get("/.well-known/webfinger", handlers.WebFingerHandler)
		r.Get("/users/{username}", handlers.ActorHandler)
		r.Get("/users/{username}/outbox", handlers.OutboxHandler)
		r.Post("/users/{username}/inbox", handlers.InboxHandler)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Get("/", handlers.HomeHandler)
		r.Post("/posts", handlers.CreatePost)
		r.Get("/logout", handlers.LogoutHandler)
		r.Post("/posts/{postID}/like", handlers.LikeHandler)
		r.Post("/posts/{postID}/unlike", handlers.UnlikeHandler)
		r.Post("/posts/{postID}/boost", handlers.BoostHandler)
		r.Post("/posts/{postID}/unboost", handlers.UnboostHandler)
		r.Get("/posts/{postID}/reply-form", handlers.ReplyFormHandler)
		r.Post("/posts/{postID}/reply", handlers.ReplyHandler)
		r.Get("/remote/lookup", handlers.LookupHandler)
		r.Get("/@{identifier}", handlers.ProfileHandler)
		r.Get("/profile/edit", handlers.ProfileEditHandler)
		r.Put("/profile", handlers.ProfileUpdateHandler)
		r.Post("/follow/{username}", handlers.FollowHandler)
		r.Delete("/follow/{username}", handlers.UnfollowHandler)
		r.Get("/timeline/following", handlers.FollowingTimelineHandler)
		r.Get("/timeline/local", handlers.LocalTimelineHandler)
	})

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
