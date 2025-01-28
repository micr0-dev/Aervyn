package models

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func FetchRemoteProfile(profileURL string) (*Profile, error) {
	log.Printf("Fetching remote profile from: %s", profileURL)

	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/activity+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch profile (status %d): %s", resp.StatusCode, string(body))
	}

	var actorData struct {
		ID                string `json:"id"`
		Type              string `json:"type"`
		PreferredUsername string `json:"preferredUsername"`
		Name              string `json:"name"`
		Summary           string `json:"summary"`
		Outbox            string `json:"outbox"`
		PublicKey         struct {
			PublicKeyPem string `json:"publicKeyPem"`
		} `json:"publicKey"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&actorData); err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(profileURL)
	if err != nil {
		return nil, err
	}

	profile := &Profile{
		ID:          actorData.ID,
		Username:    actorData.PreferredUsername,
		Domain:      parsedURL.Host,
		DisplayName: actorData.Name,
		Bio:         actorData.Summary,
		PublicKey:   actorData.PublicKey.PublicKeyPem,
		OutboxURL:   actorData.Outbox,
		IsLocal:     false,
		CreatedAt:   time.Now(),
	}

	log.Printf("Fetched profile for @%s@%s", profile.Username, profile.Domain)
	return profile, nil
}
func WebFingerLookup(username, domain string) (string, error) {
	webfingerURL := fmt.Sprintf("https://%s/.well-known/webfinger?resource=acct:%s@%s",
		domain, username, domain)

	resp, err := http.Get(webfingerURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("webfinger lookup failed: %d", resp.StatusCode)
	}

	var wf struct {
		Links []struct {
			Rel  string `json:"rel"`
			Type string `json:"type"`
			Href string `json:"href"`
		} `json:"links"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wf); err != nil {
		return "", err
	}

	for _, link := range wf.Links {
		if link.Rel == "self" && link.Type == "application/activity+json" {
			return link.Href, nil
		}
	}

	return "", fmt.Errorf("no ActivityPub profile found")
}
