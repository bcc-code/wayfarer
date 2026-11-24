package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services"
)

// mockTeamUpdateRoleChecker is a mock implementation of teamUpdateRoleChecker
type mockTeamUpdateRoleChecker struct {
	mock.Mock
}

func (m *mockTeamUpdateRoleChecker) IsAdmin(ctx context.Context, userID string) bool {
	args := m.Called(ctx, userID)
	return args.Bool(0)
}

func (m *mockTeamUpdateRoleChecker) CanManageProject(ctx context.Context, userID, projectID string) bool {
	args := m.Called(ctx, userID, projectID)
	return args.Bool(0)
}

func (m *mockTeamUpdateRoleChecker) LoadUserRoles(ctx context.Context, userID string) ([]*sqlc.UserRole, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*sqlc.UserRole), args.Error(1)
}

// TestBuildTeamFilterParamsCursor tests the buildTeamFilterParamsCursor function
func TestBuildTeamFilterParamsCursor(t *testing.T) {
	tests := []struct {
		name        string
		filter      *model.TeamFilter
		first       *int
		after       *string
		last        *int
		before      *string
		expectError bool
		errorMsg    string
		check       func(*testing.T, sqlc.GetTeamsFilteredCursorParams)
	}{
		{
			name: "forward pagination with all filters",
			filter: &model.TeamFilter{
				ProjectID:   stringPtr("PR123"),
				SuperTeamID: stringPtr("ST001"),
				Ids:         []string{"TM001", "TM002"},
				NoSuperTeam: boolPtr(false),
				MinMembers:  intPtr(5),
				MaxMembers:  intPtr(15),
			},
			first:       intPtr(10),
			after:       stringPtr("TM005"),
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "ST001", params.Superteamid)
				assert.Equal(t, []string{"TM001", "TM002"}, params.Ids)
				assert.False(t, params.Nosuperteam)
				assert.Equal(t, int32(5), params.Minmembers)
				assert.Equal(t, int32(15), params.Maxmembers)
				assert.Equal(t, int32(11), params.Querylimit) // 10 + 1 for hasMore check
				assert.False(t, params.Isbackward)
				assert.Equal(t, "TM005", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "backward pagination with before cursor",
			filter: &model.TeamFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       nil,
			after:       nil,
			last:        intPtr(5),
			before:      stringPtr("TM100"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, int32(6), params.Querylimit) // 5 + 1 for hasMore check
				assert.True(t, params.Isbackward)
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "TM100", params.Beforecursor)
			},
		},
		{
			name:        "both first and last specified - error",
			filter:      &model.TeamFilter{},
			first:       intPtr(10),
			after:       nil,
			last:        intPtr(5),
			before:      nil,
			expectError: true,
			errorMsg:    "cannot specify both first and last",
		},
		{
			name:        "default pagination - no first or last",
			filter:      &model.TeamFilter{},
			first:       nil,
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit) // default 10 + 1
				assert.False(t, params.Isbackward)
			},
		},
		{
			name: "only noSuperTeam filter",
			filter: &model.TeamFilter{
				NoSuperTeam: boolPtr(true),
			},
			first:       intPtr(20),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.True(t, params.Nosuperteam)
				assert.Equal(t, int32(21), params.Querylimit) // 20 + 1
			},
		},
		{
			name: "member count filters only",
			filter: &model.TeamFilter{
				MinMembers: intPtr(3),
				MaxMembers: intPtr(10),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, int32(3), params.Minmembers)
				assert.Equal(t, int32(10), params.Maxmembers)
			},
		},
		{
			name: "empty cursors",
			filter: &model.TeamFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(10),
			after:       stringPtr(""),
			last:        nil,
			before:      stringPtr(""),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, "", params.Aftercursor)
				assert.Equal(t, "", params.Beforecursor)
			},
		},
		{
			name: "forward pagination with after and before cursors",
			filter: &model.TeamFilter{
				ProjectID: stringPtr("PR123"),
			},
			first:       intPtr(15),
			after:       stringPtr("TM010"),
			last:        nil,
			before:      stringPtr("TM050"),
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, int32(16), params.Querylimit) // 15 + 1
				assert.False(t, params.Isbackward)
				assert.Equal(t, "TM010", params.Aftercursor)
				assert.Equal(t, "TM050", params.Beforecursor)
			},
		},
		{
			name: "minimal filter with IDs",
			filter: &model.TeamFilter{
				Ids: []string{"TM001", "TM002", "TM003"},
			},
			first:       intPtr(3),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, []string{"TM001", "TM002", "TM003"}, params.Ids)
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
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, int32(11), params.Querylimit)
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "superTeamId filter",
			filter: &model.TeamFilter{
				SuperTeamID: stringPtr("ST999"),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, "ST999", params.Superteamid)
			},
		},
		{
			name: "noSuperTeam true overrides superTeamId",
			filter: &model.TeamFilter{
				SuperTeamID: stringPtr("ST001"),
				NoSuperTeam: boolPtr(true),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.True(t, params.Nosuperteam)
				// superTeamId will also be set but SQL query should prioritize noSuperTeam
				assert.Equal(t, "ST001", params.Superteamid)
			},
		},
		{
			name: "only minMembers filter",
			filter: &model.TeamFilter{
				MinMembers: intPtr(5),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, int32(5), params.Minmembers)
				assert.Equal(t, int32(0), params.Maxmembers) // default 0 means no limit
			},
		},
		{
			name: "only maxMembers filter",
			filter: &model.TeamFilter{
				MaxMembers: intPtr(20),
			},
			first:       intPtr(10),
			after:       nil,
			last:        nil,
			before:      nil,
			expectError: false,
			check: func(t *testing.T, params sqlc.GetTeamsFilteredCursorParams) {
				assert.Equal(t, int32(0), params.Minmembers) // default 0 means no limit
				assert.Equal(t, int32(20), params.Maxmembers)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildTeamFilterParamsCursor(tt.filter, tt.first, tt.after, tt.last, tt.before)

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

// TestBuildCountTeamsFilterParams tests the buildCountTeamsFilterParams function
func TestBuildCountTeamsFilterParams(t *testing.T) {
	tests := []struct {
		name   string
		filter *model.TeamFilter
		check  func(*testing.T, sqlc.CountTeamsFilteredParams)
	}{
		{
			name: "all filters populated",
			filter: &model.TeamFilter{
				ProjectID:   stringPtr("PR123"),
				SuperTeamID: stringPtr("ST001"),
				Ids:         []string{"TM001", "TM002"},
				NoSuperTeam: boolPtr(false),
				MinMembers:  intPtr(5),
				MaxMembers:  intPtr(15),
			},
			check: func(t *testing.T, params sqlc.CountTeamsFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, "ST001", params.Superteamid)
				assert.Equal(t, []string{"TM001", "TM002"}, params.Ids)
				assert.False(t, params.Nosuperteam)
				assert.Equal(t, int32(5), params.Minmembers)
				assert.Equal(t, int32(15), params.Maxmembers)
			},
		},
		{
			name: "only project filter",
			filter: &model.TeamFilter{
				ProjectID: stringPtr("PR999"),
			},
			check: func(t *testing.T, params sqlc.CountTeamsFilteredParams) {
				assert.Equal(t, "PR999", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "only IDs filter",
			filter: &model.TeamFilter{
				Ids: []string{"TM100", "TM200", "TM300"},
			},
			check: func(t *testing.T, params sqlc.CountTeamsFilteredParams) {
				assert.Equal(t, []string{"TM100", "TM200", "TM300"}, params.Ids)
				assert.Equal(t, "", params.Projectid)
			},
		},
		{
			name:   "empty filter",
			filter: &model.TeamFilter{},
			check: func(t *testing.T, params sqlc.CountTeamsFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
				assert.False(t, params.Nosuperteam)
			},
		},
		{
			name: "only noSuperTeam filter",
			filter: &model.TeamFilter{
				NoSuperTeam: boolPtr(true),
			},
			check: func(t *testing.T, params sqlc.CountTeamsFilteredParams) {
				assert.True(t, params.Nosuperteam)
			},
		},
		{
			name: "project and member count filters combined",
			filter: &model.TeamFilter{
				ProjectID:  stringPtr("PR123"),
				MinMembers: intPtr(3),
				MaxMembers: intPtr(10),
			},
			check: func(t *testing.T, params sqlc.CountTeamsFilteredParams) {
				assert.Equal(t, "PR123", params.Projectid)
				assert.Equal(t, int32(3), params.Minmembers)
				assert.Equal(t, int32(10), params.Maxmembers)
			},
		},
		{
			name: "empty IDs array",
			filter: &model.TeamFilter{
				Ids: []string{},
			},
			check: func(t *testing.T, params sqlc.CountTeamsFilteredParams) {
				assert.NotNil(t, params.Ids)
				assert.Empty(t, params.Ids)
			},
		},
		{
			name:   "nil filter",
			filter: nil,
			check: func(t *testing.T, params sqlc.CountTeamsFilteredParams) {
				assert.Equal(t, "", params.Projectid)
				assert.Nil(t, params.Ids)
			},
		},
		{
			name: "superTeamId filter",
			filter: &model.TeamFilter{
				SuperTeamID: stringPtr("ST555"),
			},
			check: func(t *testing.T, params sqlc.CountTeamsFilteredParams) {
				assert.Equal(t, "ST555", params.Superteamid)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCountTeamsFilterParams(tt.filter)
			tt.check(t, result)
		})
	}
}

// TestBuildTeamCacheKeyParams tests the buildTeamCacheKeyParams function
func TestBuildTeamCacheKeyParams(t *testing.T) {
	tests := []struct {
		name   string
		filter *model.TeamFilter
		first  *int
		after  *string
		last   *int
		before *string
		check  func(*testing.T, map[string]string)
	}{
		{
			name: "all parameters populated",
			filter: &model.TeamFilter{
				ProjectID:   stringPtr("PR123"),
				SuperTeamID: stringPtr("ST001"),
				Ids:         []string{"TM001", "TM002"},
				NoSuperTeam: boolPtr(false),
				MinMembers:  intPtr(5),
				MaxMembers:  intPtr(15),
			},
			first:  intPtr(10),
			after:  stringPtr("cursor123"),
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "PR123", params["projectid"])
				assert.Equal(t, "ST001", params["superteamid"])
				assert.Contains(t, params["ids"], "TM001")
				assert.Contains(t, params["ids"], "TM002")
				assert.Equal(t, "false", params["nosuperteam"])
				assert.Equal(t, "5", params["minmembers"])
				assert.Equal(t, "15", params["maxmembers"])
				assert.Equal(t, "10", params["first"])
				assert.Equal(t, "cursor123", params["after"])
				assert.NotContains(t, params, "last")
				assert.NotContains(t, params, "before")
			},
		},
		{
			name: "backward pagination",
			filter: &model.TeamFilter{
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
			filter: &model.TeamFilter{},
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
			filter: &model.TeamFilter{
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
			filter: &model.TeamFilter{
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
			filter: &model.TeamFilter{},
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
			name: "noSuperTeam filter",
			filter: &model.TeamFilter{
				NoSuperTeam: boolPtr(true),
			},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "true", params["nosuperteam"])
			},
		},
		{
			name: "member count filters",
			filter: &model.TeamFilter{
				MinMembers: intPtr(3),
				MaxMembers: intPtr(10),
			},
			first:  intPtr(10),
			after:  nil,
			last:   nil,
			before: nil,
			check: func(t *testing.T, params map[string]string) {
				assert.Equal(t, "3", params["minmembers"])
				assert.Equal(t, "10", params["maxmembers"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTeamCacheKeyParams(tt.filter, tt.first, tt.after, tt.last, tt.before)
			tt.check(t, result)
		})
	}
}

func TestValidateTeamUpdateInput(t *testing.T) {
	ctx := context.Background()
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// teamLeadChecker returns a mock where the user is only a team lead (not admin/project admin/church admin).
	teamLeadChecker := func() *mockTeamUpdateRoleChecker {
		m := new(mockTeamUpdateRoleChecker)
		m.On("IsAdmin", mock.Anything, userID).Return(false)
		m.On("CanManageProject", mock.Anything, userID, projectID).Return(false)
		m.On("LoadUserRoles", mock.Anything, userID).Return([]*sqlc.UserRole{
			{Role: string(services.RoleTeamLead)},
		}, nil)
		return m
	}

	tests := []struct {
		name        string
		input       model.UpdateTeamInput
		setup       func() *mockTeamUpdateRoleChecker
		expectError bool
		errorMsg    string
	}{
		{
			name:  "team lead can update name",
			input: model.UpdateTeamInput{Name: stringPtr("New Name")},
			setup: func() *mockTeamUpdateRoleChecker {
				// No role checks needed — guard is not triggered for name-only updates
				return new(mockTeamUpdateRoleChecker)
			},
			expectError: false,
		},
		{
			name:        "team lead cannot update description",
			input:       model.UpdateTeamInput{Description: stringPtr("New Desc")},
			setup:       teamLeadChecker,
			expectError: true,
			errorMsg:    "unauthorized: team leads can only update team name",
		},
		{
			name:        "team lead cannot update leaderboardExcluded",
			input:       model.UpdateTeamInput{LeaderboardExcluded: boolPtr(true)},
			setup:       teamLeadChecker,
			expectError: true,
			errorMsg:    "unauthorized: team leads can only update team name",
		},
		{
			name: "team lead cannot update name and description together",
			input: model.UpdateTeamInput{
				Name:        stringPtr("New Name"),
				Description: stringPtr("New Desc"),
			},
			setup:       teamLeadChecker,
			expectError: true,
			errorMsg:    "unauthorized: team leads can only update team name",
		},
		{
			name: "admin can update all fields",
			input: model.UpdateTeamInput{
				Name:                stringPtr("New Name"),
				Description:         stringPtr("New Desc"),
				LeaderboardExcluded: boolPtr(true),
			},
			setup: func() *mockTeamUpdateRoleChecker {
				m := new(mockTeamUpdateRoleChecker)
				m.On("IsAdmin", mock.Anything, userID).Return(true)
				return m
			},
			expectError: false,
		},
		{
			name: "project admin can update all fields",
			input: model.UpdateTeamInput{
				Description:         stringPtr("New Desc"),
				LeaderboardExcluded: boolPtr(false),
			},
			setup: func() *mockTeamUpdateRoleChecker {
				m := new(mockTeamUpdateRoleChecker)
				m.On("IsAdmin", mock.Anything, userID).Return(false)
				m.On("CanManageProject", mock.Anything, userID, projectID).Return(true)
				return m
			},
			expectError: false,
		},
		{
			name:  "church admin can update description",
			input: model.UpdateTeamInput{Description: stringPtr("New Desc")},
			setup: func() *mockTeamUpdateRoleChecker {
				m := new(mockTeamUpdateRoleChecker)
				m.On("IsAdmin", mock.Anything, userID).Return(false)
				m.On("CanManageProject", mock.Anything, userID, projectID).Return(false)
				m.On("LoadUserRoles", mock.Anything, userID).Return([]*sqlc.UserRole{
					{Role: string(services.RoleChurchAdmin)},
				}, nil)
				return m
			},
			expectError: false,
		},
		{
			name:  "empty input is allowed for any role",
			input: model.UpdateTeamInput{},
			setup: func() *mockTeamUpdateRoleChecker {
				return new(mockTeamUpdateRoleChecker)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := tt.setup()
			err := validateTeamUpdateInput(ctx, checker, userID, projectID, tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}

			checker.AssertExpectations(t)
		})
	}
}

func TestBuildTeamCacheKeyParams_Deterministic(t *testing.T) {
	filter := &model.TeamFilter{
		ProjectID:  stringPtr("PR123"),
		MinMembers: intPtr(5),
		MaxMembers: intPtr(10),
	}
	first := intPtr(10)
	after := stringPtr("cursor")

	// Generate params multiple times
	results := make([]map[string]string, 5)
	for i := 0; i < 5; i++ {
		results[i] = buildTeamCacheKeyParams(filter, first, after, nil, nil)
	}

	// All results should be equal
	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i], "buildTeamCacheKeyParams should be deterministic")
	}
}

func TestWithMeTag(t *testing.T) {
	viewer := "US01K8XV6VK9ED2GBZSQ2VDTAT8A"
	teamLead := "US01K8XV6VK9ED2GBZSQ2VDTAT8B"
	other := "US01K8XV6VK9ED2GBZSQ2VDTAT8C"

	entries := func() []model.LeaderboardEntry {
		return []model.LeaderboardEntry{
			{ID: teamLead, Tags: []model.LeaderboardEntryTag{model.LeaderboardEntryTagTeamLead}},
			{ID: viewer, Tags: []model.LeaderboardEntryTag{}},
			{ID: other, Tags: []model.LeaderboardEntryTag{}},
		}
	}

	t.Run("tags only the viewer's entry", func(t *testing.T) {
		got := withMeTag(entries(), viewer)

		assert.Equal(t, []model.LeaderboardEntryTag{model.LeaderboardEntryTagTeamLead}, got[0].Tags)
		assert.Equal(t, []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe}, got[1].Tags)
		assert.Empty(t, got[2].Tags)
	})

	t.Run("keeps existing tags after ME", func(t *testing.T) {
		got := withMeTag(entries(), teamLead)

		assert.Equal(t, []model.LeaderboardEntryTag{model.LeaderboardEntryTagMe, model.LeaderboardEntryTagTeamLead}, got[0].Tags)
		assert.Empty(t, got[1].Tags)
	})

	t.Run("viewer not in the team gets no ME tag", func(t *testing.T) {
		got := withMeTag(entries(), "US01K8XV6VK9ED2GBZSQ2VDTAT8Z")

		for _, e := range got {
			assert.NotContains(t, e.Tags, model.LeaderboardEntryTagMe)
		}
	})

	t.Run("no viewer gets no ME tag", func(t *testing.T) {
		got := withMeTag(entries(), "")

		for _, e := range got {
			assert.NotContains(t, e.Tags, model.LeaderboardEntryTagMe)
		}
	})

	t.Run("does not modify the loader's entries", func(t *testing.T) {
		in := entries()
		_ = withMeTag(in, teamLead)

		assert.Equal(t, []model.LeaderboardEntryTag{model.LeaderboardEntryTagTeamLead}, in[0].Tags)
	})

	t.Run("two viewers sharing one result each see only themselves", func(t *testing.T) {
		shared := entries()
		forViewer := withMeTag(shared, viewer)
		forOther := withMeTag(shared, other)

		assert.Contains(t, forViewer[1].Tags, model.LeaderboardEntryTagMe)
		assert.NotContains(t, forViewer[2].Tags, model.LeaderboardEntryTagMe)
		assert.Contains(t, forOther[2].Tags, model.LeaderboardEntryTagMe)
		assert.NotContains(t, forOther[1].Tags, model.LeaderboardEntryTagMe)
	})
}

func TestTeamMembersFor(t *testing.T) {
	loaded := []*model.TeamMember{
		{ID: "US1", UserID: "US1", Name: "Counted"},
		{ID: "US2", UserID: "US2", Name: "Blocked"},
	}
	blockedIDs := []string{"US2"}

	t.Run("shows blocked members, not counting, to callers who may see them", func(t *testing.T) {
		got := teamMembersFor(loaded, blockedIDs, true)

		require.Len(t, got, 2)
		assert.True(t, got[0].CountsTowardScores)
		assert.False(t, got[1].CountsTowardScores)
	})

	t.Run("leaves blocked members out for other callers", func(t *testing.T) {
		got := teamMembersFor(loaded, blockedIDs, false)

		require.Len(t, got, 1)
		assert.Equal(t, "US1", got[0].UserID)
		assert.True(t, got[0].CountsTowardScores)
	})

	t.Run("does not modify the loaded members, which the cache shares", func(t *testing.T) {
		teamMembersFor(loaded, nil, true)

		assert.False(t, loaded[0].CountsTowardScores)
	})
}
