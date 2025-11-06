package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// TestBuildSuperTeamFilterParamsCursor tests the buildSuperTeamFilterParamsCursor function
func TestBuildSuperTeamFilterParamsCursor(t *testing.T) {
	tests := []struct {
		name        string
		filter      *model.SuperTeamFilter
		first       *int
		after       *string
		last        *int
		before      *string
		expectError bool
		errorMsg    string
		check       func(*testing.T, sqlc.GetSuperTeamsFilteredCursorParams)
	}{
		{
			name: "forward pagination with all filters",
			filter: &model.SuperTeamFilter{
				ProjectID:  stringPtr("PR123"),
				Ids:        []string{"ST001", "ST002"},
				MinTeams:   intPtr(2),
				MaxTeams:   intPtr(5),
				MinMembers: intPtr(10),
				MaxMembers: intPtr(50),
			},
			first:       intPtr(10),
			after:       stringPtr("ST005"),
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, []string{"ST001", "ST002"}, params.Ids)
				assert.Equal(t, int32(2), params.Minteams)
				assert.Equal(t, int32(5), params.Maxteams)
				assert.Equal(t, int32(10), params.Minmembers)
				assert.Equal(t, int32(50), params.Maxmembers)
				assert.Equal(t, int32(11), params.Querylimit) // 10 + 1 for hasMore check
				assert.False(t, params.Isbackward)
				assert.Equal(t, "ST005", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "backward pagination with before cursor",
			filter: &model.SuperTeamFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       nil,
			after:       nil,
			last:        intPtr(5),
			before:      stringPtr("ST100"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, int32(6), params.Querylimit) // 5 + 1 for hasMore check
				assert.True(t, params.Isbackward)
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "ST100", params.Beforecursor)
			},
		},
		{
			name:        "both first and last specified - error",
			filter:      &model.SuperTeamFilter{},
			first:       intPtr(10),
			after:       nil,
			last:        intPtr(5),
			before:      nil,
			expectError: true,
			errorMsg:    "cannot specify both first and last",
		},
		{
			name:        "default pagination - no first or last",
			filter:      &model.SuperTeamFilter{},
			first:       nil,
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit) // default 10 + 1
				assert.False(t, params.Isbackward)
			},
		},
		{
			name: "only team count filters",
			filter: &model.SuperTeamFilter{
				MinTeams: intPtr(2),
				MaxTeams: intPtr(10),
			},
			first:       intPtr(20),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, int32(2), params.Minteams)
				assert.Equal(t, int32(10), params.Maxteams)
				assert.Equal(t, int32(21), params.Querylimit) // 20 + 1
			},
		},
		{
			name: "only member count filters",
			filter: &model.SuperTeamFilter{
				MinMembers: intPtr(15),
				MaxMembers: intPtr(100),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, int32(15), params.Minmembers)
				assert.Equal(t, int32(100), params.Maxmembers)
			},
		},
		{
			name: "empty cursors",
			filter: &model.SuperTeamFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(10),
			after:       stringPtr(""),
			last:        nil,
			before:      stringPtr(""),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "forward pagination with after and before cursors",
			filter: &model.SuperTeamFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(15),
			after:       stringPtr("ST010"),
			last:        nil,
			before:      stringPtr("ST050"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, int32(16), params.Querylimit) // 15 + 1
				assert.False(t, params.Isbackward)
				assert.Equal(t, "ST010", params.Aftercursor)
				assert.Equal(t, "ST050", params.Beforecursor)
			},
		},
		{
			name: "minimal filter with IDs",
			filter: &model.SuperTeamFilter{
				Ids: []string{"ST001", "ST002", "ST003"},
			},
			first:       intPtr(3),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, []string{"ST001", "ST002", "ST003"}, params.Ids)
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
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit)
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "only minTeams filter",
			filter: &model.SuperTeamFilter{
				MinTeams: intPtr(3),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, int32(3), params.Minteams)
				assert.Equal(t, int32(0), params.Maxteams) // default 0 means no limit
			},
		},
		{
			name: "only maxTeams filter",
			filter: &model.SuperTeamFilter{
				MaxTeams: intPtr(8),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, int32(0), params.Minteams) // default 0 means no limit
				assert.Equal(t, int32(8), params.Maxteams)
			},
		},
		{
			name: "only minMembers filter",
			filter: &model.SuperTeamFilter{
				MinMembers: intPtr(20),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, int32(20), params.Minmembers)
				assert.Equal(t, int32(0), params.Maxmembers) // default 0 means no limit
			},
		},
		{
			name: "only maxMembers filter",
			filter: &model.SuperTeamFilter{
				MaxMembers: intPtr(75),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetSuperTeamsFilteredCursorParams) {
				assert.Equal(t, int32(0), params.Minmembers) // default 0 means no limit
				assert.Equal(t, int32(75), params.Maxmembers)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildSuperTeamFilterParamsCursor(tt.filter, tt.first, tt.after, tt.last, tt.before)

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

// TestBuildCountSuperTeamsFilterParams tests the buildCountSuperTeamsFilterParams function
func TestBuildCountSuperTeamsFilterParams(t *testing.T) {
	tests := []struct {
		name   string
		filter *model.SuperTeamFilter
		check  func(*testing.T, sqlc.CountSuperTeamsFilteredParams)
	}{
		{
			name: "all filters populated",
			filter: &model.SuperTeamFilter{
				ProjectID:  stringPtr("PR123"),
				Ids:        []string{"ST001", "ST002"},
				MinTeams:   intPtr(2),
				MaxTeams:   intPtr(5),
				MinMembers: intPtr(10),
				MaxMembers: intPtr(50),
			},
			check: func(t *testing.T, params sqlc.CountSuperTeamsFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, []string{"ST001", "ST002"}, params.Ids)
				assert.Equal(t, int32(2), params.Minteams)
				assert.Equal(t, int32(5), params.Maxteams)
				assert.Equal(t, int32(10), params.Minmembers)
				assert.Equal(t, int32(50), params.Maxmembers)
			},
		},
		{
			name: "only project filter",
			filter: &model.SuperTeamFilter{
				ProjectID: stringPtr("PR999"),
			},
			check: func(t *testing.T, params sqlc.CountSuperTeamsFilteredParams) {
				assert.Equal(t, "PR999", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "only IDs filter",
			filter: &model.SuperTeamFilter{
				Ids: []string{"ST100", "ST200", "ST300"},
			},
			check: func(t *testing.T, params sqlc.CountSuperTeamsFilteredParams) {
				assert.Equal(t, []string{"ST100", "ST200", "ST300"}, params.Ids)
				assert.Equal(t, "", params.Projectid)
			},
		},
		{
			name:   "empty filter",
			filter: &model.SuperTeamFilter{},
			check: func(t *testing.T, params sqlc.CountSuperTeamsFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "project and team count filters combined",
			filter: &model.SuperTeamFilter{
				ProjectID: stringPtr("PR123"),
				MinTeams:  intPtr(2),
				MaxTeams:  intPtr(8),
			},
			check: func(t *testing.T, params sqlc.CountSuperTeamsFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, int32(2), params.Minteams)
				assert.Equal(t, int32(8), params.Maxteams)
			},
		},
		{
			name: "project and member count filters combined",
			filter: &model.SuperTeamFilter{
				ProjectID:  stringPtr("PR123"),
				MinMembers: intPtr(15),
				MaxMembers: intPtr(75),
			},
			check: func(t *testing.T, params sqlc.CountSuperTeamsFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, int32(15), params.Minmembers)
				assert.Equal(t, int32(75), params.Maxmembers)
			},
		},
		{
			name: "empty IDs array",
			filter: &model.SuperTeamFilter{
				Ids: []string{},
			},
			check: func(t *testing.T, params sqlc.CountSuperTeamsFilteredParams) {
				assert.NotNil(t, params.Ids)
				assert.Empty(t, params.Ids)
			},
		},
		{
			name:   "nil filter",
			filter: nil,
			check: func(t *testing.T, params sqlc.CountSuperTeamsFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCountSuperTeamsFilterParams(tt.filter)
			tt.check(t, result)
		})
	}
}

// TestBuildSuperTeamCacheKeyParams tests the buildSuperTeamCacheKeyParams function
func TestBuildSuperTeamCacheKeyParams(t *testing.T) {
	tests := []struct {
		name   string
		filter *model.SuperTeamFilter
		first  *int
		after  *string
		last   *int
		before *string
		check  func(*testing.T, map[string]string)
	}{
		{
			name: "all parameters populated",
			filter: &model.SuperTeamFilter{
				ProjectID:  stringPtr("PR123"),
				Ids:        []string{"ST001", "ST002"},
				MinTeams:   intPtr(2),
				MaxTeams:   intPtr(5),
				MinMembers: intPtr(10),
				MaxMembers: intPtr(50),
			},
			first:  intPtr(10),
			after:  stringPtr("cursor123"),
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Contains(t, params["ids"], "ST001")
				assert.Contains(t, params["ids"], "ST002")
				assert.Equal(t, "2", params["minteams"])
				assert.Equal(t, "5", params["maxteams"])
				assert.Equal(t, "10", params["minmembers"])
				assert.Equal(t, "50", params["maxmembers"])
				assert.Equal(t, "10", params["first"])
				assert.Equal(t, "cursor123", params["after"])
				assert.NotContains(t, params, "last")
				assert.NotContains(t, params, "before")
			},
		},
		{
			name: "backward pagination",
			filter: &model.SuperTeamFilter{
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
			filter: &model.SuperTeamFilter{},
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
			filter: &model.SuperTeamFilter{
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
			filter: &model.SuperTeamFilter{
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
			filter: &model.SuperTeamFilter{},
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
			name: "team count filters",
			filter: &model.SuperTeamFilter{
				MinTeams: intPtr(3),
				MaxTeams: intPtr(10),
			},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "3", params["minteams"])
				assert.Equal(t, "10", params["maxteams"])
			},
		},
		{
			name: "member count filters",
			filter: &model.SuperTeamFilter{
				MinMembers: intPtr(20),
				MaxMembers: intPtr(100),
			},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "20", params["minmembers"])
				assert.Equal(t, "100", params["maxmembers"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSuperTeamCacheKeyParams(tt.filter, tt.first, tt.after, tt.last, tt.before)
			tt.check(t, result)
		})
	}
}

func TestBuildSuperTeamCacheKeyParams_Deterministic(t *testing.T) {
	filter := &model.SuperTeamFilter{
		ProjectID:  stringPtr("PR123"),
		MinTeams:   intPtr(2),
		MinMembers: intPtr(10),
	}
	first := intPtr(10)
	after := stringPtr("cursor")

	// Generate params multiple times
	results := make([]map[string]string, 5)
	for i := 0; i < 5; i++ {
		results[i] = buildSuperTeamCacheKeyParams(filter, first, after, nil, nil)
	}

	// All results should be equal
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i], "buildSuperTeamCacheKeyParams should be deterministic")
	}
}
