package updown

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusPageService_List(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/status_pages", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		writeJSON(w, http.StatusOK, `[
			{"token":"sp1","url":"https://status.example.com","name":"My Status","description":"Desc","visibility":"public","checks":["aaaa","bbbb"]},
			{"token":"sp2","url":"https://status2.example.com","name":"Other","visibility":"private","checks":["cccc"]}
		]`)
	})

	pages, resp, err := client.StatusPage.List()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, pages, 2)
	assert.Equal(t, "sp1", pages[0].Token)
	assert.Equal(t, "public", pages[0].Visibility)
	assert.Equal(t, []string{"aaaa", "bbbb"}, pages[0].Checks)
}

func TestStatusPageService_List_Error(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/status_pages", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, `{"message":"server error"}`)
	})

	_, resp, err := client.StatusPage.List()
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestStatusPageService_Add(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/status_pages", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var item StatusPageItem
		err := json.NewDecoder(r.Body).Decode(&item)
		require.NoError(t, err)
		assert.Equal(t, "My Page", item.Name)
		assert.Equal(t, []string{"aaaa", "bbbb"}, item.Checks)

		writeJSON(w, http.StatusCreated, `{"token":"sp3","url":"https://status3.example.com","name":"My Page","visibility":"public","checks":["aaaa","bbbb"]}`)
	})

	page, resp, err := client.StatusPage.Add(StatusPageItem{
		Name:   "My Page",
		Checks: []string{"aaaa", "bbbb"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "sp3", page.Token)
	assert.Equal(t, "My Page", page.Name)
}

func TestStatusPageService_Add_Error(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/status_pages", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusUnprocessableEntity, `{"message":"checks is required"}`)
	})

	_, resp, err := client.StatusPage.Add(StatusPageItem{})
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestStatusPageService_Update(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/status_pages/sp1", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		var item StatusPageItem
		err := json.NewDecoder(r.Body).Decode(&item)
		require.NoError(t, err)
		assert.Equal(t, "Updated Page", item.Name)

		writeJSON(w, http.StatusOK, `{"token":"sp1","url":"https://status.example.com","name":"Updated Page","visibility":"public","checks":["aaaa"]}`)
	})

	page, resp, err := client.StatusPage.Update("sp1", StatusPageItem{
		Name:   "Updated Page",
		Checks: []string{"aaaa"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Updated Page", page.Name)
}

func TestStatusPageService_Update_Error(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/status_pages/sp1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
	})

	_, resp, err := client.StatusPage.Update("sp1", StatusPageItem{
		Checks: []string{"aaaa"},
	})
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestStatusPageService_Remove(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/status_pages/sp1", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		writeJSON(w, http.StatusOK, `{"deleted":true}`)
	})

	deleted, resp, err := client.StatusPage.Remove("sp1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, deleted)
}

func TestStatusPageService_Remove_NotFound(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/status_pages/missing", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
	})

	_, resp, err := client.StatusPage.Remove("missing")
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
