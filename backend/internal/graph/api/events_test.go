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

// TestBuildEventFilterParamsCursor tests the buildEventFilterParamsCursor function
func TestBuildEventFilterParamsCursor(t *testing.T) {
	// Test dates
	startDate := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 10, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name        string
		filter      *model.EventFilter
		first       *int
		after       *string
		last        *int
		before      *string
		expectError bool
		errorMsg    string
		check       func(*testing.T, sqlc.GetEventsFilteredCursorParams)
	}{
		{
			name: "forward pagination with all filters",
			filter: &model.EventFilter{
				ProjectID:       stringPtr("PR123"),
				Ids:             []string{"EV001", "EV002"},
				StartDateAfter:  &scalars.DateTime{Time: startDate},
				StartDateBefore: &scalars.DateTime{Time: endDate},
				EndDateAfter:    &scalars.DateTime{Time: startDate},
				EndDateBefore:   &scalars.DateTime{Time: endDate},
			},
			first:       intPtr(10),
			after:       stringPtr("EV005"),
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, []string{"EV001", "EV002"}, params.Ids)
				assert.True(t, params.Startdateafter.Valid)
				assert.Equal(t, startDate, params.Startdateafter.Time)
				assert.True(t, params.Startdatebefore.Valid)
				assert.Equal(t, endDate, params.Startdatebefore.Time)
				assert.True(t, params.Enddateafter.Valid)
				assert.Equal(t, startDate, params.Enddateafter.Time)
				assert.True(t, params.Enddatebefore.Valid)
				assert.Equal(t, endDate, params.Enddatebefore.Time)
				assert.Equal(t, int32(11), params.Querylimit) // 10 + 1 for hasMore check
				assert.False(t, params.Isbackward)
				assert.Equal(t, "EV005", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "backward pagination with before cursor",
			filter: &model.EventFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       nil,
			after:       nil,
			last:        intPtr(5),
			before:      stringPtr("EV100"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, int32(6), params.Querylimit) // 5 + 1 for hasMore check
				assert.True(t, params.Isbackward)
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "EV100", params.Beforecursor)
			},
		},
		{
			name:        "both first and last specified - error",
			filter:      &model.EventFilter{},
			first:       intPtr(10),
			after:       nil,
			last:        intPtr(5),
			before:      nil,
			expectError: true,
			errorMsg:    "cannot specify both first and last",
		},
		{
			name:        "default pagination - no first or last",
			filter:      &model.EventFilter{},
			first:       nil,
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit) // default 10 + 1
				assert.False(t, params.Isbackward)
			},
		},
		{
			name: "only start date filters",
			filter: &model.EventFilter{
				StartDateAfter: &scalars.DateTime{Time: startDate},
			},
			first:       intPtr(20),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.True(t, params.Startdateafter.Valid)
				assert.Equal(t, startDate, params.Startdateafter.Time)
				assert.False(t, params.Startdatebefore.Valid)
			},
		},
		{
			name: "only end date filters",
			filter: &model.EventFilter{
				EndDateBefore: &scalars.DateTime{Time: endDate},
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.True(t, params.Enddatebefore.Valid)
				assert.Equal(t, endDate, params.Enddatebefore.Time)
				assert.False(t, params.Enddateafter.Valid)
			},
		},
		{
			name: "empty cursors",
			filter: &model.EventFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(10),
			after:       stringPtr(""),
			last:        nil,
			before:      stringPtr(""),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "forward pagination with after and before cursors",
			filter: &model.EventFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(15),
			after:       stringPtr("EV010"),
			last:        nil,
			before:      stringPtr("EV050"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.Equal(t, int32(16), params.Querylimit) // 15 + 1
				assert.False(t, params.Isbackward)
				assert.Equal(t, "EV010", params.Aftercursor)
				assert.Equal(t, "EV050", params.Beforecursor)
			},
		},
		{
			name: "minimal filter with IDs",
			filter: &model.EventFilter{
				Ids: []string{"EV001", "EV002", "EV003"},
			},
			first:       intPtr(3),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.Equal(t, []string{"EV001", "EV002", "EV003"}, params.Ids)
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
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit)
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "all date range filters",
			filter: &model.EventFilter{
				StartDateAfter:  &scalars.DateTime{Time: startDate},
				StartDateBefore: &scalars.DateTime{Time: endDate},
				EndDateAfter:    &scalars.DateTime{Time: startDate},
				EndDateBefore:   &scalars.DateTime{Time: endDate},
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetEventsFilteredCursorParams) {
				assert.True(t, params.Startdateafter.Valid)
				assert.True(t, params.Startdatebefore.Valid)
				assert.True(t, params.Enddateafter.Valid)
				assert.True(t, params.Enddatebefore.Valid)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildEventFilterParamsCursor(tt.filter, tt.first, tt.after, tt.last, tt.before)

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

// TestBuildCountEventsFilterParams tests the buildCountEventsFilterParams function
func TestBuildCountEventsFilterParams(t *testing.T) {
	// Test dates
	startDate := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 10, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name   string
		filter *model.EventFilter
		check  func(*testing.T, sqlc.CountEventsFilteredParams)
	}{
		{
			name: "all filters populated",
			filter: &model.EventFilter{
				ProjectID:       stringPtr("PR123"),
				Ids:             []string{"EV001", "EV002"},
				StartDateAfter:  &scalars.DateTime{Time: startDate},
				StartDateBefore: &scalars.DateTime{Time: endDate},
				EndDateAfter:    &scalars.DateTime{Time: startDate},
				EndDateBefore:   &scalars.DateTime{Time: endDate},
			},
			check: func(t *testing.T, params sqlc.CountEventsFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, []string{"EV001", "EV002"}, params.Ids)
				assert.True(t, params.Startdateafter.Valid)
				assert.Equal(t, startDate, params.Startdateafter.Time)
				assert.True(t, params.Startdatebefore.Valid)
				assert.Equal(t, endDate, params.Startdatebefore.Time)
				assert.True(t, params.Enddateafter.Valid)
				assert.Equal(t, startDate, params.Enddateafter.Time)
				assert.True(t, params.Enddatebefore.Valid)
				assert.Equal(t, endDate, params.Enddatebefore.Time)
			},
		},
		{
			name: "only project filter",
			filter: &model.EventFilter{
				ProjectID: stringPtr("PR999"),
			},
			check: func(t *testing.T, params sqlc.CountEventsFilteredParams) {
				assert.Equal(t, "PR999", params.Projectid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Startdateafter.Valid)
			},
		},
		{
			name: "only IDs filter",
			filter: &model.EventFilter{
				Ids: []string{"EV100", "EV200", "EV300"},
			},
			check: func(t *testing.T, params sqlc.CountEventsFilteredParams) {
				assert.Equal(t, []string{"EV100", "EV200", "EV300"}, params.Ids)
				assert.Equal(t, "", params.Projectid)
			},
		},
		{
			name:   "empty filter",
			filter: &model.EventFilter{},
			check: func(t *testing.T, params sqlc.CountEventsFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Startdateafter.Valid)
				assert.False(t, params.Startdatebefore.Valid)
				assert.False(t, params.Enddateafter.Valid)
				assert.False(t, params.Enddatebefore.Valid)
			},
		},
		{
			name: "only start date after filter",
			filter: &model.EventFilter{
				StartDateAfter: &scalars.DateTime{Time: startDate},
			},
			check: func(t *testing.T, params sqlc.CountEventsFilteredParams) {
				assert.True(t, params.Startdateafter.Valid)
				assert.Equal(t, startDate, params.Startdateafter.Time)
				assert.False(t, params.Startdatebefore.Valid)
			},
		},
		{
			name: "project and date filters combined",
			filter: &model.EventFilter{
				ProjectID:      stringPtr("PR123"),
				StartDateAfter: &scalars.DateTime{Time: startDate},
				EndDateBefore:  &scalars.DateTime{Time: endDate},
			},
			check: func(t *testing.T, params sqlc.CountEventsFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.True(t, params.Startdateafter.Valid)
				assert.True(t, params.Enddatebefore.Valid)
			},
		},
		{
			name: "empty IDs array",
			filter: &model.EventFilter{
				Ids: []string{},
			},
			check: func(t *testing.T, params sqlc.CountEventsFilteredParams) {
				assert.NotNil(t, params.Ids)
				assert.Empty(t, params.Ids)
			},
		},
		{
			name:   "nil filter",
			filter: nil,
			check: func(t *testing.T, params sqlc.CountEventsFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCountEventsFilterParams(tt.filter)
			tt.check(t, result)
		})
	}
}

