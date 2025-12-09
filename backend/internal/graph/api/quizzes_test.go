package api

import (
	"testing"
	"time"

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
