package updown

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipientService_List(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/recipients", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		writeJSON(w, http.StatusOK, `[
			{"id":"r1","type":"email","name":"Admin","value":"admin@example.com"},
			{"id":"r2","type":"webhook","name":"Hook","value":"https://hook.example.com"}
		]`)
	})

	recipients, resp, err := client.Recipient.List()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, recipients, 2)
	assert.Equal(t, "r1", recipients[0].ID)
	assert.Equal(t, RecipientType("email"), recipients[0].Type)
}

func TestRecipientService_List_Error(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/recipients", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, `{"message":"server error"}`)
	})

	_, resp, err := client.Recipient.List()
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestRecipientService_Add(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/recipients", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var item RecipientItem
		err := json.NewDecoder(r.Body).Decode(&item)
		require.NoError(t, err)
		assert.Equal(t, RecipientTypeEmail, item.Type)
		assert.Equal(t, "user@example.com", item.Value)

		writeJSON(w, http.StatusCreated, `{"id":"r3","type":"email","value":"user@example.com"}`)
	})

	recipient, resp, err := client.Recipient.Add(RecipientItem{
		Type:  RecipientTypeEmail,
		Value: "user@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "r3", recipient.ID)
}

func TestRecipientService_Remove(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/recipients/r1", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		writeJSON(w, http.StatusOK, `{"deleted":true}`)
	})

	deleted, resp, err := client.Recipient.Remove("r1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, deleted)
}

func TestRecipientService_Remove_NotFound(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	mux.HandleFunc("/recipients/missing", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
	})

	_, resp, err := client.Recipient.Remove("missing")
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
