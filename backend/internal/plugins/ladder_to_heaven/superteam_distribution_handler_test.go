package ladder_to_heaven

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuperteamDistributionHandler_Preview_MissingProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &superteamDistributionHandler{
		db:        nil,
		cache:     nil,
		jwtConfig: config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/preview-superteams", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", testUserID)
	c.Set("user_roles", []string{"admin"})

	handler.preview(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "project_id is required", response["error"])
}

func TestSuperteamDistributionHandler_Preview_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &superteamDistributionHandler{
		db:        nil,
		cache:     nil,
		jwtConfig: config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/preview-superteams?project_id=PR123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// No user_id set = not authenticated

	handler.preview(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSuperteamDistributionHandler_Preview_InsufficientPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &superteamDistributionHandler{
		db:        nil,
		cache:     nil,
		jwtConfig: config.JWTConfig{Issuer: "wayfarer"},
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/preview-superteams?project_id=PR123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", testUserID)
	c.Set("user_roles", []string{"user"}) // Not admin

	handler.preview(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSuperteamDistributionHandler_Preview_EmptyTeams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQuerier := &mockDistributionQuerier{
		teams: []*sqlc.GetTeamsWithScoresForDistributionRow{},
	}

	handler := &superteamDistributionHandler{
		db:          nil,
		cache:       nil,
		jwtConfig:   config.JWTConfig{Issuer: "wayfarer"},
		testQuerier: mockQuerier,
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/preview-superteams?project_id=PR123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", testUserID)
	c.Set("user_roles", []string{"admin"})

	handler.preview(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DistributionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Empty(t, response.Superteams)
	assert.Zero(t, response.Variance)
}

func TestSuperteamDistributionHandler_Preview_WithTeams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQuerier := &mockDistributionQuerier{
		teams: []*sqlc.GetTeamsWithScoresForDistributionRow{
			{TeamID: "TM001", TeamName: "Team 1", ChurchID: "CH001", TotalScore: 1000, MemberCount: 5},
			{TeamID: "TM002", TeamName: "Team 2", ChurchID: "CH001", TotalScore: 800, MemberCount: 4},
			{TeamID: "TM003", TeamName: "Team 3", ChurchID: "CH002", TotalScore: 1200, MemberCount: 6},
			{TeamID: "TM004", TeamName: "Team 4", ChurchID: "CH002", TotalScore: 600, MemberCount: 3},
			{TeamID: "TM005", TeamName: "Team 5", ChurchID: "CH003", TotalScore: 900, MemberCount: 4},
			{TeamID: "TM006", TeamName: "Team 6", ChurchID: "CH004", TotalScore: 700, MemberCount: 3},
		},
	}

	handler := &superteamDistributionHandler{
		db:          nil,
		cache:       nil,
		jwtConfig:   config.JWTConfig{Issuer: "wayfarer"},
		testQuerier: mockQuerier,
	}

	req := httptest.NewRequest(http.MethodGet, "/plugins/ladder-to-heaven/preview-superteams?project_id=PR123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", testUserID)
	c.Set("user_roles", []string{"admin"})

	handler.preview(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DistributionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have 4 superteams
	assert.Len(t, response.Superteams, 4)

	// Verify superteam names
	names := make([]string, 4)
	for i, st := range response.Superteams {
		names[i] = st.Name
	}
	assert.Contains(t, names, "Purple")
	assert.Contains(t, names, "Green")
	assert.Contains(t, names, "Red")
	assert.Contains(t, names, "Yellow")

	// No IDs should be assigned in preview
	for _, st := range response.Superteams {
		assert.Empty(t, st.SuperTeamID, "Preview should not assign IDs")
	}

	// Calculate total teams distributed
	totalTeams := 0
	for _, st := range response.Superteams {
		totalTeams += st.TeamCount
	}
	assert.Equal(t, 6, totalTeams, "All 6 teams should be distributed")

	// Check that variance is computed
	assert.GreaterOrEqual(t, response.Variance, 0.0)
}

func TestSuperteamDistributionHandler_CalculateVariance(t *testing.T) {
	handler := &superteamDistributionHandler{}

	tests := []struct {
		name     string
		buckets  []SuperteamResult
		expected float64
	}{
		{
			name:     "empty buckets",
			buckets:  []SuperteamResult{},
			expected: 0,
		},
		{
			name: "perfectly balanced",
			buckets: []SuperteamResult{
				{TotalScore: 1000},
				{TotalScore: 1000},
				{TotalScore: 1000},
				{TotalScore: 1000},
			},
			expected: 0,
		},
		{
			name: "one off",
			buckets: []SuperteamResult{
				{TotalScore: 1100},
				{TotalScore: 1000},
				{TotalScore: 1000},
				{TotalScore: 900},
			},
			// Mean = 1000, variance = ((100)^2 + 0 + 0 + (100)^2) / 4 = 5000
			expected: 5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.calculateVariance(tt.buckets)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSuperteamDistributionHandler_CalculateDistribution_ChurchCohesion(t *testing.T) {
	handler := &superteamDistributionHandler{}

	// With church cohesion as a hard constraint, all teams from the same church
	// should always be kept together, regardless of score balance impact.
	teams := []*sqlc.GetTeamsWithScoresForDistributionRow{
		{TeamID: "TM001", TeamName: "Team 1", ChurchID: "CH001", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM002", TeamName: "Team 2", ChurchID: "CH001", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM003", TeamName: "Team 3", ChurchID: "CH001", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM004", TeamName: "Team 4", ChurchID: "CH001", TotalScore: 1000, MemberCount: 5},
	}

	result := handler.calculateDistribution(teams, false)

	// All 4 teams from the same church should be in the same superteam
	var superteamWithChurch001 *SuperteamResult
	for i := range result.Superteams {
		for _, team := range result.Superteams[i].Teams {
			if team.ChurchID == "CH001" {
				superteamWithChurch001 = &result.Superteams[i]
				break
			}
		}
		if superteamWithChurch001 != nil {
			break
		}
	}

	require.NotNil(t, superteamWithChurch001, "Should find superteam with CH001 teams")
	assert.Equal(t, 4, superteamWithChurch001.TeamCount, "All 4 teams from CH001 should be in same superteam")

	// Verify all teams in that superteam are from CH001
	for _, team := range superteamWithChurch001.Teams {
		assert.Equal(t, "CH001", team.ChurchID)
	}

	// All 4 teams should be distributed
	totalTeams := 0
	for _, st := range result.Superteams {
		totalTeams += st.TeamCount
	}
	assert.Equal(t, 4, totalTeams)

	// Total score should be 4000
	assert.Equal(t, int64(4000), superteamWithChurch001.TotalScore)
}

func TestSuperteamDistributionHandler_CalculateDistribution_ChurchPreference(t *testing.T) {
	handler := &superteamDistributionHandler{}

	// When churches can be kept together without exceeding the 10% threshold, they should be.
	// Total: 4000, target: 1000, threshold: 100
	// Church A has 2 teams totaling 600 (within threshold for one bucket)
	// Church B has 2 teams totaling 600 (within threshold for one bucket)
	// Church C has 2 teams totaling 1400 (within threshold for one bucket)
	// Church D has 2 teams totaling 1400 (within threshold for one bucket)
	teams := []*sqlc.GetTeamsWithScoresForDistributionRow{
		{TeamID: "TM001", TeamName: "Team 1", ChurchID: "CH_A", TotalScore: 400, MemberCount: 3},
		{TeamID: "TM002", TeamName: "Team 2", ChurchID: "CH_A", TotalScore: 200, MemberCount: 2},
		{TeamID: "TM003", TeamName: "Team 3", ChurchID: "CH_B", TotalScore: 400, MemberCount: 3},
		{TeamID: "TM004", TeamName: "Team 4", ChurchID: "CH_B", TotalScore: 200, MemberCount: 2},
		{TeamID: "TM005", TeamName: "Team 5", ChurchID: "CH_C", TotalScore: 800, MemberCount: 5},
		{TeamID: "TM006", TeamName: "Team 6", ChurchID: "CH_C", TotalScore: 600, MemberCount: 4},
		{TeamID: "TM007", TeamName: "Team 7", ChurchID: "CH_D", TotalScore: 800, MemberCount: 5},
		{TeamID: "TM008", TeamName: "Team 8", ChurchID: "CH_D", TotalScore: 600, MemberCount: 4},
	}

	result := handler.calculateDistribution(teams, false)

	// Count how many superteams each church appears in
	churchSuperteams := make(map[string]map[string]bool)
	for _, st := range result.Superteams {
		for _, team := range st.Teams {
			if team.ChurchID != "" {
				if churchSuperteams[team.ChurchID] == nil {
					churchSuperteams[team.ChurchID] = make(map[string]bool)
				}
				churchSuperteams[team.ChurchID][st.Name] = true
			}
		}
	}

	// Each church should ideally be in only 1 superteam (kept together)
	for churchID, superteams := range churchSuperteams {
		assert.LessOrEqual(t, len(superteams), 2,
			"Church %s should be in at most 2 superteams (prefer keeping together)", churchID)
	}

	// All 8 teams should be distributed
	totalTeams := 0
	for _, st := range result.Superteams {
		totalTeams += st.TeamCount
	}
	assert.Equal(t, 8, totalTeams)
}

func TestSuperteamDistributionHandler_CalculateDistribution_BalancedAssignment(t *testing.T) {
	handler := &superteamDistributionHandler{}

	// Create 4 teams with equal total scores from different churches
	teams := []*sqlc.GetTeamsWithScoresForDistributionRow{
		{TeamID: "TM001", TeamName: "Team 1", ChurchID: "CH001", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM002", TeamName: "Team 2", ChurchID: "CH002", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM003", TeamName: "Team 3", ChurchID: "CH003", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM004", TeamName: "Team 4", ChurchID: "CH004", TotalScore: 1000, MemberCount: 5},
	}

	result := handler.calculateDistribution(teams, true)

	// Each superteam should have exactly one team
	for _, st := range result.Superteams {
		assert.Equal(t, 1, st.TeamCount, "Each superteam should have exactly one team")
		assert.Equal(t, int64(1000), st.TotalScore, "Each superteam should have total score 1000")
	}

	// Variance should be 0 (perfectly balanced)
	assert.Equal(t, float64(0), result.Variance)
}

func TestSuperteamDistributionHandler_CalculateDistribution_UnevenChurchSizes(t *testing.T) {
	handler := &superteamDistributionHandler{}

	// Test that even with very uneven church sizes, all teams from the same
	// church are kept together (church cohesion is a hard constraint).
	teams := []*sqlc.GetTeamsWithScoresForDistributionRow{
		// Large church with 8 teams
		{TeamID: "TM001", TeamName: "Team 1", ChurchID: "CH_LARGE", TotalScore: 1000, MemberCount: 8},
		{TeamID: "TM002", TeamName: "Team 2", ChurchID: "CH_LARGE", TotalScore: 900, MemberCount: 7},
		{TeamID: "TM003", TeamName: "Team 3", ChurchID: "CH_LARGE", TotalScore: 850, MemberCount: 6},
		{TeamID: "TM004", TeamName: "Team 4", ChurchID: "CH_LARGE", TotalScore: 800, MemberCount: 8},
		{TeamID: "TM005", TeamName: "Team 5", ChurchID: "CH_LARGE", TotalScore: 750, MemberCount: 7},
		{TeamID: "TM006", TeamName: "Team 6", ChurchID: "CH_LARGE", TotalScore: 700, MemberCount: 6},
		{TeamID: "TM007", TeamName: "Team 7", ChurchID: "CH_LARGE", TotalScore: 650, MemberCount: 5},
		{TeamID: "TM008", TeamName: "Team 8", ChurchID: "CH_LARGE", TotalScore: 600, MemberCount: 5},
		// Smaller churches with 2 teams each
		{TeamID: "TM009", TeamName: "Team 9", ChurchID: "CH_SMALL1", TotalScore: 500, MemberCount: 4},
		{TeamID: "TM010", TeamName: "Team 10", ChurchID: "CH_SMALL1", TotalScore: 450, MemberCount: 4},
		{TeamID: "TM011", TeamName: "Team 11", ChurchID: "CH_SMALL2", TotalScore: 400, MemberCount: 3},
		{TeamID: "TM012", TeamName: "Team 12", ChurchID: "CH_SMALL2", TotalScore: 350, MemberCount: 3},
	}

	result := handler.calculateDistribution(teams, false)

	// All 12 teams should be distributed
	totalTeams := 0
	for _, st := range result.Superteams {
		totalTeams += st.TeamCount
	}
	assert.Equal(t, 12, totalTeams)

	// The large church should be kept together in one superteam (church cohesion)
	largeChurchSuperteams := make(map[string]bool)
	for _, st := range result.Superteams {
		for _, team := range st.Teams {
			if team.ChurchID == "CH_LARGE" {
				largeChurchSuperteams[st.Name] = true
			}
		}
	}
	assert.Equal(t, 1, len(largeChurchSuperteams),
		"Large church should be kept together in one superteam")

	// Small churches should also each be in one superteam
	small1Superteams := make(map[string]bool)
	small2Superteams := make(map[string]bool)
	for _, st := range result.Superteams {
		for _, team := range st.Teams {
			if team.ChurchID == "CH_SMALL1" {
				small1Superteams[st.Name] = true
			}
			if team.ChurchID == "CH_SMALL2" {
				small2Superteams[st.Name] = true
			}
		}
	}
	assert.Equal(t, 1, len(small1Superteams), "CH_SMALL1 should be kept together")
	assert.Equal(t, 1, len(small2Superteams), "CH_SMALL2 should be kept together")
}

func TestSuperteamDistributionHandler_CalculateDistribution_GeneratesIDs(t *testing.T) {
	handler := &superteamDistributionHandler{}

	teams := []*sqlc.GetTeamsWithScoresForDistributionRow{
		{TeamID: "TM001", TeamName: "Team 1", ChurchID: "CH001", TotalScore: 1000, MemberCount: 5},
	}

	result := handler.calculateDistribution(teams, true)

	// All superteams should have IDs when generateIDs is true
	for _, st := range result.Superteams {
		assert.NotEmpty(t, st.SuperTeamID, "SuperTeamID should be generated")
		assert.Len(t, st.SuperTeamID, 28, "SuperTeamID should be 28 characters (ST prefix + ULID)")
		assert.Equal(t, "ST", st.SuperTeamID[:2], "SuperTeamID should have ST prefix")
	}

	// Test without ID generation
	resultNoIDs := handler.calculateDistribution(teams, false)
	for _, st := range resultNoIDs.Superteams {
		assert.Empty(t, st.SuperTeamID, "SuperTeamID should be empty when generateIDs is false")
	}
}

func TestHasDistributionRole(t *testing.T) {
	tests := []struct {
		name      string
		userRoles []string
		expected  bool
	}{
		{
			name:      "no roles",
			userRoles: []string{},
			expected:  false,
		},
		{
			name:      "user role only",
			userRoles: []string{"user"},
			expected:  false,
		},
		{
			name:      "church_admin role - not allowed",
			userRoles: []string{"church_admin"},
			expected:  false,
		},
		{
			name:      "admin role",
			userRoles: []string{"admin"},
			expected:  true,
		},
		{
			name:      "superadmin role",
			userRoles: []string{"superadmin"},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasDistributionRole(tt.userRoles)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock implementation for testing
type mockDistributionQuerier struct {
	teams []*sqlc.GetTeamsWithScoresForDistributionRow
}

func (m *mockDistributionQuerier) GetTeamsWithScoresForDistribution(_ context.Context, _ string) ([]*sqlc.GetTeamsWithScoresForDistributionRow, error) {
	return m.teams, nil
}

func (m *mockDistributionQuerier) ClearSuperTeamAssignmentsForProject(_ context.Context, _ string) error {
	return nil
}

func (m *mockDistributionQuerier) DeleteSuperTeamsByProjectID(_ context.Context, _ string) error {
	return nil
}

func (m *mockDistributionQuerier) CreateSuperTeam(_ context.Context, _ sqlc.CreateSuperTeamParams) (*sqlc.SuperTeam, error) {
	return &sqlc.SuperTeam{}, nil
}

func (m *mockDistributionQuerier) AssignTeamToSuperTeam(_ context.Context, _ sqlc.AssignTeamToSuperTeamParams) error {
	return nil
}

// visualizeDistribution prints a visual representation of the superteam distribution
func visualizeDistribution(t *testing.T, result DistributionResponse) {
	t.Helper()

	// Calculate totals
	var totalScore int64
	var totalTeams int
	for _, st := range result.Superteams {
		totalScore += st.TotalScore
		totalTeams += st.TeamCount
	}
	targetScore := totalScore / 4

	t.Logf("\n")
	t.Logf("============================================================")
	t.Logf("SUPERTEAM DISTRIBUTION VISUALIZATION")
	t.Logf("============================================================")
	t.Logf("Total Teams: %d | Total Score: %d | Target per ST: %d", totalTeams, totalScore, targetScore)
	t.Logf("Deviation Threshold (5%%): %d", totalScore/80)
	t.Logf("------------------------------------------------------------")

	// Collect all unique churches and assign them short labels
	churchLabels := make(map[string]string)
	labelIndex := 0
	labels := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P"}
	for _, st := range result.Superteams {
		for _, team := range st.Teams {
			if team.ChurchID != "" {
				if _, exists := churchLabels[team.ChurchID]; !exists {
					if labelIndex < len(labels) {
						churchLabels[team.ChurchID] = labels[labelIndex]
						labelIndex++
					} else {
						churchLabels[team.ChurchID] = "?"
					}
				}
			}
		}
	}

	// Print church legend
	t.Logf("\nCHURCH LEGEND:")
	for churchID, label := range churchLabels {
		t.Logf("  [%s] = %s", label, churchID)
	}
	t.Logf("")

	// Print each superteam
	for _, st := range result.Superteams {
		deviation := st.TotalScore - targetScore
		deviationPct := float64(deviation) / float64(targetScore) * 100

		t.Logf("%-8s | Teams: %3d | Score: %8d | Deviation: %+6d (%+.1f%%)",
			st.Name, st.TeamCount, st.TotalScore, deviation, deviationPct)

		// Group teams by church for visualization
		churchTeams := make(map[string][]TeamInfo)
		for _, team := range st.Teams {
			label := churchLabels[team.ChurchID]
			if label == "" {
				label = "?"
			}
			churchTeams[label] = append(churchTeams[label], team)
		}

		// Build visual bar showing church distribution
		var bar string
		for label, teams := range churchTeams {
			for range teams {
				bar += "[" + label + "]"
			}
		}
		t.Logf("         | %s", bar)
		t.Logf("")
	}

	// Print church distribution summary
	t.Logf("------------------------------------------------------------")
	t.Logf("CHURCH DISTRIBUTION SUMMARY:")
	churchInSuperteams := make(map[string][]string)
	for _, st := range result.Superteams {
		for _, team := range st.Teams {
			if team.ChurchID != "" {
				label := churchLabels[team.ChurchID]
				found := false
				for _, existing := range churchInSuperteams[label] {
					if existing == st.Name {
						found = true
						break
					}
				}
				if !found {
					churchInSuperteams[label] = append(churchInSuperteams[label], st.Name)
				}
			}
		}
	}
	for label, superteams := range churchInSuperteams {
		status := "KEPT TOGETHER"
		if len(superteams) > 1 {
			status = "SPLIT"
		}
		t.Logf("  [%s] in %d superteam(s): %v - %s", label, len(superteams), superteams, status)
	}
	t.Logf("============================================================\n")
}

func TestSuperteamDistributionHandler_VisualizeDistribution(t *testing.T) {
	handler := &superteamDistributionHandler{}

	// Realistic scenario with mixed church sizes
	teams := []*sqlc.GetTeamsWithScoresForDistributionRow{
		// Large church (like Oslo) with many teams
		{TeamID: "TM001", TeamName: "Oslo Team 1", ChurchID: "CH_OSLO", TotalScore: 50000, MemberCount: 8},
		{TeamID: "TM002", TeamName: "Oslo Team 2", ChurchID: "CH_OSLO", TotalScore: 45000, MemberCount: 7},
		{TeamID: "TM003", TeamName: "Oslo Team 3", ChurchID: "CH_OSLO", TotalScore: 42000, MemberCount: 6},
		{TeamID: "TM004", TeamName: "Oslo Team 4", ChurchID: "CH_OSLO", TotalScore: 38000, MemberCount: 8},
		{TeamID: "TM005", TeamName: "Oslo Team 5", ChurchID: "CH_OSLO", TotalScore: 35000, MemberCount: 7},
		{TeamID: "TM006", TeamName: "Oslo Team 6", ChurchID: "CH_OSLO", TotalScore: 32000, MemberCount: 6},
		{TeamID: "TM007", TeamName: "Oslo Team 7", ChurchID: "CH_OSLO", TotalScore: 28000, MemberCount: 5},
		{TeamID: "TM008", TeamName: "Oslo Team 8", ChurchID: "CH_OSLO", TotalScore: 25000, MemberCount: 5},

		// Medium church
		{TeamID: "TM009", TeamName: "Bergen Team 1", ChurchID: "CH_BERGEN", TotalScore: 30000, MemberCount: 6},
		{TeamID: "TM010", TeamName: "Bergen Team 2", ChurchID: "CH_BERGEN", TotalScore: 28000, MemberCount: 5},
		{TeamID: "TM011", TeamName: "Bergen Team 3", ChurchID: "CH_BERGEN", TotalScore: 22000, MemberCount: 4},

		// Medium church
		{TeamID: "TM012", TeamName: "Trondheim Team 1", ChurchID: "CH_TRONDHEIM", TotalScore: 27000, MemberCount: 5},
		{TeamID: "TM013", TeamName: "Trondheim Team 2", ChurchID: "CH_TRONDHEIM", TotalScore: 24000, MemberCount: 5},
		{TeamID: "TM014", TeamName: "Trondheim Team 3", ChurchID: "CH_TRONDHEIM", TotalScore: 20000, MemberCount: 4},

		// Small churches
		{TeamID: "TM015", TeamName: "Stavanger Team 1", ChurchID: "CH_STAVANGER", TotalScore: 18000, MemberCount: 4},
		{TeamID: "TM016", TeamName: "Stavanger Team 2", ChurchID: "CH_STAVANGER", TotalScore: 15000, MemberCount: 3},

		{TeamID: "TM017", TeamName: "Kristiansand Team 1", ChurchID: "CH_KRISTIANSAND", TotalScore: 16000, MemberCount: 4},
		{TeamID: "TM018", TeamName: "Kristiansand Team 2", ChurchID: "CH_KRISTIANSAND", TotalScore: 12000, MemberCount: 3},

		{TeamID: "TM019", TeamName: "Tromso Team 1", ChurchID: "CH_TROMSO", TotalScore: 14000, MemberCount: 3},
		{TeamID: "TM020", TeamName: "Tromso Team 2", ChurchID: "CH_TROMSO", TotalScore: 10000, MemberCount: 3},
	}

	result := handler.calculateDistribution(teams, false)

	visualizeDistribution(t, result)

	// Basic assertions
	totalTeams := 0
	for _, st := range result.Superteams {
		totalTeams += st.TeamCount
	}
	assert.Equal(t, 20, totalTeams, "All teams should be distributed")
}

func TestCalculateDistribution_PriorityChurches(t *testing.T) {
	// Temporarily override priorityChurchIDs for this test
	origPriority := priorityChurchIDs
	priorityChurchIDs = []string{"CH_P1", "CH_P2", "CH_P3", "CH_P4"}
	defer func() { priorityChurchIDs = origPriority }()

	handler := &superteamDistributionHandler{}

	// 4 priority churches + 4 other churches, 2 teams each = 16 teams
	teams := []*sqlc.GetTeamsWithScoresForDistributionRow{
		// Priority churches
		{TeamID: "TM001", TeamName: "P1 Team 1", ChurchID: "CH_P1", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM002", TeamName: "P1 Team 2", ChurchID: "CH_P1", TotalScore: 800, MemberCount: 4},
		{TeamID: "TM003", TeamName: "P2 Team 1", ChurchID: "CH_P2", TotalScore: 950, MemberCount: 5},
		{TeamID: "TM004", TeamName: "P2 Team 2", ChurchID: "CH_P2", TotalScore: 750, MemberCount: 4},
		{TeamID: "TM005", TeamName: "P3 Team 1", ChurchID: "CH_P3", TotalScore: 900, MemberCount: 5},
		{TeamID: "TM006", TeamName: "P3 Team 2", ChurchID: "CH_P3", TotalScore: 700, MemberCount: 4},
		{TeamID: "TM007", TeamName: "P4 Team 1", ChurchID: "CH_P4", TotalScore: 850, MemberCount: 5},
		{TeamID: "TM008", TeamName: "P4 Team 2", ChurchID: "CH_P4", TotalScore: 650, MemberCount: 4},
		// Other churches
		{TeamID: "TM009", TeamName: "O1 Team 1", ChurchID: "CH_O1", TotalScore: 500, MemberCount: 3},
		{TeamID: "TM010", TeamName: "O1 Team 2", ChurchID: "CH_O1", TotalScore: 400, MemberCount: 3},
		{TeamID: "TM011", TeamName: "O2 Team 1", ChurchID: "CH_O2", TotalScore: 450, MemberCount: 3},
		{TeamID: "TM012", TeamName: "O2 Team 2", ChurchID: "CH_O2", TotalScore: 350, MemberCount: 3},
		{TeamID: "TM013", TeamName: "O3 Team 1", ChurchID: "CH_O3", TotalScore: 300, MemberCount: 2},
		{TeamID: "TM014", TeamName: "O3 Team 2", ChurchID: "CH_O3", TotalScore: 250, MemberCount: 2},
		{TeamID: "TM015", TeamName: "O4 Team 1", ChurchID: "CH_O4", TotalScore: 200, MemberCount: 2},
		{TeamID: "TM016", TeamName: "O4 Team 2", ChurchID: "CH_O4", TotalScore: 150, MemberCount: 2},
	}

	result := handler.calculateDistribution(teams, false)

	// Each priority church must have at least one team in its assigned bucket
	// priorityChurchIDs[i] maps to bucket i (Purple=0, Green=1, Red=2, Yellow=3)
	for i, churchID := range priorityChurchIDs {
		found := false
		for _, team := range result.Superteams[i].Teams {
			if team.ChurchID == churchID {
				found = true
				break
			}
		}
		assert.True(t, found, "Priority church %s should have a team in superteam %s (bucket %d)",
			churchID, superteamNames[i], i)
	}

	// All 16 teams should be distributed
	totalTeams := 0
	for _, st := range result.Superteams {
		totalTeams += st.TeamCount
	}
	assert.Equal(t, 16, totalTeams, "All 16 teams should be distributed")

	visualizeDistribution(t, result)
}

func TestCalculateDistribution_PriorityChurches_WithRefinement(t *testing.T) {
	// Test that refinement doesn't undo priority assignments.
	// Create a scenario where a priority church has only 1 team in its bucket,
	// and refinement might want to swap it out to consolidate another church.
	origPriority := priorityChurchIDs
	priorityChurchIDs = []string{"CH_P1", "CH_P2", "CH_P3", "CH_P4"}
	defer func() { priorityChurchIDs = origPriority }()

	handler := &superteamDistributionHandler{}

	// Priority churches each have exactly 1 team (so refinement can't swap it away
	// without violating the constraint).
	// Other churches are designed to create pressure to consolidate into priority buckets.
	teams := []*sqlc.GetTeamsWithScoresForDistributionRow{
		// Priority churches: 1 team each
		{TeamID: "TM001", TeamName: "P1 Team", ChurchID: "CH_P1", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM002", TeamName: "P2 Team", ChurchID: "CH_P2", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM003", TeamName: "P3 Team", ChurchID: "CH_P3", TotalScore: 1000, MemberCount: 5},
		{TeamID: "TM004", TeamName: "P4 Team", ChurchID: "CH_P4", TotalScore: 1000, MemberCount: 5},
		// Split church with teams that greedy assigns across buckets
		{TeamID: "TM005", TeamName: "Split Team 1", ChurchID: "CH_SPLIT", TotalScore: 900, MemberCount: 4},
		{TeamID: "TM006", TeamName: "Split Team 2", ChurchID: "CH_SPLIT", TotalScore: 850, MemberCount: 4},
		{TeamID: "TM007", TeamName: "Split Team 3", ChurchID: "CH_SPLIT", TotalScore: 800, MemberCount: 4},
		{TeamID: "TM008", TeamName: "Split Team 4", ChurchID: "CH_SPLIT", TotalScore: 750, MemberCount: 4},
		// Filler teams
		{TeamID: "TM009", TeamName: "Fill Team 1", ChurchID: "CH_FILL1", TotalScore: 200, MemberCount: 2},
		{TeamID: "TM010", TeamName: "Fill Team 2", ChurchID: "CH_FILL2", TotalScore: 200, MemberCount: 2},
		{TeamID: "TM011", TeamName: "Fill Team 3", ChurchID: "CH_FILL3", TotalScore: 200, MemberCount: 2},
		{TeamID: "TM012", TeamName: "Fill Team 4", ChurchID: "CH_FILL4", TotalScore: 200, MemberCount: 2},
	}

	result := handler.calculateDistribution(teams, false)

	// Each priority church must still have a team in its assigned bucket after refinement
	for i, churchID := range priorityChurchIDs {
		found := false
		for _, team := range result.Superteams[i].Teams {
			if team.ChurchID == churchID {
				found = true
				break
			}
		}
		assert.True(t, found, "After refinement, priority church %s must still have a team in superteam %s (bucket %d)",
			churchID, superteamNames[i], i)
	}

	// All 12 teams should be distributed
	totalTeams := 0
	for _, st := range result.Superteams {
		totalTeams += st.TeamCount
	}
	assert.Equal(t, 12, totalTeams, "All 12 teams should be distributed")

	visualizeDistribution(t, result)
}

func TestSuperteamDistributionHandler_ChurchesNeverSplit(t *testing.T) {
	handler := &superteamDistributionHandler{}

	// Create a scenario with multiple churches of varying sizes
	// This test verifies that NO church is ever split across superteams
	teams := []*sqlc.GetTeamsWithScoresForDistributionRow{
		// Church A: 5 teams
		{TeamID: "TM01", TeamName: "A1", ChurchID: "CH_A", TotalScore: 5000, MemberCount: 5},
		{TeamID: "TM02", TeamName: "A2", ChurchID: "CH_A", TotalScore: 4000, MemberCount: 4},
		{TeamID: "TM03", TeamName: "A3", ChurchID: "CH_A", TotalScore: 3000, MemberCount: 3},
		{TeamID: "TM04", TeamName: "A4", ChurchID: "CH_A", TotalScore: 2000, MemberCount: 2},
		{TeamID: "TM05", TeamName: "A5", ChurchID: "CH_A", TotalScore: 1000, MemberCount: 1},
		// Church B: 3 teams
		{TeamID: "TM06", TeamName: "B1", ChurchID: "CH_B", TotalScore: 4500, MemberCount: 5},
		{TeamID: "TM07", TeamName: "B2", ChurchID: "CH_B", TotalScore: 3500, MemberCount: 4},
		{TeamID: "TM08", TeamName: "B3", ChurchID: "CH_B", TotalScore: 2500, MemberCount: 3},
		// Church C: 3 teams
		{TeamID: "TM09", TeamName: "C1", ChurchID: "CH_C", TotalScore: 4000, MemberCount: 4},
		{TeamID: "TM10", TeamName: "C2", ChurchID: "CH_C", TotalScore: 3000, MemberCount: 3},
		{TeamID: "TM11", TeamName: "C3", ChurchID: "CH_C", TotalScore: 2000, MemberCount: 2},
		// Church D: 2 teams
		{TeamID: "TM12", TeamName: "D1", ChurchID: "CH_D", TotalScore: 3500, MemberCount: 4},
		{TeamID: "TM13", TeamName: "D2", ChurchID: "CH_D", TotalScore: 2500, MemberCount: 3},
		// Church E: 1 team
		{TeamID: "TM14", TeamName: "E1", ChurchID: "CH_E", TotalScore: 6000, MemberCount: 6},
	}

	result := handler.calculateDistribution(teams, false)

	// Collect which superteam each church appears in
	churchToSuperteams := make(map[string]map[string]bool)
	for _, st := range result.Superteams {
		for _, team := range st.Teams {
			if churchToSuperteams[team.ChurchID] == nil {
				churchToSuperteams[team.ChurchID] = make(map[string]bool)
			}
			churchToSuperteams[team.ChurchID][st.Name] = true
		}
	}

	// Verify that NO church appears in more than one superteam
	for churchID, superteams := range churchToSuperteams {
		assert.Equal(t, 1, len(superteams),
			"Church %s should be in exactly 1 superteam, but is in %d: %v",
			churchID, len(superteams), superteams)
	}

	// All 14 teams should be distributed
	totalTeams := 0
	for _, st := range result.Superteams {
		totalTeams += st.TeamCount
	}
	assert.Equal(t, 14, totalTeams)

	visualizeDistribution(t, result)
}
