package activitypub

import "time"

type Actor struct {
	Context           []string  `json:"@context"`
	ID                string    `json:"id"`
	Type              string    `json:"type"`
	PreferredUsername string    `json:"preferredUsername"`
	Name              string    `json:"name"`
	Summary           string    `json:"summary,omitempty"`
	Inbox             string    `json:"inbox"`
	Outbox            string    `json:"outbox"`
	Following         string    `json:"following"`
	Followers         string    `json:"followers"`
	PublicKey         PublicKey `json:"publicKey"`
}

type PublicKey struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	PublicKeyPem string `json:"publicKeyPem"`
}

type OrderedCollection struct {
	Context      []string   `json:"@context,omitempty"`
	Type         string     `json:"type"`
	ID           string     `json:"id"`
	TotalItems   int        `json:"totalItems"`
	OrderedItems []Activity `json:"orderedItems,omitempty"`
	First        string     `json:"first,omitempty"`
	Last         string     `json:"last,omitempty"`
}

type Activity struct {
	Context   interface{} `json:"@context,omitempty"`
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Actor     string      `json:"actor"`
	Object    interface{} `json:"object"`
	To        []string    `json:"to,omitempty"`
	Cc        []string    `json:"cc,omitempty"`
	Published time.Time   `json:"published"`
}

type Note struct {
	Type         string    `json:"type"`
	ID           string    `json:"id"`
	Content      string    `json:"content"`
	Published    time.Time `json:"published"`
	AttributedTo string    `json:"attributedTo"`
	To           []string  `json:"to,omitempty"`
	Cc           []string  `json:"cc,omitempty"`
}

const (
	PublicAddress = "https://www.w3.org/ns/activitystreams#Public"
)
