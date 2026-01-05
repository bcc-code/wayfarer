package ladder_to_heaven

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/plugins"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateDeadlinePoints(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		consumedAt     time.Time
		completeBy     time.Time
		hasDeadline    bool
		expectedPoints int
	}{
		{
			name:           "no deadline set - full points",
			consumedAt:     now,
			completeBy:     time.Time{},
			hasDeadline:    false,
			expectedPoints: pointsOnTime,
		},
		{
			name:           "completed before deadline - full points",
			consumedAt:     now,
			completeBy:     now.Add(24 * time.Hour),
			hasDeadline:    true,
			expectedPoints: pointsOnTime,
		},
		{
			name:           "completed exactly on deadline - full points",
			consumedAt:     now,
			completeBy:     now,
			hasDeadline:    true,
			expectedPoints: pointsOnTime,
		},
		{
			name:           "completed after deadline - reduced points",
			consumedAt:     now,
			completeBy:     now.Add(-24 * time.Hour),
			hasDeadline:    true,
			expectedPoints: pointsLate,
		},
		{
			name:           "completed 1 second after deadline - reduced points",
			consumedAt:     now.Add(1 * time.Second),
			completeBy:     now,
			hasDeadline:    true,
			expectedPoints: pointsLate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := calculateDeadlinePoints(tt.consumedAt, tt.completeBy, tt.hasDeadline)
			assert.Equal(t, tt.expectedPoints, points)
		})
	}
}

func TestContentEventHandler_FeatureDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cacheInstance, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())
	handler := &contentEventHandler{
		db:            nil,
		cache:         cacheInstance,
		achievementID: "", // Feature disabled
	}

	payload := contentEventRequest{
		EventType: "external_content_event",
		Timestamp: time.Now(),
		ProjectID: "PR01234567890123456789012345",
		User: &contentEventUserData{
			ID:        "US01234567890123456789012345",
			MembersID: "550e8400-e29b-41d4-a716-446655440000",
			Email:     "test@example.com",
			Name:      "Test User",
		},
		Data: contentEventData{
			TaskID:          "task-123",
			PlanID:          "plan-456",
			ContentProgress: 1.0,
			ConsumedAt:      time.Now().Format(time.RFC3339),
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/plugins/ladder-to-heaven/content-event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.handle(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestContentEventRequest_ValidationErrors(t *testing.T) {
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
				"user": map[string]interface{}{
					"id": "US01234567890123456789012345",
				},
				"data": map[string]interface{}{
					"task_id":     "task-123",
					"consumed_at": time.Now().Format(time.RFC3339),
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing user",
			payload: map[string]interface{}{
				"event_type": "external_content_event",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"data": map[string]interface{}{
					"task_id":     "task-123",
					"consumed_at": time.Now().Format(time.RFC3339),
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing user.id",
			payload: map[string]interface{}{
				"event_type": "external_content_event",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"user": map[string]interface{}{
					"email": "test@example.com",
				},
				"data": map[string]interface{}{
					"task_id":     "task-123",
					"consumed_at": time.Now().Format(time.RFC3339),
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing data",
			payload: map[string]interface{}{
				"event_type": "external_content_event",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"user": map[string]interface{}{
					"id": "US01234567890123456789012345",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing data.task_id",
			payload: map[string]interface{}{
				"event_type": "external_content_event",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"user": map[string]interface{}{
					"id": "US01234567890123456789012345",
				},
				"data": map[string]interface{}{
					"consumed_at": time.Now().Format(time.RFC3339),
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing data.consumed_at",
			payload: map[string]interface{}{
				"event_type": "external_content_event",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"user": map[string]interface{}{
					"id": "US01234567890123456789012345",
				},
				"data": map[string]interface{}{
					"task_id": "task-123",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "invalid timestamp format",
			payload: map[string]interface{}{
				"event_type": "external_content_event",
				"timestamp":  "not-a-timestamp",
				"project_id": "PR01234567890123456789012345",
				"user": map[string]interface{}{
					"id": "US01234567890123456789012345",
				},
				"data": map[string]interface{}{
					"task_id":     "task-123",
					"consumed_at": time.Now().Format(time.RFC3339),
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
			handler := &contentEventHandler{
				db:            nil,
				cache:         cacheInstance,
				achievementID: "AC01234567890123456789012345",
			}

			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/plugins/ladder-to-heaven/content-event", bytes.NewReader(body))
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

func TestPointConstants(t *testing.T) {
	assert.Equal(t, 50, pointsOnTime, "Points for on-time completion should be 50")
	assert.Equal(t, 25, pointsLate, "Points for late completion should be 25")
}

func TestPlugin_Disabled(t *testing.T) {
	plugin := NewPlugin(Config{
		AchievementID: "", // Disabled
	})

	assert.Equal(t, "Ladder to Heaven", plugin.Name())
	assert.False(t, plugin.Enabled(), "Plugin should be disabled when AchievementID is empty")
}

func TestPlugin_Enabled(t *testing.T) {
	plugin := NewPlugin(Config{
		AchievementID: "AC01234567890123456789012345",
	})

	assert.Equal(t, "Ladder to Heaven", plugin.Name())
	assert.True(t, plugin.Enabled(), "Plugin should be enabled when AchievementID is set")
}

func TestPlugin_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	cacheInstance, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())

	plugin := NewPlugin(Config{
		AchievementID: "AC01234567890123456789012345",
	})

	err := plugin.Register(
		router,
		plugins.Dependencies{
			DB:    nil,
			Cache: cacheInstance,
		},
		func(c *gin.Context) { c.Next() },
	)

	assert.NoError(t, err, "Plugin registration should succeed")
}
