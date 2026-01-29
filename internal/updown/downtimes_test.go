package updown

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDowntimeService_List(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc/downtimes", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		writeJSON(w, http.StatusOK, `[
			{"error":"timeout","started_at":"2024-01-01","ended_at":"2024-01-02","duration":3600},
			{"error":"connection refused","started_at":"2024-02-01","ended_at":"2024-02-02","duration":7200}
		]`)
	})

	downs, resp, err := client.Downtime.List("abc", 1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, downs, 2)
	assert.Equal(t, "timeout", downs[0].Error)
	assert.Equal(t, 3600, downs[0].Duration)
}

func TestDowntimeService_List_Page2(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc/downtimes", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		writeJSON(w, http.StatusOK, `[]`)
	})

	_, resp, err := client.Downtime.List("abc", 2)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDowntimeService_List_NegativePage(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc/downtimes", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		writeJSON(w, http.StatusOK, `[]`)
	})

	_, resp, err := client.Downtime.List("abc", -5)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDowntimeService_List_Empty(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc/downtimes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `[]`)
	})

	downs, resp, err := client.Downtime.List("abc", 1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, downs)
}

func TestDowntimeService_List_Error(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc/downtimes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, `{"message":"server error"}`)
	})

	_, resp, err := client.Downtime.List("abc", 1)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestMaxInt(t *testing.T) {
	assert.Equal(t, 5, maxInt(5, 3))
	assert.Equal(t, 5, maxInt(3, 5))
	assert.Equal(t, 1, maxInt(1, 1))
	assert.Equal(t, 1, maxInt(1, -10))
	assert.Equal(t, 0, maxInt(0, -1))
}
