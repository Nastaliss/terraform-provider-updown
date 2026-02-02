// Package updown provides a Go client for the updown.io monitoring API.
package updown

import (
	"fmt"
	"net/http"
)

// StatusPage represents a status page returned by the API
type StatusPage struct {
	Token       string   `json:"token,omitempty"`
	URL         string   `json:"url,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Visibility  string   `json:"visibility,omitempty"`
	AccessKey   string   `json:"access_key,omitempty"`
	Checks      []string `json:"checks,omitempty"`
}

// StatusPageItem represents a status page you want to create or update
type StatusPageItem struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Visibility  string   `json:"visibility,omitempty"`
	Checks      []string `json:"checks"`
}

// StatusPageService interacts with the status_pages section of the API
type StatusPageService struct {
	client *Client
}

// List lists all the status pages
func (s *StatusPageService) List() ([]StatusPage, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "status_pages", nil)
	if err != nil {
		return nil, nil, err
	}

	var res []StatusPage
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return nil, resp, err
	}

	return res, resp, err
}

// Add creates a new status page
func (s *StatusPageService) Add(data StatusPageItem) (StatusPage, *http.Response, error) {
	req, err := s.client.NewRequest("POST", "status_pages", data)
	if err != nil {
		return StatusPage{}, nil, err
	}

	var res StatusPage
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return StatusPage{}, resp, err
	}

	return res, resp, err
}

// Update updates an existing status page by its token
func (s *StatusPageService) Update(token string, data StatusPageItem) (StatusPage, *http.Response, error) {
	req, err := s.client.NewRequest("PUT", fmt.Sprintf("status_pages/%s", token), data)
	if err != nil {
		return StatusPage{}, nil, err
	}

	var res StatusPage
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return StatusPage{}, resp, err
	}

	return res, resp, err
}

// Remove removes a status page by its token
func (s *StatusPageService) Remove(token string) (bool, *http.Response, error) {
	req, err := s.client.NewRequest("DELETE", fmt.Sprintf("status_pages/%s", token), nil)
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
