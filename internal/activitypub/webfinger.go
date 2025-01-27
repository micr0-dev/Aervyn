package activitypub

import (
	"Aervyn/internal/config"
	"fmt"
)

type WebFingerResponse struct {
	Subject string          `json:"subject"`
	Links   []WebFingerLink `json:"links"`
}

type WebFingerLink struct {
	Rel  string `json:"rel"`
	Type string `json:"type,omitempty"`
	Href string `json:"href"`
}

func NewWebFingerResponse(username string) WebFingerResponse {
	actorURL := config.GetActorURL(username)
	return WebFingerResponse{
		Subject: fmt.Sprintf("acct:%s@%s", username, config.Domain),
		Links: []WebFingerLink{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: actorURL,
			},
		},
	}
}
