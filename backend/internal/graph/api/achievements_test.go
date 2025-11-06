package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// TestBuildAchievementFilterParamsCursor tests the buildAchievementFilterParamsCursor function
func TestBuildAchievementFilterParamsCursor(t *testing.T) {
	tests := []struct {
		name        string
		filter      *model.AchievementFilter
		first       *int
		after       *string
		last        *int
		before      *string
		expectError bool
		errorMsg    string
		check       func(*testing.T, sqlc.GetAchievementsFilteredCursorParams)
	}{
		{
			name: "forward pagination with all filters",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
				EventID:   stringPtr("EV001"),
				Ids:       []string{"AC001", "AC002"},
			},
			first:       intPtr(10),
			after:       stringPtr("AC005"),
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "EV001", params.Eventid)
				assert.Equal(t, []string{"AC001", "AC002"}, params.Ids)
				assert.Equal(t, int32(11), params.Querylimit) // 10 + 1 for hasMore check
				assert.False(t, params.Isbackward)
				assert.Equal(t, "AC005", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "backward pagination with before cursor",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       nil,
			after:       nil,
			last:        intPtr(5),
			before:      stringPtr("AC100"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, int32(6), params.Querylimit) // 5 + 1 for hasMore check
				assert.True(t, params.Isbackward)
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "AC100", params.Beforecursor)
			},
		},
		{
			name:        "both first and last specified - error",
			filter:      &model.AchievementFilter{},
			first:       intPtr(10),
			after:       nil,
			last:        intPtr(5),
			before:      nil,
			expectError: true,
			errorMsg:    "cannot specify both first and last",
		},
		{
			name:        "default pagination - no first or last",
			filter:      &model.AchievementFilter{},
			first:       nil,
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit) // default 10 + 1
				assert.False(t, params.Isbackward)
			},
		},
		{
			name: "only projectId filter",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR999"),
			},
			first:       intPtr(20),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, "PR999", params.Projectid)
				assert.Equal(t, int32(21), params.Querylimit) // 20 + 1
			},
		},
		{
			name: "only eventId filter",
			filter: &model.AchievementFilter{
				EventID: stringPtr("EV123"),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, "EV123", params.Eventid)
			},
		},
		{
			name: "empty cursors",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(10),
			after:       stringPtr(""),
			last:        nil,
			before:      stringPtr(""),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "forward pagination with after and before cursors",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(15),
			after:       stringPtr("AC010"),
			last:        nil,
			before:      stringPtr("AC050"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, int32(16), params.Querylimit) // 15 + 1
				assert.False(t, params.Isbackward)
				assert.Equal(t, "AC010", params.Aftercursor)
				assert.Equal(t, "AC050", params.Beforecursor)
			},
		},
		{
			name: "minimal filter with IDs",
			filter: &model.AchievementFilter{
				Ids: []string{"AC001", "AC002", "AC003"},
			},
			first:       intPtr(3),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, []string{"AC001", "AC002", "AC003"}, params.Ids)
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
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit)
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "projectId and eventId filter combined",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
				EventID:   stringPtr("EV456"),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetAchievementsFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "EV456", params.Eventid)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildAchievementFilterParamsCursor(tt.filter, tt.first, tt.after, tt.last, tt.before)

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

// TestBuildCountAchievementsFilterParams tests the buildCountAchievementsFilterParams function
func TestBuildCountAchievementsFilterParams(t *testing.T) {
	tests := []struct {
		name   string
		filter *model.AchievementFilter
		check  func(*testing.T, sqlc.CountAchievementsFilteredParams)
	}{
		{
			name: "all filters populated",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
				EventID:   stringPtr("EV001"),
				Ids:       []string{"AC001", "AC002"},
			},
			check: func(t *testing.T, params sqlc.CountAchievementsFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "EV001", params.Eventid)
				assert.Equal(t, []string{"AC001", "AC002"}, params.Ids)
			},
		},
		{
			name: "only project filter",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR999"),
			},
			check: func(t *testing.T, params sqlc.CountAchievementsFilteredParams) {
				assert.Equal(t, "PR999", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "only IDs filter",
			filter: &model.AchievementFilter{
				Ids: []string{"AC100", "AC200", "AC300"},
			},
			check: func(t *testing.T, params sqlc.CountAchievementsFilteredParams) {
				assert.Equal(t, []string{"AC100", "AC200", "AC300"}, params.Ids)
				assert.Equal(t, "", params.Projectid)
			},
		},
		{
			name:   "empty filter",
			filter: &model.AchievementFilter{},
			check: func(t *testing.T, params sqlc.CountAchievementsFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Equal(t, "", params.Eventid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "only eventId filter",
			filter: &model.AchievementFilter{
				EventID: stringPtr("EV777"),
			},
			check: func(t *testing.T, params sqlc.CountAchievementsFilteredParams) {
				assert.Equal(t, "EV777", params.Eventid)
			},
		},
		{
			name: "project and event filters combined",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
				EventID:   stringPtr("EV456"),
			},
			check: func(t *testing.T, params sqlc.CountAchievementsFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "EV456", params.Eventid)
			},
		},
		{
			name: "empty IDs array",
			filter: &model.AchievementFilter{
				Ids: []string{},
			},
			check: func(t *testing.T, params sqlc.CountAchievementsFilteredParams) {
				assert.NotNil(t, params.Ids)
				assert.Empty(t, params.Ids)
			},
		},
		{
			name:   "nil filter",
			filter: nil,
			check: func(t *testing.T, params sqlc.CountAchievementsFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Equal(t, "", params.Eventid)
				assert.Nil(t, params.Ids)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCountAchievementsFilterParams(tt.filter)
			tt.check(t, result)
		})
	}
}