// TestBuildEventCacheKeyParams tests the buildEventCacheKeyParams function
func TestBuildEventCacheKeyParams(t *testing.T) {
	// Test dates
	startDate := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 10, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name   string
		filter *model.EventFilter
		first  *int
		after  *string
		last   *int
		before *string
		check  func(*testing.T, map[string]string)
	}{
		{
			name: "all parameters populated",
			filter: &model.EventFilter{
				ProjectID:       stringPtr("PR123"),
				Ids:             []string{"EV001", "EV002"},
				StartDateAfter:  &scalars.DateTime{Time: startDate},
				StartDateBefore: &scalars.DateTime{Time: endDate},
				EndDateAfter:    &scalars.DateTime{Time: startDate},
				EndDateBefore:   &scalars.DateTime{Time: endDate},
			},
			first:  intPtr(10),
			after:  stringPtr("cursor123"),
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Contains(t, params["ids"], "EV001")
				assert.Contains(t, params["ids"], "EV002")
				assert.Contains(t, params, "startdateafter")
				assert.Contains(t, params, "startdatebefore")
				assert.Contains(t, params, "enddateafter")
				assert.Contains(t, params, "enddatebefore")
				assert.Equal(t, "10", params["first"])
				assert.Equal(t, "cursor123", params["after"])
				assert.NotContains(t, params, "last")
				assert.NotContains(t, params, "before")
			},
		},
		{
			name: "backward pagination",
			filter: &model.EventFilter{
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
				assert.NotContains(t, params, "ids")
			},
		},
		{
			name:   "empty filter",
			filter: &model.EventFilter{},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "10", params["first"])
				assert.NotContains(t, params, "projectid")
			},
		},
		{
			name: "empty string cursors ignored",
			filter: &model.EventFilter{
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
			filter: &model.EventFilter{
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
			name: "only pagination params",
			filter: &model.EventFilter{},
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
			name: "date filters formatted correctly",
			filter: &model.EventFilter{
				StartDateAfter: &scalars.DateTime{Time: startDate},
				EndDateBefore:  &scalars.DateTime{Time: endDate},
			},
			first:  nil,
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Contains(t, params, "startdateafter")
				assert.Contains(t, params, "enddatebefore")
				// Check format contains date components
				assert.Contains(t, params["startdateafter"], "2025")
				assert.Contains(t, params["enddatebefore"], "2025")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildEventCacheKeyParams(tt.filter, tt.first, tt.after, tt.last, tt.before)
			tt.check(t, result)
		})
	}
}

func TestBuildEventCacheKeyParams_Deterministic(t *testing.T) {
	startDate := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	filter := &model.EventFilter{
		ProjectID:      stringPtr("PR123"),
		StartDateAfter: &scalars.DateTime{Time: startDate},
	}
	first := intPtr(10)
	after := stringPtr("cursor")

	// Generate params multiple times
	results := make([]map[string]string, 5)
	for i := 0; i < 5; i++ {
		results[i] = buildEventCacheKeyParams(filter, first, after, nil, nil)
	}

	// All results should be equal
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i], "buildEventCacheKeyParams should be deterministic")
	}
}
