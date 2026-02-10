package ssf

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	return New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
	}, slog.Default())
}

func TestGetMonthlyContentEvents(t *testing.T) {
	response := ContentEventsResponse{
		Items: []ContentEvent{
			{
				PersonID:        "person-1",
				TaskID:          "task-1",
				PlanID:          "plan-1",
				Timestamp:       time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
				ContentProgress: 0.75,
			},
			{
				PersonID:        "person-2",
				TaskID:          "task-2",
				PlanID:          "plan-1",
				Timestamp:       time.Date(2025, 6, 16, 14, 0, 0, 0, time.UTC),
				ContentProgress: 1.0,
			},
		},
		Page:    1,
		HasMore: true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/bcc/content-events/monthly", r.URL.Path)
		assert.Equal(t, "2025", r.URL.Query().Get("year"))
		assert.Equal(t, "6", r.URL.Query().Get("month"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "test-key", r.Header.Get("X-SSSF-API-Key"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.GetMonthlyContentEvents(context.Background(), 2025, 6, 1)

	require.NoError(t, err)
	assert.Equal(t, 2, len(result.Items))
	assert.Equal(t, 1, result.Page)
	assert.True(t, result.HasMore)

	assert.Equal(t, "person-1", result.Items[0].PersonID)
	assert.Equal(t, "task-1", result.Items[0].TaskID)
	assert.Equal(t, "plan-1", result.Items[0].PlanID)
	assert.Equal(t, 0.75, result.Items[0].ContentProgress)

	assert.Equal(t, "person-2", result.Items[1].PersonID)
	assert.Equal(t, 1.0, result.Items[1].ContentProgress)
}

func TestGetMonthlyContentEvents_EmptyResponse(t *testing.T) {
	response := ContentEventsResponse{
		Items:   []ContentEvent{},
		Page:    1,
		HasMore: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.GetMonthlyContentEvents(context.Background(), 2025, 1, 1)

	require.NoError(t, err)
	assert.Empty(t, result.Items)
	assert.False(t, result.HasMore)
}

func TestGetMonthlyContentEvents_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.GetMonthlyContentEvents(context.Background(), 2025, 6, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get monthly content events")
}

func TestGetMemberContentEvents(t *testing.T) {
	response := ContentEventsResponse{
		Items: []ContentEvent{
			{
				PersonID:        "person-abc",
				TaskID:          "task-10",
				PlanID:          "plan-5",
				Timestamp:       time.Date(2025, 3, 20, 8, 0, 0, 0, time.UTC),
				ContentProgress: 0.5,
			},
		},
		Page:    1,
		HasMore: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/bcc/content-events/member/person-abc", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "test-key", r.Header.Get("X-SSSF-API-Key"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.GetMemberContentEvents(context.Background(), "person-abc", 1)

	require.NoError(t, err)
	assert.Equal(t, 1, len(result.Items))
	assert.Equal(t, 1, result.Page)
	assert.False(t, result.HasMore)

	assert.Equal(t, "person-abc", result.Items[0].PersonID)
	assert.Equal(t, "task-10", result.Items[0].TaskID)
	assert.Equal(t, "plan-5", result.Items[0].PlanID)
	assert.Equal(t, 0.5, result.Items[0].ContentProgress)
}

func TestGetMemberContentEvents_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.GetMemberContentEvents(context.Background(), "nonexistent", 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get content events for member")
}
