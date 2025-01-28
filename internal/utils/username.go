package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func ValidateAndNormalizeUsername(username string) (string, error) {
	// Remove @ prefix if present
	username = strings.TrimPrefix(username, "@")

	// Split into username and domain if remote
	parts := strings.Split(username, "@")
	localPart := strings.ToLower(parts[0]) // Convert to lowercase

	// Validate local part
	if !usernameRegex.MatchString(localPart) {
		return "", fmt.Errorf("username can only contain letters, numbers, and underscores")
	}

	if len(localPart) < 1 || len(localPart) > 30 {
		return "", fmt.Errorf("username must be between 1 and 30 characters")
	}

	// If it's a remote user, validate and normalize domain
	if len(parts) > 1 {
		domain := strings.ToLower(parts[1]) // Convert domain to lowercase
		return fmt.Sprintf("%s@%s", localPart, domain), nil
	}

	return localPart, nil
}
