package ladder_to_heaven

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountCorrectPositions(t *testing.T) {
	tests := []struct {
		name      string
		submitted []string
		correct   []*model.QuizPredefinedAnswer
		expected  int
	}{
		{
			name:      "all correct - 4 items",
			submitted: []string{"A", "B", "C", "D"},
			correct: []*model.QuizPredefinedAnswer{
				{ID: "A", AnswerOrder: 0},
				{ID: "B", AnswerOrder: 1},
				{ID: "C", AnswerOrder: 2},
				{ID: "D", AnswerOrder: 3},
			},
			expected: 4,
		},
		{
			name:      "none correct - all wrong",
			submitted: []string{"D", "C", "B", "A"},
			correct: []*model.QuizPredefinedAnswer{
				{ID: "A", AnswerOrder: 0},
				{ID: "B", AnswerOrder: 1},
				{ID: "C", AnswerOrder: 2},
				{ID: "D", AnswerOrder: 3},
			},
			expected: 0,
		},
		{
			name:      "two correct - first and last",
			submitted: []string{"A", "C", "B", "D"},
			correct: []*model.QuizPredefinedAnswer{
				{ID: "A", AnswerOrder: 0},
				{ID: "B", AnswerOrder: 1},
				{ID: "C", AnswerOrder: 2},
				{ID: "D", AnswerOrder: 3},
			},
			expected: 2,
		},
		{
			name:      "one correct - first only",
			submitted: []string{"A", "D", "B", "C"},
			correct: []*model.QuizPredefinedAnswer{
				{ID: "A", AnswerOrder: 0},
				{ID: "B", AnswerOrder: 1},
				{ID: "C", AnswerOrder: 2},
				{ID: "D", AnswerOrder: 3},
			},
			expected: 1,
		},
		{
			name:      "empty submitted",
			submitted: []string{},
			correct: []*model.QuizPredefinedAnswer{
				{ID: "A", AnswerOrder: 0},
				{ID: "B", AnswerOrder: 1},
			},
			expected: 0,
		},
		{
			name:      "empty correct",
			submitted: []string{"A", "B"},
			correct:   []*model.QuizPredefinedAnswer{},
			expected:  0,
		},
		{
			name:      "mismatched lengths - more submitted",
			submitted: []string{"A", "B", "C", "D", "E"},
			correct: []*model.QuizPredefinedAnswer{
				{ID: "A", AnswerOrder: 0},
				{ID: "B", AnswerOrder: 1},
				{ID: "C", AnswerOrder: 2},
				{ID: "D", AnswerOrder: 3},
			},
			expected: 4,
		},
		{
			name:      "mismatched lengths - more correct",
			submitted: []string{"A", "B"},
			correct: []*model.QuizPredefinedAnswer{
				{ID: "A", AnswerOrder: 0},
				{ID: "B", AnswerOrder: 1},
				{ID: "C", AnswerOrder: 2},
				{ID: "D", AnswerOrder: 3},
			},
			expected: 2,
		},
		{
			name:      "three items - two correct",
			submitted: []string{"A", "B", "X"},
			correct: []*model.QuizPredefinedAnswer{
				{ID: "A", AnswerOrder: 0},
				{ID: "B", AnswerOrder: 1},
				{ID: "C", AnswerOrder: 2},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countCorrectPositions(tt.submitted, tt.correct)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateBetMultiplier(t *testing.T) {
	tests := []struct {
		name         string
		correctCount int
		totalItems   int
		expected     float64
	}{
		{
			name:         "all correct - 4 items",
			correctCount: 4,
			totalItems:   4,
			expected:     2.0,
		},
		{
			name:         "all correct - 3 items",
			correctCount: 3,
			totalItems:   3,
			expected:     2.0,
		},
		{
			name:         "all correct - 2 items",
			correctCount: 2,
			totalItems:   2,
			expected:     2.0,
		},
		{
			name:         "all correct - 1 item",
			correctCount: 1,
			totalItems:   1,
			expected:     2.0,
		},
		{
			name:         "two correct out of four",
			correctCount: 2,
			totalItems:   4,
			expected:     1.5,
		},
		{
			name:         "two correct out of three",
			correctCount: 2,
			totalItems:   3,
			expected:     1.5,
		},
		{
			name:         "one correct out of four",
			correctCount: 1,
			totalItems:   4,
			expected:     1.2,
		},
		{
			name:         "one correct out of three",
			correctCount: 1,
			totalItems:   3,
			expected:     1.2,
		},
		{
			name:         "none correct out of four - penalty",
			correctCount: 0,
			totalItems:   4,
			expected:     -1.0,
		},
		{
			name:         "none correct out of three - no penalty (not 4 items)",
			correctCount: 0,
			totalItems:   3,
			expected:     0,
		},
		{
			name:         "three correct out of four - not all, not two, not one",
			correctCount: 3,
			totalItems:   4,
			expected:     0,
		},
		{
			name:         "empty items",
			correctCount: 0,
			totalItems:   0,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBetMultiplier(tt.correctCount, tt.totalItems)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseOrderingResponse(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    []byte
		expected    []string
		expectError bool
	}{
		{
			name:        "valid array",
			jsonData:    []byte(`["A", "B", "C", "D"]`),
			expected:    []string{"A", "B", "C", "D"},
			expectError: false,
		},
		{
			name:        "empty array",
			jsonData:    []byte(`[]`),
			expected:    []string{},
			expectError: false,
		},
		{
			name:        "nil input",
			jsonData:    nil,
			expected:    nil,
			expectError: false,
		},
		{
			name:        "empty input",
			jsonData:    []byte{},
			expected:    nil,
			expectError: false,
		},
		{
			name:        "invalid json",
			jsonData:    []byte(`not json`),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "json object instead of array",
			jsonData:    []byte(`{"a": 1}`),
			expected:    nil,
			expectError: true,
		},
		{
			name:        "single item",
			jsonData:    []byte(`["ONLY"]`),
			expected:    []string{"ONLY"},
			expectError: false,
		},
		{
			name:        "ulid-style ids",
			jsonData:    []byte(`["QA01ARZ3NDEKTSV4RRFFQ69G5FAV", "QA01ARZ3NDEKTSV4RRFFQ69G5FAW"]`),
			expected:    []string{"QA01ARZ3NDEKTSV4RRFFQ69G5FAV", "QA01ARZ3NDEKTSV4RRFFQ69G5FAW"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseOrderingResponse(tt.jsonData)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSessionFinishedHandler_InvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cacheInstance, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())
	handler := &quizFinalizedHandler{
		db:        nil,
		cache:     cacheInstance,
		secretKey: "secret-key",
	}

	payload := sessionFinishedRequest{
		EventType: "quiz_session_finished",
		Timestamp: time.Now(),
		ProjectID: "PR01234567890123456789012345",
		Data: sessionFinishedData{
			SessionID: "SS01234567890123456789012345",
			QuizID:    "QZ01234567890123456789012345",
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/plugins/ladder-to-heaven/quiz-finalized", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "invalid-signature")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.handle(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid signature", response["error"])
}

func TestSessionFinishedRequest_ValidationErrors(t *testing.T) {
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
					"session_id": "SS01234567890123456789012345",
					"quiz_id":    "QZ01234567890123456789012345",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing data.session_id",
			payload: map[string]interface{}{
				"event_type": "quiz_session_finished",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"data": map[string]interface{}{
					"quiz_id": "QZ01234567890123456789012345",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "missing data.quiz_id",
			payload: map[string]interface{}{
				"event_type": "quiz_session_finished",
				"timestamp":  time.Now().Format(time.RFC3339),
				"project_id": "PR01234567890123456789012345",
				"data": map[string]interface{}{
					"session_id": "SS01234567890123456789012345",
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
			handler := &quizFinalizedHandler{
				db:        nil,
				cache:     cacheInstance,
				secretKey: "",
			}

			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/plugins/ladder-to-heaven/quiz-finalized", bytes.NewReader(body))
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

func TestBettingConstants(t *testing.T) {
	assert.Equal(t, 2.0, multiplierAllCorrect, "Multiplier for all correct should be 2.0")
	assert.Equal(t, 1.5, multiplierTwoCorrect, "Multiplier for two correct should be 1.5")
	assert.Equal(t, 1.2, multiplierOneCorrect, "Multiplier for one correct should be 1.2")
	assert.Equal(t, -1.0, multiplierAllWrong, "Multiplier for all wrong should be -1.0 (penalty)")
}

func TestBettingPointsCalculation(t *testing.T) {
	tests := []struct {
		name         string
		betAmount    int
		correctCount int
		totalItems   int
		expected     int
	}{
		{
			name:         "100 bet, all correct (4 items) = 200 points",
			betAmount:    100,
			correctCount: 4,
			totalItems:   4,
			expected:     200,
		},
		{
			name:         "100 bet, two correct = 150 points",
			betAmount:    100,
			correctCount: 2,
			totalItems:   4,
			expected:     150,
		},
		{
			name:         "100 bet, one correct = 120 points",
			betAmount:    100,
			correctCount: 1,
			totalItems:   4,
			expected:     120,
		},
		{
			name:         "100 bet, none correct (4 items) = -100 points (penalty)",
			betAmount:    100,
			correctCount: 0,
			totalItems:   4,
			expected:     -100,
		},
		{
			name:         "50 bet, all correct = 100 points",
			betAmount:    50,
			correctCount: 3,
			totalItems:   3,
			expected:     100,
		},
		{
			name:         "75 bet, two correct = 112 points (truncated)",
			betAmount:    75,
			correctCount: 2,
			totalItems:   4,
			expected:     112, // 75 * 1.5 = 112.5, truncated to 112
		},
		{
			name:         "100 bet, three correct out of four = 0 points (no bonus)",
			betAmount:    100,
			correctCount: 3,
			totalItems:   4,
			expected:     0,
		},
		{
			name:         "100 bet, none correct out of 3 items = 0 points (no penalty for non-4 items)",
			betAmount:    100,
			correctCount: 0,
			totalItems:   3,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multiplier := calculateBetMultiplier(tt.correctCount, tt.totalItems)
			bonusPoints := int(float64(tt.betAmount) * multiplier)
			assert.Equal(t, tt.expected, bonusPoints)
		})
	}
}

// Helper to create a QuizPredefinedAnswer for testing
func createTestAnswer(id string, order int) *model.QuizPredefinedAnswer {
	return &model.QuizPredefinedAnswer{
		ID:          id,
		QuestionID:  "QQ01234567890123456789012345",
		AnswerText:  "Answer " + id,
		AnswerOrder: order,
	}
}

func TestCountCorrectPositions_WithRealAnswerStructure(t *testing.T) {
	// Test with actual QuizPredefinedAnswer structure
	correct := []*model.QuizPredefinedAnswer{
		createTestAnswer("QA01ARZ3NDEKTSV4RRFFQ69G5FA1", 0),
		createTestAnswer("QA01ARZ3NDEKTSV4RRFFQ69G5FA2", 1),
		createTestAnswer("QA01ARZ3NDEKTSV4RRFFQ69G5FA3", 2),
		createTestAnswer("QA01ARZ3NDEKTSV4RRFFQ69G5FA4", 3),
	}

	submitted := []string{
		"QA01ARZ3NDEKTSV4RRFFQ69G5FA1",
		"QA01ARZ3NDEKTSV4RRFFQ69G5FA3", // Wrong position
		"QA01ARZ3NDEKTSV4RRFFQ69G5FA2", // Wrong position
		"QA01ARZ3NDEKTSV4RRFFQ69G5FA4",
	}

	result := countCorrectPositions(submitted, correct)
	assert.Equal(t, 2, result, "Should count 2 correct positions (first and last)")
}

func TestPenaltyOnlyAppliesTo4Items(t *testing.T) {
	tests := []struct {
		name        string
		totalItems  int
		wantPenalty bool
	}{
		{"3 items - no penalty", 3, false},
		{"4 items - penalty", 4, true},
		{"5 items - no penalty", 5, false},
		{"2 items - no penalty", 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multiplier := calculateBetMultiplier(0, tt.totalItems)
			if tt.wantPenalty {
				assert.Equal(t, -1.0, multiplier, "Should apply penalty for 4 items")
			} else {
				assert.Equal(t, 0.0, multiplier, "Should not apply penalty for non-4 items")
			}
		})
	}
}
