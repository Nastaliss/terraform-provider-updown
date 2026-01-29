package updown

import (
	"fmt"
	"net/http"
)

// RecipientType represents the type of a recipient
type RecipientType string

// Recipient types
const (
	RecipientTypeEmail          RecipientType = "email"
	RecipientTypeSMS            RecipientType = "sms"
	RecipientTypeWebhook        RecipientType = "webhook"
	RecipientTypeSlackCompatible RecipientType = "slack_compatible"
	RecipientTypeMSTeams        RecipientType = "msteams"
)

// Recipient represents a recipient/channel for alerts
type Recipient struct {
	ID    string        `json:"id,omitempty"`
	Type  RecipientType `json:"type,omitempty"`
	Name  string        `json:"name,omitempty"`
	Value string        `json:"value,omitempty"`
}

// RecipientItem represents a new recipient you want to create
type RecipientItem struct {
	// Type of recipient (email, sms, webhook, slack_compatible, msteams)
	Type RecipientType `json:"type"`
	// The recipient value (email address, phone number or URL)
	Value string `json:"value"`
	// Optional user-friendly label (webhooks only)
	Name string `json:"name,omitempty"`
	// Initial state for all checks: true = selected on all existing checks
	Selected bool `json:"selected,omitempty"`
}

// RecipientService interacts with the recipients section of the API
type RecipientService struct {
	client *Client
}

// List lists all the recipients
func (s *RecipientService) List() ([]Recipient, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "recipients", nil)
	if err != nil {
		return nil, nil, err
	}

	var res []Recipient
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return nil, resp, err
	}

	return res, resp, err
}

// Add adds a new recipient
func (s *RecipientService) Add(data RecipientItem) (Recipient, *http.Response, error) {
	req, err := s.client.NewRequest("POST", "recipients", data)
	if err != nil {
		return Recipient{}, nil, err
	}

	var res Recipient
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return Recipient{}, resp, err
	}

	return res, resp, err
}

// Remove removes a recipient by its ID
func (s *RecipientService) Remove(id string) (bool, *http.Response, error) {
	req, err := s.client.NewRequest("DELETE", fmt.Sprintf("recipients/%s", id), nil)
	if err != nil {
		return false, nil, err
	}

	var res struct {
		Deleted bool `json:"deleted,omitempty"`
	}

	resp, err := s.client.Do(req, &res)
	if err != nil {
		return false, resp, err
	}

	return res.Deleted, resp, err
}
