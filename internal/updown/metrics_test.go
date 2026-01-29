package updown

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricService_List(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc/metrics", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "host", r.URL.Query().Get("group"))
		assert.Equal(t, "2024-01-01", r.URL.Query().Get("from"))
		assert.Equal(t, "2024-01-31", r.URL.Query().Get("to"))
		writeJSON(w, http.StatusOK, `{
			"lan": {"apdex": 0.99, "requests": {"samples": 100}},
			"fra": {"apdex": 0.95, "requests": {"samples": 50}}
		}`)
	})

	metrics, resp, err := client.Metric.List("abc", "host", "2024-01-01", "2024-01-31")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, metrics, 2)
	assert.Equal(t, 0.99, metrics["lan"].Apdex)
	assert.Equal(t, 100, metrics["lan"].Requests.Samples)
}

func TestMetricService_List_NoFromTo(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc/metrics", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "host", r.URL.Query().Get("group"))
		assert.Equal(t, "", r.URL.Query().Get("from"))
		assert.Equal(t, "", r.URL.Query().Get("to"))
		writeJSON(w, http.StatusOK, `{"lan": {"apdex": 0.99}}`)
	})

	metrics, resp, err := client.Metric.List("abc", "host", "", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, metrics, 1)
}

func TestMetricService_List_Error(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc/metrics", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, `{"message":"server error"}`)
	})

	_, resp, err := client.Metric.List("abc", "host", "", "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
