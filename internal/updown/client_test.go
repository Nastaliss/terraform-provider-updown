package updown

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setup creates a test HTTP server, a Client configured to talk to it, and a
// teardown function that must be called when the test is done.
func setup() (mux *http.ServeMux, client *Client, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	client = NewClient("test-api-key", nil)
	u, _ := url.Parse(server.URL + "/")
	client.BaseURL = u

	return mux, client, server.Close
}

// writeJSON writes a JSON response with the given status code and body string.
func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("my-key", nil)

	assert.Equal(t, "https://updown.io/api/", c.BaseURL.String())
	assert.Equal(t, "Go Updown v0.3", c.UserAgent)
	assert.Equal(t, "my-key", c.APIKey)
	assert.NotNil(t, c.Check)
	assert.NotNil(t, c.Downtime)
	assert.NotNil(t, c.Metric)
	assert.NotNil(t, c.Node)
	assert.NotNil(t, c.Recipient)
}

func TestNewClient_CustomHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c := NewClient("key", custom)

	assert.Equal(t, custom, c.client)
}

func TestNewRequest_GET(t *testing.T) {
	c := NewClient("my-key", nil)

	req, err := c.NewRequest("GET", "checks", nil)
	require.NoError(t, err)

	assert.Equal(t, "https://updown.io/api/checks", req.URL.String())
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, "Go Updown v0.3", req.Header.Get("User-Agent"))
	assert.Equal(t, "my-key", req.Header.Get("X-API-KEY"))
}

func TestNewRequest_POST_WithBody(t *testing.T) {
	c := NewClient("key", nil)

	body := CheckItem{URL: "https://example.com", Period: 60}
	req, err := c.NewRequest("POST", "checks", body)
	require.NoError(t, err)

	var decoded CheckItem
	err = json.NewDecoder(req.Body).Decode(&decoded)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", decoded.URL)
	assert.Equal(t, 60, decoded.Period)
}

func TestNewRequest_InvalidURL(t *testing.T) {
	c := NewClient("key", nil)

	_, err := c.NewRequest("GET", ":%invalid", nil)
	assert.Error(t, err)
}

func TestDo_Success(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, `[{"token":"abc","url":"https://example.com"}]`)
	})

	req, _ := client.NewRequest("GET", "checks", nil)
	var checks []Check
	resp, err := client.Do(req, &checks)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, checks, 1)
	assert.Equal(t, "abc", checks[0].Token)
}

func TestDo_WriterInterface(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/raw", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "raw body content")
	})

	req, _ := client.NewRequest("GET", "raw", nil)
	var buf bytes.Buffer
	resp, err := client.Do(req, &buf)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "raw body content", buf.String())
}

func TestDo_HTTPError(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/fail", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, `{"message":"server error"}`)
	})

	req, _ := client.NewRequest("GET", "fail", nil)
	resp, err := client.Do(req, nil)

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.IsType(t, &ErrorResponse{}, err)
}

func TestDo_NilV(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/empty", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, _ := client.NewRequest("GET", "empty", nil)
	resp, err := client.Do(req, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCheckResponse_2xx(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusOK}
	assert.NoError(t, CheckResponse(resp))

	resp = &http.Response{StatusCode: 299}
	assert.NoError(t, CheckResponse(resp))
}

func TestCheckResponse_4xx(t *testing.T) {
	body := `{"message":"not found"}`
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		Request:    &http.Request{Method: "GET", URL: &url.URL{Path: "/checks/abc"}},
	}

	err := CheckResponse(resp)
	assert.Error(t, err)
	errResp, ok := err.(*ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, errResp.Response.StatusCode)
}

func TestCheckResponse_EmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		Request:    &http.Request{Method: "GET", URL: &url.URL{Path: "/checks"}},
	}

	err := CheckResponse(resp)
	assert.Error(t, err)
	assert.IsType(t, &ErrorResponse{}, err)
}

func TestErrorResponse_Error(t *testing.T) {
	u, _ := url.Parse("https://updown.io/api/checks/abc")
	errResp := &ErrorResponse{
		Response: &http.Response{
			StatusCode: 404,
			Request:    &http.Request{Method: "GET", URL: u},
		},
		Message: "not found",
	}

	expected := "GET https://updown.io/api/checks/abc: 404 not found"
	assert.Equal(t, expected, errResp.Error())
}
