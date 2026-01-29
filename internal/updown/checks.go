// Package updown provides a Go client for the updown.io monitoring API.
package updown

import (
	"errors"
	"fmt"
	"net/http"
)

// SSL represents the SSL section of a check
type SSL struct {
	TestedAt string `json:"tested_at,omitempty"`
	Valid    bool   `json:"valid,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Check represents a check performed by Updown on a regular basis
type Check struct {
	Token             string            `json:"token,omitempty"`
	URL               string            `json:"url,omitempty"`
	Type              string            `json:"type,omitempty"`
	Alias             string            `json:"alias,omitempty"`
	LastStatus        int               `json:"last_status,omitempty"`
	Uptime            float64           `json:"uptime,omitempty"`
	Down              bool              `json:"down"`
	DownSince         string            `json:"down_since,omitempty"`
	UpSince           string            `json:"up_since,omitempty"`
	Error             string            `json:"error,omitempty"`
	Period            int               `json:"period,omitempty"`
	Apdex             float64           `json:"apdex_t,omitempty"`
	Enabled           bool              `json:"enabled"`
	Published         bool              `json:"published"`
	LastCheckAt       string            `json:"last_check_at,omitempty"`
	NextCheckAt       string            `json:"next_check_at,omitempty"`
	FaviconURL        string            `json:"favicon_url,omitempty"`
	SSL               SSL               `json:"ssl,omitempty"`
	StringMatch       string            `json:"string_match,omitempty"`
	MuteUntil         string            `json:"mute_until,omitempty"`
	DisabledLocations []string          `json:"disabled_locations,omitempty"`
	RecipientIDs      []string          `json:"recipients,omitempty"`
	CustomHeaders     map[string]string `json:"custom_headers,omitempty"`
}

// CheckItem represents a new check you want to be performed by Updown
type CheckItem struct {
	// The URL you want to monitor (not used for pulse checks)
	URL string `json:"url,omitempty"`
	// Type of check (https, http, icmp, pulse, tcp, tcps)
	Type string `json:"type,omitempty"`
	// Interval in seconds (15, 30, 60, 120, 300, 600, 1800, 3600 for regular; 15 to 2678400 for pulse)
	Period int `json:"period,omitempty"`
	// APDEX threshold in seconds (0.125, 0.25, 0.5, 1.0, 2.0, 4.0 or 8.0)
	Apdex float64 `json:"apdex_t,omitempty"`
	// Is the check enabled
	Enabled bool `json:"enabled"`
	// Shall the status page be public
	Published bool `json:"published"`
	// Human readable name
	Alias string `json:"alias,omitempty"`
	// Search for this string in the page
	StringMatch string `json:"string_match,omitempty"`
	// Mute notifications until given time, accepts a time, 'recovery' or 'forever'
	MuteUntil string `json:"mute_until,omitempty"`
	// Disabled monitoring locations. It's an array of abbreviated location names
	DisabledLocations []string `json:"disabled_locations,omitempty"`
	// Selected alert recipients. It's an array of recipient IDs
	RecipientIDs []string `json:"recipients,omitempty"`
	// The HTTP headers you want in updown requests
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// CheckService interacts with the checks section of the API
type CheckService struct {
	client *Client
	cache  Cache
}

type removeResponse struct {
	Deleted bool `json:"deleted,omitempty"`
}

// ErrTokenNotFound indicates that we cannot find a token for the given name
var ErrTokenNotFound = errors.New("Could not determine a token for the given name")

// TokenForAlias finds the Updown token for a check's alias
func (s *CheckService) TokenForAlias(name string) (string, error) {
	// Retrieve from cache
	if has, val := s.cache.Get(name); has {
		return val, nil
	}

	// List all checks
	checks, _, err := s.List()
	if err != nil {
		return "", err
	}

	// And try to find the appropriate name
	token, found := "", false
	for _, check := range checks {
		s.cache.Put(check.Alias, check.Token)
		if check.Alias == name {
			found, token = true, check.Token
		}
	}

	if found {
		return token, nil
	}

	// Could not find a match
	return "", ErrTokenNotFound
}

// List lists all the checks
func (s *CheckService) List() ([]Check, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "checks", nil)
	if err != nil {
		return nil, nil, err
	}

	var res []Check
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return nil, resp, err
	}

	return res, resp, err
}

// Get gets a single check by its token
func (s *CheckService) Get(token string) (Check, *http.Response, error) {
	req, err := s.client.NewRequest("GET", pathForToken(token), nil)
	if err != nil {
		return Check{}, nil, err
	}

	var res Check
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return Check{}, resp, err
	}

	return res, resp, err
}

// Add adds a new check you want to be performed
func (s *CheckService) Add(data CheckItem) (Check, *http.Response, error) {
	req, err := s.client.NewRequest("POST", "checks", data)
	if err != nil {
		return Check{}, nil, err
	}

	var res Check
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return Check{}, resp, err
	}

	return res, resp, err
}

// Update updates a check performed by Updown
func (s *CheckService) Update(token string, data CheckItem) (Check, *http.Response, error) {
	req, err := s.client.NewRequest("PUT", pathForToken(token), data)
	if err != nil {
		return Check{}, nil, err
	}

	var res Check
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return Check{}, resp, err
	}

	return res, resp, err
}

// Remove removes a check from Updown by its token
func (s *CheckService) Remove(token string) (bool, *http.Response, error) {
	req, err := s.client.NewRequest("DELETE", pathForToken(token), nil)
	if err != nil {
		return false, nil, err
	}

	var res removeResponse
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return false, resp, err
	}

	return res.Deleted, resp, err
}

func pathForToken(token string) string {
	return fmt.Sprintf("checks/%s", token)
}
