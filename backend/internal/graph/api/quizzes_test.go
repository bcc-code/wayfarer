package api

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
)

// TestBuildQuizFilterParamsCursor tests the buildQuizFilterParamsCursor function
func TestBuildQuizFilterParamsCursor(t *testing.T) {
	publishedAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	publishedBefore := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name        string
		filter      *model.QuizFilter
		first       *int
		after       *string
		last        *int
		before      *string
		expectError bool
		errorMsg    string
		check       func(*testing.T, sqlc.GetQuizzesFilteredCursorParams)
	}{
		{
			name: "forward pagination with all filters",
			filter: &model.QuizFilter{
				ProjectID:       stringPtr("PR123"),
				ChallengeID:     stringPtr("CL001"),
				Ids:             []string{"QZ001", "QZ002"},
				PublishedAfter:  &scalars.DateTime{Time: publishedAfter},
				PublishedBefore: &scalars.DateTime{Time: publishedBefore},
			},
			first:       intPtr(10),
			after:       stringPtr("QZ005"),
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetQuizzesFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "CL001", params.Challengeid)
				assert.Equal(t, []string{"QZ001", "QZ002"}, params.Ids)
				assert.True(t, params.Publishedafter.Valid)
				assert.Equal(t, publishedAfter, params.Publishedafter.Time)
				assert.True(t, params.Publishedbefore.Valid)
				assert.Equal(t, publishedBefore, params.Publishedbefore.Time)
				assert.Equal(t, int32(11), params.Querylimit) // 10 + 1 for hasMore check
				assert.False(t, params.Isbackward)
				assert.Equal(t, "QZ005", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "backward pagination with before cursor",
			filter: &model.QuizFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       nil,
			after:       nil,
			last:        intPtr(5),
			before:      stringPtr("QZ100"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetQuizzesFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, int32(6), params.Querylimit) // 5 + 1 for hasMore check
				assert.True(t, params.Isbackward)
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "QZ100", params.Beforecursor)
			},
		},
		{
			name:        "both first and last specified - error",
			filter:      &model.QuizFilter{},
			first:       intPtr(10),
			after:       nil,
			last:        intPtr(5),
			before:      nil,
			expectError: true,
			errorMsg:    "cannot specify both first and last",
		},
		{
			name:        "default pagination - no first or last",
			filter:      &model.QuizFilter{},
			first:       nil,
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetQuizzesFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit) // default 10 + 1
				assert.False(t, params.Isbackward)
			},
		},
		{
			name: "only projectId filter",
			filter: &model.QuizFilter{
				ProjectID: stringPtr("PR999"),
			},
			first:       intPtr(20),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetQuizzesFilteredCursorParams) {
				assert.Equal(t, "PR999", params.Projectid)
				assert.Equal(t, int32(21), params.Querylimit) // 20 + 1
			},
		},
		{
			name: "only challengeId filter",
			filter: &model.QuizFilter{
				ChallengeID: stringPtr("CL123"),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetQuizzesFilteredCursorParams) {
				assert.Equal(t, "CL123", params.Challengeid)
			},
		},
		{
			name:        "nil filter",
			filter:      nil,
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetQuizzesFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit)
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Publishedafter.Valid)
				assert.False(t, params.Publishedbefore.Valid)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildQuizFilterParamsCursor(tt.filter, tt.first, tt.after, tt.last, tt.before)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}