// TestBuildAchievementCacheKeyParams tests the buildAchievementCacheKeyParams function
func TestBuildAchievementCacheKeyParams(t *testing.T) {
	tests := []struct {
		name   string
		filter *model.AchievementFilter
		first  *int
		after  *string
		last   *int
		before *string
		check  func(*testing.T, map[string]string)
	}{
		{
			name: "all parameters populated",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
				EventID:   stringPtr("EV001"),
				Ids:       []string{"AC001", "AC002"},
			},
			first:  intPtr(10),
			after:  stringPtr("cursor123"),
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Equal(t, "EV001", params["eventid"])
				assert.Contains(t, params["ids"], "AC001")
				assert.Contains(t, params["ids"], "AC002")
				assert.Equal(t, "10", params["first"])
				assert.Equal(t, "cursor123", params["after"])
				assert.NotContains(t, params, "last")
				assert.NotContains(t, params, "before")
			},
		},
		{
			name: "backward pagination",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:  nil,
			after:  nil,
			last:   intPtr(5),
			before: stringPtr("cursor456"),
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Equal(t, "5", params["last"])
				assert.Equal(t, "cursor456", params["before"])
				assert.NotContains(t, params, "first")
				assert.NotContains(t, params, "after")
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
				assert.Equal(t, "10", params["first"])
				assert.NotContains(t, params, "projectid")
				assert.NotContains(t, params, "eventid")
				assert.NotContains(t, params, "ids")
			},
		},
		{
			name:   "empty filter",
			filter: &model.AchievementFilter{},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "10", params["first"])
				assert.NotContains(t, params, "projectid")
				assert.NotContains(t, params, "eventid")
			},
		},
		{
			name: "empty string cursors ignored",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:  intPtr(10),
			after:  stringPtr(""),
			last:   nil,
			before: stringPtr(""),
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Equal(t, "10", params["first"])
				assert.NotContains(t, params, "after")
				assert.NotContains(t, params, "before")
			},
		},
		{
			name: "empty IDs array ignored",
			filter: &model.AchievementFilter{
				ProjectID: stringPtr("PR123"),
				Ids:       []string{},
			},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.NotContains(t, params, "ids")
			},
		},
		{
			name:   "only pagination params",
			filter: &model.AchievementFilter{},
			first:  intPtr(20),
			after:  stringPtr("aftercursor"),
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "20", params["first"])
				assert.Equal(t, "aftercursor", params["after"])
				assert.Equal(t, 2, len(params))
			},
		},
		{
			name: "eventId filter",
			filter: &model.AchievementFilter{
				EventID: stringPtr("EV123"),
			},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "EV123", params["eventid"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAchievementCacheKeyParams(tt.filter, tt.first, tt.after, tt.last, tt.before)
			tt.check(t, result)
		})
	}
}

func TestBuildAchievementCacheKeyParams_Deterministic(t *testing.T) {
	filter := &model.AchievementFilter{
		ProjectID: stringPtr("PR123"),
		EventID:   stringPtr("EV456"),
	}
	first := intPtr(10)
	after := stringPtr("cursor")

	// Generate params multiple times
	results := make([]map[string]string, 5)
	for i := 0; i < 5; i++ {
		results[i] = buildAchievementCacheKeyParams(filter, first, after, nil, nil)
	}

	// All results should be equal
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i], "buildAchievementCacheKeyParams should be deterministic")
	}
}
