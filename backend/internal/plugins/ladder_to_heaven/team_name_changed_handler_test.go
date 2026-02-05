package ladder_to_heaven

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

	req := httptest.NewRequest(http.MethodPost, "/plugins/ladder-to-heaven/team-name-changed", bytes.NewReader(body))
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
			}

			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/plugins/ladder-to-heaven/team-name-changed", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

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
