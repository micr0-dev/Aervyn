package main

import (
	"Aervyn/internal/config"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Initialize config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Routes
	r.Get("/", handlers.HomeHandler)
	r.Get("/.well-known/webfinger", handlers.WebFingerHandler)
	r.Get("/users/{username}", handlers.UserProfileHandler)

	// Start server
	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
