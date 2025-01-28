package utils

import (
	"html"
	"regexp"
	"strings"
)

var (
	mentionRegex  = regexp.MustCompile(`<span[^>]*><a href="([^"]+)"[^>]*>@<span>([^<]+)</span></a></span>`)
	emphasisRegex = regexp.MustCompile(`<em>([^<]+)</em>`)
)

func SanitizeHTML(content string) string {
	// Convert HTML entities
	content = html.UnescapeString(content)

	// Handle mentions
	content = mentionRegex.ReplaceAllString(content, "@$2")

	// Handle emphasis
	content = emphasisRegex.ReplaceAllString(content, "_$1_")

	// Remove any remaining HTML tags
	content = regexp.MustCompile("<[^>]*>").ReplaceAllString(content, "")

	// Clean up extra whitespace
	content = strings.TrimSpace(content)

	return content
}
