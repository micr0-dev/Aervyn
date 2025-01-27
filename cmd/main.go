// cmd/main.go
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	"Aervyn/internal/handlers"
	"Aervyn/internal/models"
)

func main() {
	// Initialize database
	if err := models.InitDB(); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Routes
	r.Get("/", handlers.HomeHandler)
	r.Post("/posts", handlers.CreatePost)

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
