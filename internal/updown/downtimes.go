// Package updown provides a Go client for the updown.io monitoring API.
package updown

import (
	"fmt"
	"net/http"
	"strconv"
)

// Downtime represents a downtime period for a check
type Downtime struct {
	Error     string `json:"error,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
	EndedAt   string `json:"ended_at,omitempty"`
	Duration  int    `json:"duration,omitempty"`
}

// DowntimeService interacts with the downtimes section of the API
type DowntimeService struct {
	client *Client
}

// List lists all known downtimes for a check
func (s *DowntimeService) List(token string, pageNb int) ([]Downtime, *http.Response, error) {
	path := fmt.Sprintf("checks/%s/downtimes?page=%s", token, strconv.Itoa(maxInt(1, pageNb)))
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var res []Downtime
	resp, err := s.client.Do(req, &res)
	if err != nil {
		return nil, resp, err
	}

	return res, resp, err
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
