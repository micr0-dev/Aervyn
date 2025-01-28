package handlers

import (
	"Aervyn/internal/utils"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var templates *template.Template

func init() {
	funcMap := template.FuncMap{
		"formatTime": formatTime,
		"sanitize":   utils.SanitizeHTML,
	}

	templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html"))
	log.Printf("Loaded templates: %v", templates.DefinedTemplates())
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "yesterday"
	default:
		return t.Format("Jan 2")
	}
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
