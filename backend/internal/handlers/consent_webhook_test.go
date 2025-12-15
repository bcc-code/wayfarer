package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsentEventRequest_ValidationErrors(t *testing.T) {
	// These tests only cover validation errors that occur BEFORE any DB calls
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing members_id",
			payload: map[string]interface{}{
				"consent_key": "privacy_policy",
				"action":      "ACCEPTED",
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing consent_key",
			payload: map[string]interface{}{
				"members_id": "550e8400-e29b-41d4-a716-446655440000",
				"action":     "ACCEPTED",
				"timestamp":  time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing action",
			payload: map[string]interface{}{
				"members_id":  "550e8400-e29b-41d4-a716-446655440000",
				"consent_key": "privacy_policy",
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "invalid action value",
			payload: map[string]interface{}{
				"members_id":  "550e8400-e29b-41d4-a716-446655440000",
				"consent_key": "privacy_policy",
				"action":      "INVALID",
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing timestamp",
			payload: map[string]interface{}{
				"members_id":  "550e8400-e29b-41d4-a716-446655440000",
				"consent_key": "privacy_policy",
				"action":      "ACCEPTED",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "invalid members_id format",
			payload: map[string]interface{}{
				"members_id":  "not-a-uuid",
				"consent_key": "privacy_policy",
				"action":      "ACCEPTED",
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid members_id format",
		},
		{
			name:           "empty body",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "invalid timestamp format",
			payload: map[string]interface{}{
				"members_id":  "550e8400-e29b-41d4-a716-446655440000",
				"consent_key": "privacy_policy",
				"action":      "ACCEPTED",
				"timestamp":   "not-a-timestamp",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			// Create handler without DB - tests validation only
			cacheInstance, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())
			handler := &ConsentWebhookHandler{
				DB:    nil,
				Cache: cacheInstance,
			}

			// Create request
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/consent-events", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("api_key_source", "test-source")

			// Call handler
			handler.HandleConsentEvent(c)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			errorMsg, ok := response["error"].(string)
			assert.True(t, ok)
			assert.Contains(t, errorMsg, tt.expectedError)
		})
	}
}

func TestConsentWebhookHandler_MissingAPIKeySource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cacheInstance, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())
	handler := &ConsentWebhookHandler{
		DB:    nil,
		Cache: cacheInstance,
	}

	payload := map[string]interface{}{
		"members_id":  "550e8400-e29b-41d4-a716-446655440000",
		"consent_key": "privacy_policy",
		"action":      "ACCEPTED",
		"timestamp":   time.Now().Format(time.RFC3339),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/consent-events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Note: NOT setting api_key_source in context

	handler.HandleConsentEvent(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response["error"])
}

func TestConsentEventRequest_ActionValues(t *testing.T) {
	// Test that only ACCEPTED and REJECTED are valid
	// Invalid actions should return 400 Bad Request
	invalidActions := []string{
		"accepted", // case sensitive
		"rejected",
		"PENDING",
		"REVOKED",
		"",
	}

	for _, action := range invalidActions {
		name := "action_" + action
		if action == "" {
			name = "action_empty"
		}
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			cacheInstance, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())
			handler := &ConsentWebhookHandler{
				DB:    nil,
				Cache: cacheInstance,
			}

			payload := map[string]interface{}{
				"members_id":  "550e8400-e29b-41d4-a716-446655440000",
				"consent_key": "privacy_policy",
				"action":      action,
				"timestamp":   time.Now().Format(time.RFC3339),
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/consent-events", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("api_key_source", "test-source")

			handler.HandleConsentEvent(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "invalid request body", response["error"])
		})
	}
}
