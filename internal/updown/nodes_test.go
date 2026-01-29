package updown

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeService_List(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		writeJSON(w, http.StatusOK, `{
			"lan": {"ip": "1.2.3.4", "ip6": "::1", "city": "Los Angeles", "country": "United States", "country_code": "US"},
			"fra": {"ip": "5.6.7.8", "ip6": "::2", "city": "Frankfurt", "country": "Germany", "country_code": "DE"}
		}`)
	})

	nodes, resp, err := client.Node.List()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, nodes, 2)
	assert.Equal(t, "1.2.3.4", nodes["lan"].IP)
	assert.Equal(t, "Frankfurt", nodes["fra"].City)
}

func TestNodeService_List_Error(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/nodes", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, `{"message":"server error"}`)
	})

	_, resp, err := client.Node.List()
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestNodeService_ListIPv4(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/nodes/ipv4", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		writeJSON(w, http.StatusOK, `["1.2.3.4", "5.6.7.8"]`)
	})

	ips, resp, err := client.Node.ListIPv4()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, IPs{"1.2.3.4", "5.6.7.8"}, ips)
}

func TestNodeService_ListIPv6(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/nodes/ipv6", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		writeJSON(w, http.StatusOK, `["2001:db8::1", "2001:db8::2"]`)
	})

	ips, resp, err := client.Node.ListIPv6()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, IPs{"2001:db8::1", "2001:db8::2"}, ips)
}
