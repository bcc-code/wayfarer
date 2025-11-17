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

// TestBuildChallengeFilterParamsCursor tests the buildChallengeFilterParamsCursor function
func TestBuildChallengeFilterParamsCursor(t *testing.T) {
	publishedAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	publishedBefore := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name        string
		filter      *model.ChallengeFilter
		first       *int
		after       *string
		last        *int
		before      *string
		expectError bool
		errorMsg    string
		check       func(*testing.T, sqlc.GetChallengesFilteredCursorParams)
	}{
		{
			name: "forward pagination with all filters",
			filter: &model.ChallengeFilter{
				ProjectID:       stringPtr("PR123"),
				EventID:         stringPtr("EV001"),
				Ids:             []string{"CL001", "CL002"},
				PublishedAfter:  &scalars.DateTime{Time: publishedAfter},
				PublishedBefore: &scalars.DateTime{Time: publishedBefore},
			},
			first:       intPtr(10),
			after:       stringPtr("CL005"),
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "EV001", params.Eventid)
				assert.Equal(t, []string{"CL001", "CL002"}, params.Ids)
				assert.True(t, params.Publishedafter.Valid)
				assert.Equal(t, publishedAfter, params.Publishedafter.Time)
				assert.True(t, params.Publishedbefore.Valid)
				assert.Equal(t, publishedBefore, params.Publishedbefore.Time)
				assert.Equal(t, int32(11), params.Querylimit) // 10 + 1 for hasMore check
				assert.False(t, params.Isbackward)
				assert.Equal(t, "CL005", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "backward pagination with before cursor",
			filter: &model.ChallengeFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       nil,
			after:       nil,
			last:        intPtr(5),
			before:      stringPtr("CL100"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, int32(6), params.Querylimit) // 5 + 1 for hasMore check
				assert.True(t, params.Isbackward)
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "CL100", params.Beforecursor)
			},
		},
		{
			name:        "both first and last specified - error",
			filter:      &model.ChallengeFilter{},
			first:       intPtr(10),
			after:       nil,
			last:        intPtr(5),
			before:      nil,
			expectError: true,
			errorMsg:    "cannot specify both first and last",
		},
		{
			name:        "default pagination - no first or last",
			filter:      &model.ChallengeFilter{},
			first:       nil,
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit) // default 10 + 1
				assert.False(t, params.Isbackward)
			},
		},
		{
			name: "only projectId filter",
			filter: &model.ChallengeFilter{
				ProjectID: stringPtr("PR999"),
			},
			first:       intPtr(20),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, "PR999", params.Projectid)
				assert.Equal(t, int32(21), params.Querylimit) // 20 + 1
			},
		},
		{
			name: "only eventId filter",
			filter: &model.ChallengeFilter{
				EventID: stringPtr("EV123"),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, "EV123", params.Eventid)
			},
		},
		{
			name: "empty cursors",
			filter: &model.ChallengeFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(10),
			after:       stringPtr(""),
			last:        nil,
			before:      stringPtr(""),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "forward pagination with after and before cursors",
			filter: &model.ChallengeFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(15),
			after:       stringPtr("CL010"),
			last:        nil,
			before:      stringPtr("CL050"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, int32(16), params.Querylimit) // 15 + 1
				assert.False(t, params.Isbackward)
				assert.Equal(t, "CL010", params.Aftercursor)
				assert.Equal(t, "CL050", params.Beforecursor)
			},
		},
		{
			name: "minimal filter with IDs",
			filter: &model.ChallengeFilter{
				Ids: []string{"CL001", "CL002", "CL003"},
			},
			first:       intPtr(3),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, []string{"CL001", "CL002", "CL003"}, params.Ids)
				assert.Equal(t, int32(4), params.Querylimit) // 3 + 1
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
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit)
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Publishedafter.Valid)
				assert.False(t, params.Publishedbefore.Valid)
			},
		},
		{
			name: "projectId and eventId filter combined",
			filter: &model.ChallengeFilter{
				ProjectID: stringPtr("PR123"),
				EventID:   stringPtr("EV456"),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "EV456", params.Eventid)
			},
		},
		{
			name: "only publishedAfter filter",
			filter: &model.ChallengeFilter{
				PublishedAfter: &scalars.DateTime{Time: publishedAfter},
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.True(t, params.Publishedafter.Valid)
				assert.Equal(t, publishedAfter, params.Publishedafter.Time)
				assert.False(t, params.Publishedbefore.Valid)
			},
		},
		{
			name: "only publishedBefore filter",
			filter: &model.ChallengeFilter{
				PublishedBefore: &scalars.DateTime{Time: publishedBefore},
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetChallengesFilteredCursorParams) {
				assert.False(t, params.Publishedafter.Valid)
				assert.True(t, params.Publishedbefore.Valid)
				assert.Equal(t, publishedBefore, params.Publishedbefore.Time)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildChallengeFilterParamsCursor(tt.filter, tt.first, tt.after, tt.last, tt.before)

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

// TestBuildCountChallengesFilterParams tests the buildCountChallengesFilterParams function
func TestBuildCountChallengesFilterParams(t *testing.T) {
	publishedAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	publishedBefore := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name   string
		filter *model.ChallengeFilter
		check  func(*testing.T, sqlc.CountChallengesFilteredParams)
	}{
		{
			name: "all filters populated",
			filter: &model.ChallengeFilter{
				ProjectID:       stringPtr("PR123"),
				EventID:         stringPtr("EV001"),
				Ids:             []string{"CL001", "CL002"},
				PublishedAfter:  &scalars.DateTime{Time: publishedAfter},
				PublishedBefore: &scalars.DateTime{Time: publishedBefore},
			},
			check: func(t *testing.T, params sqlc.CountChallengesFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "EV001", params.Eventid)
				assert.Equal(t, []string{"CL001", "CL002"}, params.Ids)
				assert.True(t, params.Publishedafter.Valid)
				assert.Equal(t, publishedAfter, params.Publishedafter.Time)
				assert.True(t, params.Publishedbefore.Valid)
				assert.Equal(t, publishedBefore, params.Publishedbefore.Time)
			},
		},
		{
			name: "only project filter",
			filter: &model.ChallengeFilter{
				ProjectID: stringPtr("PR999"),
			},
			check: func(t *testing.T, params sqlc.CountChallengesFilteredParams) {
				assert.Equal(t, "PR999", params.Projectid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Publishedafter.Valid)
				assert.False(t, params.Publishedbefore.Valid)
			},
		},
		{
			name: "only IDs filter",
			filter: &model.ChallengeFilter{
				Ids: []string{"CL100", "CL200", "CL300"},
			},
			check: func(t *testing.T, params sqlc.CountChallengesFilteredParams) {
				assert.Equal(t, []string{"CL100", "CL200", "CL300"}, params.Ids)
				assert.Equal(t, "", params.Projectid)
			},
		},
		{
			name:   "empty filter",
			filter: &model.ChallengeFilter{},
			check: func(t *testing.T, params sqlc.CountChallengesFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Equal(t, "", params.Eventid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Publishedafter.Valid)
				assert.False(t, params.Publishedbefore.Valid)
			},
		},
		{
			name: "only eventId filter",
			filter: &model.ChallengeFilter{
				EventID: stringPtr("EV777"),
			},
			check: func(t *testing.T, params sqlc.CountChallengesFilteredParams) {
				assert.Equal(t, "EV777", params.Eventid)
			},
		},
		{
			name: "project and event filters combined",
			filter: &model.ChallengeFilter{
				ProjectID: stringPtr("PR123"),
				EventID:   stringPtr("EV456"),
			},
			check: func(t *testing.T, params sqlc.CountChallengesFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "EV456", params.Eventid)
			},
		},
		{
			name: "empty IDs array",
			filter: &model.ChallengeFilter{
				Ids: []string{},
			},
			check: func(t *testing.T, params sqlc.CountChallengesFilteredParams) {
				assert.NotNil(t, params.Ids)
				assert.Empty(t, params.Ids)
			},
		},
		{
			name:   "nil filter",
			filter: nil,
			check: func(t *testing.T, params sqlc.CountChallengesFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Equal(t, "", params.Eventid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Publishedafter.Valid)
				assert.False(t, params.Publishedbefore.Valid)
			},
		},
		{
			name: "date filters only",
			filter: &model.ChallengeFilter{
				PublishedAfter:  &scalars.DateTime{Time: publishedAfter},
				PublishedBefore: &scalars.DateTime{Time: publishedBefore},
			},
			check: func(t *testing.T, params sqlc.CountChallengesFilteredParams) {
				assert.True(t, params.Publishedafter.Valid)
				assert.Equal(t, publishedAfter, params.Publishedafter.Time)
				assert.True(t, params.Publishedbefore.Valid)
				assert.Equal(t, publishedBefore, params.Publishedbefore.Time)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCountChallengesFilterParams(tt.filter)

			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// TestBuildChallengeCacheKeyParams tests the buildChallengeCacheKeyParams function
func TestBuildChallengeCacheKeyParams(t *testing.T) {
	publishedAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	publishedBefore := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name   string
		filter *model.ChallengeFilter
		first  *int
		after  *string
		last   *int
		before *string
		check  func(*testing.T, map[string]string)
	}{
		{
			name: "all parameters populated",
			filter: &model.ChallengeFilter{
				ProjectID:       stringPtr("PR123"),
				EventID:         stringPtr("EV001"),
				Ids:             []string{"CL001", "CL002"},
				PublishedAfter:  &scalars.DateTime{Time: publishedAfter},
				PublishedBefore: &scalars.DateTime{Time: publishedBefore},
			},
			first:  intPtr(10),
			after:  stringPtr("CL005"),
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Equal(t, "EV001", params["eventid"])
				assert.Equal(t, "[CL001 CL002]", params["ids"])
				assert.Equal(t, publishedAfter.Format(time.RFC3339), params["publishedafter"])
				assert.Equal(t, publishedBefore.Format(time.RFC3339), params["publishedbefore"])
				assert.Equal(t, "10", params["first"])
				assert.Equal(t, "CL005", params["after"])
				_, hasLast := params["last"]
				assert.False(t, hasLast)
				_, hasBefore := params["before"]
				assert.False(t, hasBefore)
			},
		},
		{
			name: "backward pagination",
			filter: &model.ChallengeFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:  nil,
			after:  nil,
			last:   intPtr(5),
			before: stringPtr("CL100"),
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Equal(t, "5", params["last"])
				assert.Equal(t, "CL100", params["before"])
				_, hasFirst := params["first"]
				assert.False(t, hasFirst)
				_, hasAfter := params["after"]
				assert.False(t, hasAfter)
			},
		},
		{
			name:   "empty filter",
			filter: &model.ChallengeFilter{},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "10", params["first"])
				_, hasProjectID := params["projectid"]
				assert.False(t, hasProjectID)
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
		{
			name: "empty cursors not included",
			filter: &model.ChallengeFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:  intPtr(10),
			after:  stringPtr(""),
			last:   nil,
			before: stringPtr(""),
			check: func(t *testing.T, params map[string]string) {
				_, hasAfter := params["after"]
				assert.False(t, hasAfter)
				_, hasBefore := params["before"]
				assert.False(t, hasBefore)
			},
		},
		{
			name: "date filters included",
			filter: &model.ChallengeFilter{
				PublishedAfter:  &scalars.DateTime{Time: publishedAfter},
				PublishedBefore: &scalars.DateTime{Time: publishedBefore},
			},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, publishedAfter.Format(time.RFC3339), params["publishedafter"])
				assert.Equal(t, publishedBefore.Format(time.RFC3339), params["publishedbefore"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildChallengeCacheKeyParams(tt.filter, tt.first, tt.after, tt.last, tt.before)

			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// NOTE: Unit tests for challenge mutation resolvers would require a proper mocking infrastructure
// for the Resolver type and its dependencies. The codebase currently only has unit tests for
// helper functions. The mutation resolvers should be tested through integration tests or
// manual testing with a real database.
