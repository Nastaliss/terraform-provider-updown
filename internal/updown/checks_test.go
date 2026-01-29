package updown

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckService_List(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		writeJSON(w, http.StatusOK, `[
			{"token":"abc","url":"https://example.com","alias":"Example"},
			{"token":"def","url":"https://test.com","alias":"Test"}
		]`)
	})

	checks, resp, err := client.Check.List()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, checks, 2)
	assert.Equal(t, "abc", checks[0].Token)
	assert.Equal(t, "def", checks[1].Token)
}

func TestCheckService_List_Error(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, `{"message":"internal error"}`)
	})

	_, resp, err := client.Check.List()
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestCheckService_Get(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		writeJSON(w, http.StatusOK, `{"token":"abc","url":"https://example.com","alias":"Example"}`)
	})

	check, resp, err := client.Check.Get("abc")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "abc", check.Token)
	assert.Equal(t, "Example", check.Alias)
}

func TestCheckService_Get_NotFound(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/missing", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
	})

	_, resp, err := client.Check.Get("missing")
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCheckService_Add(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var item CheckItem
		err := json.NewDecoder(r.Body).Decode(&item)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", item.URL)
		assert.Equal(t, 60, item.Period)

		writeJSON(w, http.StatusCreated, `{"token":"new","url":"https://example.com"}`)
	})

	check, resp, err := client.Check.Add(CheckItem{URL: "https://example.com", Period: 60})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "new", check.Token)
}

func TestCheckService_Update(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		var item CheckItem
		err := json.NewDecoder(r.Body).Decode(&item)
		require.NoError(t, err)
		assert.Equal(t, "https://updated.com", item.URL)

		writeJSON(w, http.StatusOK, `{"token":"abc","url":"https://updated.com"}`)
	})

	check, resp, err := client.Check.Update("abc", CheckItem{URL: "https://updated.com"})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "https://updated.com", check.URL)
}

func TestCheckService_Remove(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/abc", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		writeJSON(w, http.StatusOK, `{"deleted":true}`)
	})

	deleted, resp, err := client.Check.Remove("abc")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, deleted)
}

func TestCheckService_Remove_NotFound(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks/missing", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
	})

	_, resp, err := client.Check.Remove("missing")
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCheckService_TokenForAlias_Found(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `[
			{"token":"t1","alias":"Site A"},
			{"token":"t2","alias":"Site B"}
		]`)
	})

	token, err := client.Check.TokenForAlias("Site B")
	require.NoError(t, err)
	assert.Equal(t, "t2", token)
}

func TestCheckService_TokenForAlias_NotFound(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/checks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `[{"token":"t1","alias":"Site A"}]`)
	})

	token, err := client.Check.TokenForAlias("Nonexistent")
	assert.Equal(t, "", token)
	assert.Equal(t, ErrTokenNotFound, err)
}

func TestCheckService_TokenForAlias_Cached(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	var callCount int64
	mux.HandleFunc("/checks", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&callCount, 1)
		writeJSON(w, http.StatusOK, fmt.Sprintf(`[{"token":"t1","alias":"Site A"}]`))
	})

	// First call: cache miss, hits server
	token, err := client.Check.TokenForAlias("Site A")
	require.NoError(t, err)
	assert.Equal(t, "t1", token)
	assert.Equal(t, int64(1), atomic.LoadInt64(&callCount))

	// Second call: cache hit, no server call
	token, err = client.Check.TokenForAlias("Site A")
	require.NoError(t, err)
	assert.Equal(t, "t1", token)
	assert.Equal(t, int64(1), atomic.LoadInt64(&callCount))
}
