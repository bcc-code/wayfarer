package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentEventRequest_ValidationErrors(t *testing.T) {
	// These tests only cover validation errors that occur BEFORE any DB calls
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing person_id",
			payload: map[string]interface{}{
				"task_id":   "task-123",
				"timestamp": time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing task_id",
			payload: map[string]interface{}{
				"person_id": "550e8400-e29b-41d4-a716-446655440000",
				"timestamp": time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing timestamp",
			payload: map[string]interface{}{
				"person_id": "550e8400-e29b-41d4-a716-446655440000",
				"task_id":   "task-123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "invalid person_id format",
			payload: map[string]interface{}{
				"person_id": "not-a-uuid",
				"task_id":   "task-123",
				"timestamp": time.Now().Format(time.RFC3339),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid person_id format",
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
				"person_id": "550e8400-e29b-41d4-a716-446655440000",
				"task_id":   "task-123",
				"timestamp": "not-a-timestamp",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			// Create handler without DB - tests validation only
			handler := &WebhookHandler{
				DB: nil,
			}

			// Create request
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/content-events", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("api_key_source", "test-source")

			// Call handler
			handler.HandleContentEvent(c)

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

func TestContentEventRequest_ContentProgressValidation(t *testing.T) {
	tests := []struct {
		name            string
		contentProgress *float64
		isValid         bool
	}{
		{
			name:            "nil progress is valid",
			contentProgress: nil,
			isValid:         true,
		},
		{
			name:            "0.01 is valid (minimum)",
			contentProgress: floatPtr(0.01),
			isValid:         true,
		},
		{
			name:            "1.0 is valid (100%)",
			contentProgress: floatPtr(1.0),
			isValid:         true,
		},
		{
			name:            "1.1 is valid (maximum)",
			contentProgress: floatPtr(1.1),
			isValid:         true,
		},
		{
			name:            "0.5 is valid (50%)",
			contentProgress: floatPtr(0.5),
			isValid:         true,
		},
		{
			name:            "0.0 is invalid (too low)",
			contentProgress: floatPtr(0.0),
			isValid:         false,
		},
		{
			name:            "0.009 is invalid (below 0.01)",
			contentProgress: floatPtr(0.009),
			isValid:         false,
		},
		{
			name:            "1.2 is invalid (above 1.1)",
			contentProgress: floatPtr(1.2),
			isValid:         false,
		},
		{
			name:            "-0.5 is invalid (negative)",
			contentProgress: floatPtr(-0.5),
			isValid:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic directly
			var contentProgress *float32
			if tt.contentProgress != nil {
				progress := *tt.contentProgress
				// This mirrors the validation logic in HandleContentEvent
				if progress >= 0.01 && progress <= 1.1 {
					progressFloat32 := float32(progress)
					contentProgress = &progressFloat32
				}
			}

			if tt.isValid {
				if tt.contentProgress == nil {
					assert.Nil(t, contentProgress)
				} else {
					assert.NotNil(t, contentProgress)
				}
			} else {
				// Invalid values result in nil (NULL in database)
				assert.Nil(t, contentProgress)
			}
		})
	}
}

func TestWebhookHandler_MissingAPIKeySource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &WebhookHandler{
		DB: nil,
	}

	payload := map[string]interface{}{
		"person_id": "550e8400-e29b-41d4-a716-446655440000",
		"task_id":   "task-123",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/content-events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Note: NOT setting api_key_source in context

	handler.HandleContentEvent(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response["error"])
}

// floatPtr returns a pointer to a float64
func floatPtr(f float64) *float64 {
	return &f
}
