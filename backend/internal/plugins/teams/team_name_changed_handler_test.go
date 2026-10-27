package teams

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

func TestTeamNameChangedHandler_FeatureDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cacheInstance, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())
	handler := &teamNameChangedHandler{
		db:          nil,
		cache:       cacheInstance,
		challengeID: "", // Feature disabled
	}

	payload := teamNameChangedRequest{
		EventType: "team_name_changed",
		Timestamp: time.Now(),
		ProjectID: "PR01234567890123456789012345",
		Data: teamNameChangedData{
			TeamID:  "TM01234567890123456789012345",
			OldName: "Old Team Name",
			NewName: "New Team Name",
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/plugins/teams/team-name-changed", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.handle(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "feature disabled", response["error"])
}

func TestTeamNameChangedRequest_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "empty body",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing event_type",
			payload: map[string]interface{}{
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"data": map[string]interface{}{
					"team_id":  "TM01234567890123456789012345",
					"old_name": "Old Name",
					"new_name": "New Name",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing project_id",
			payload: map[string]interface{}{
				"event_type": "team_name_changed",
				"timestamp":  time.Now().Format(time.RFC3339),
				"data": map[string]interface{}{
					"team_id":  "TM01234567890123456789012345",
					"old_name": "Old Name",
					"new_name": "New Name",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing data",
			payload: map[string]interface{}{
				"event_type": "team_name_changed",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing data.team_id",
			payload: map[string]interface{}{
				"event_type": "team_name_changed",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"data": map[string]interface{}{
					"old_name": "Old Name",
					"new_name": "New Name",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing data.old_name",
			payload: map[string]interface{}{
				"event_type": "team_name_changed",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"data": map[string]interface{}{
					"team_id":  "TM01234567890123456789012345",
					"new_name": "New Name",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing data.new_name",
			payload: map[string]interface{}{
				"event_type": "team_name_changed",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"data": map[string]interface{}{
					"team_id":  "TM01234567890123456789012345",
					"old_name": "Old Name",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "invalid timestamp format",
			payload: map[string]interface{}{
				"event_type": "team_name_changed",
				"timestamp":  "not-a-timestamp",
				"project_id": "PR01234567890123456789012345",
				"data": map[string]interface{}{
					"team_id":  "TM01234567890123456789012345",
					"old_name": "Old Name",
					"new_name": "New Name",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			cacheInstance, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())
			handler := &teamNameChangedHandler{
				db:          nil,
				cache:       cacheInstance,
				challengeID: "CL01234567890123456789012345",
				secretKey:   testSecret,
			}

			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/plugins/teams/team-name-changed", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Webhook-Signature", signBody(body))

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.handle(c)

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

func TestTeamRenamePointConstants(t *testing.T) {
	assert.Equal(t, 300, pointsTeamRename, "Points for team rename should be 300")
	assert.Equal(t, "PLUGIN", sourceTypeTeamRename, "Source type for team rename should be PLUGIN")
}

const testSecret = "test-secret"

// signBody signs body the way the webhook dispatcher does.
func signBody(body []byte) string {
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestTeamNameChangedHandler_Authentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cacheInstance, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())

	payload := func(eventType string) []byte {
		body, err := json.Marshal(map[string]any{
			"event_type": "%s",
			"timestamp":  time.Now().Format(time.RFC3339),
			"project_id": "PR01234567890123456789012345",
			"data":       map[string]any{"team_id": "TM01234567890123456789012345", "old_name": "Old", "new_name": "New"},
		})
		require.NoError(t, err)
		return body
	}

	tests := []struct {
		name           string
		secretKey      string
		body           []byte
		signature      func(body []byte) string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "no secret configured rejects even a signed request",
			secretKey:      "",
			body:           payload("team_name_changed"),
			signature:      signBody,
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "webhook secret not configured",
		},
		{
			name:           "missing signature",
			secretKey:      testSecret,
			body:           payload("team_name_changed"),
			signature:      func([]byte) string { return "" },
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid signature",
		},
		{
			name:           "signature with the wrong secret",
			secretKey:      "another-secret",
			body:           payload("team_name_changed"),
			signature:      signBody,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid signature",
		},
		{
			name:           "signed request for another event type",
			secretKey:      testSecret,
			body:           payload("points_awarded"),
			signature:      signBody,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unexpected event_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretKey := tt.secretKey
			handler := &teamNameChangedHandler{cache: cacheInstance, challengeID: "CL01234567890123456789012345", secretKey: secretKey}

			req := httptest.NewRequest(http.MethodPost, "/plugins/teams/team-name-changed", bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if sig := tt.signature(tt.body); sig != "" {
				req.Header.Set("X-Webhook-Signature", sig)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.handle(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			var response map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			assert.Equal(t, tt.expectedError, response["error"])
		})
	}
}