// TestBuildCountQuizzesFilterParams tests the buildCountQuizzesFilterParams function
func TestBuildCountQuizzesFilterParams(t *testing.T) {
	publishedAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	publishedBefore := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name   string
		filter *model.QuizFilter
		check  func(*testing.T, sqlc.CountQuizzesFilteredParams)
	}{
		{
			name: "all filters populated",
			filter: &model.QuizFilter{
				ProjectID:       stringPtr("PR123"),
				ChallengeID:     stringPtr("CL001"),
				Ids:             []string{"QZ001", "QZ002"},
				PublishedAfter:  &scalars.DateTime{Time: publishedAfter},
				PublishedBefore: &scalars.DateTime{Time: publishedBefore},
			},
			check: func(t *testing.T, params sqlc.CountQuizzesFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "CL001", params.Challengeid)
				assert.Equal(t, []string{"QZ001", "QZ002"}, params.Ids)
				assert.True(t, params.Publishedafter.Valid)
				assert.Equal(t, publishedAfter, params.Publishedafter.Time)
				assert.True(t, params.Publishedbefore.Valid)
				assert.Equal(t, publishedBefore, params.Publishedbefore.Time)
			},
		},
		{
			name: "only project filter",
			filter: &model.QuizFilter{
				ProjectID: stringPtr("PR999"),
			},
			check: func(t *testing.T, params sqlc.CountQuizzesFilteredParams) {
				assert.Equal(t, "PR999", params.Projectid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Publishedafter.Valid)
				assert.False(t, params.Publishedbefore.Valid)
			},
		},
		{
			name:   "empty filter",
			filter: &model.QuizFilter{},
			check: func(t *testing.T, params sqlc.CountQuizzesFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Equal(t, "", params.Challengeid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Publishedafter.Valid)
				assert.False(t, params.Publishedbefore.Valid)
			},
		},
		{
			name:   "nil filter",
			filter: nil,
			check: func(t *testing.T, params sqlc.CountQuizzesFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Equal(t, "", params.Challengeid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Publishedafter.Valid)
				assert.False(t, params.Publishedbefore.Valid)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCountQuizzesFilterParams(tt.filter)

			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// TestBuildQuizCacheKeyParams tests the buildQuizCacheKeyParams function
func TestBuildQuizCacheKeyParams(t *testing.T) {
	publishedAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	publishedBefore := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name   string
		filter *model.QuizFilter
		first  *int
		after  *string
		last   *int
		before *string
		check  func(*testing.T, map[string]string)
	}{
		{
			name: "all parameters populated",
			filter: &model.QuizFilter{
				ProjectID:       stringPtr("PR123"),
				ChallengeID:     stringPtr("CL001"),
				Ids:             []string{"QZ001", "QZ002"},
				PublishedAfter:  &scalars.DateTime{Time: publishedAfter},
				PublishedBefore: &scalars.DateTime{Time: publishedBefore},
			},
			first:  intPtr(10),
			after:  stringPtr("QZ005"),
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Equal(t, "CL001", params["challengeid"])
				assert.Equal(t, "[QZ001 QZ002]", params["ids"])
				assert.Equal(t, publishedAfter.Format(time.RFC3339), params["publishedafter"])
				assert.Equal(t, publishedBefore.Format(time.RFC3339), params["publishedbefore"])
				assert.Equal(t, "10", params["first"])
				assert.Equal(t, "QZ005", params["after"])
				_, hasLast := params["last"]
				assert.False(t, hasLast)
				_, hasBefore := params["before"]
				assert.False(t, hasBefore)
			},
		},
		{
			name: "backward pagination",
			filter: &model.QuizFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:  nil,
			after:  nil,
			last:   intPtr(5),
			before: stringPtr("QZ100"),
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Equal(t, "5", params["last"])
				assert.Equal(t, "QZ100", params["before"])
				_, hasFirst := params["first"]
				assert.False(t, hasFirst)
				_, hasAfter := params["after"]
				assert.False(t, hasAfter)
			},
		},
		{
			name:   "nil filter",
			filter: nil,
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, 1, len(params))
				assert.Equal(t, "10", params["first"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildQuizCacheKeyParams(tt.filter, tt.first, tt.after, tt.last, tt.before)

			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// NOTE: Unit tests for quiz mutation resolvers would require a proper mocking infrastructure
// for the Resolver type and its dependencies. The codebase currently only has unit tests for
// helper functions. The mutation resolvers should be tested through integration tests or
// manual testing with a real database.

// TestConvertQuestionPoints tests the convertQuestionPoints helper function
func TestConvertQuestionPoints(t *testing.T) {
	tests := []struct {
		name     string
		input    *int32
		expected *int
	}{
		{
			name:     "nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "zero points",
			input:    int32Ptr(0),
			expected: intPtr(0),
		},
		{
			name:     "positive points",
			input:    int32Ptr(10),
			expected: intPtr(10),
		},
		{
			name:     "large points value",
			input:    int32Ptr(1000),
			expected: intPtr(1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertQuestionPoints(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

// TestConvertToPredefinedQuestionWithPoints tests that points are correctly included in question conversion
func TestConvertToPredefinedQuestionWithPoints(t *testing.T) {
	tests := []struct {
		name           string
		points         *int32
		expectedPoints *int
	}{
		{
			name:           "nil points",
			points:         nil,
			expectedPoints: nil,
		},
		{
			name:           "with points value",
			points:         int32Ptr(5),
			expectedPoints: intPtr(5),
		},
		{
			name:           "zero points",
			points:         int32Ptr(0),
			expectedPoints: intPtr(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := &mockQuestionRow{
				id:            "QQ001",
				quizID:        "QZ001",
				questionType:  "PREDEFINED",
				questionText:  "Test question?",
				questionOrder: 1,
				points:        tt.points,
			}

			result := convertToPredefinedQuestion(row)
			require.NotNil(t, result)
			assert.Equal(t, "QQ001", result.ID)

			if tt.expectedPoints == nil {
				assert.Nil(t, result.Points)
			} else {
				require.NotNil(t, result.Points)
				assert.Equal(t, *tt.expectedPoints, *result.Points)
			}
		})
	}
}

// TestConvertResponseRowWithPointsEarned tests that pointsEarned is correctly included in response conversion
func TestConvertResponseRowWithPointsEarned(t *testing.T) {
	tests := []struct {
		name                 string
		pointsEarned         *int32
		expectedPointsEarned *int
	}{
		{
			name:                 "nil points earned",
			pointsEarned:         nil,
			expectedPointsEarned: nil,
		},
		{
			name:                 "with points earned",
			pointsEarned:         int32Ptr(10),
			expectedPointsEarned: intPtr(10),
		},
		{
			name:                 "zero points earned",
			pointsEarned:         int32Ptr(0),
			expectedPointsEarned: intPtr(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := &sqlc.QuizResponse{
				ID:           "QR001",
				SubmissionID: "QS001",
				QuestionID:   "QQ001",
				PointsEarned: tt.pointsEarned,
			}

			result := convertResponseRowToInterface(row, "PREDEFINED")
			require.NotNil(t, result)

			predefinedResponse, ok := result.(*model.PredefinedResponse)
			require.True(t, ok)
			assert.Equal(t, "QR001", predefinedResponse.ID)

			if tt.expectedPointsEarned == nil {
				assert.Nil(t, predefinedResponse.PointsEarned)
			} else {
				require.NotNil(t, predefinedResponse.PointsEarned)
				assert.Equal(t, *tt.expectedPointsEarned, *predefinedResponse.PointsEarned)
			}
		})
	}
}

// mockQuestionRow implements quizQuestionRow for testing
type mockQuestionRow struct {
	id                     string
	quizID                 string
	questionType           string
	questionText           string
	questionOrder          int32
	allowMultipleSelection *bool
	timeoutSeconds         *int32
	points                 *int32
}

func (m *mockQuestionRow) GetID() string                   { return m.id }
func (m *mockQuestionRow) GetQuizID() string               { return m.quizID }
func (m *mockQuestionRow) GetQuestionType() string         { return m.questionType }
func (m *mockQuestionRow) GetQuestionText() string         { return m.questionText }
func (m *mockQuestionRow) GetQuestionOrder() int32         { return m.questionOrder }
func (m *mockQuestionRow) GetAllowMultipleSelection() *bool { return m.allowMultipleSelection }
func (m *mockQuestionRow) GetMinValue() pgtype.Numeric     { return pgtype.Numeric{} }
func (m *mockQuestionRow) GetMaxValue() pgtype.Numeric     { return pgtype.Numeric{} }
func (m *mockQuestionRow) GetStepValue() pgtype.Numeric    { return pgtype.Numeric{} }
func (m *mockQuestionRow) GetTimeoutSeconds() *int32       { return m.timeoutSeconds }
func (m *mockQuestionRow) GetPoints() *int32               { return m.points }

// Helper functions for tests
func int32Ptr(v int32) *int32 {
	return &v
}
