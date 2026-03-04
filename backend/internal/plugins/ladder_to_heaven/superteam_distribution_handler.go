package ladder_to_heaven

import (
	"context"
	"log/slog"
	"net/http"
	"sort"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/gin-gonic/gin"
)

// Fixed superteam names for distribution
var superteamNames = []string{"Purple", "Green", "Red", "Yellow"}

// Priority churches that must each be in a different superteam.
// Order maps to superteam order: [0]=Purple, [1]=Green, [2]=Red, [3]=Yellow.
var priorityChurchIDs = []string{"CH_PLACEHOLDER_1", "CH_PLACEHOLDER_2", "CH_PLACEHOLDER_3", "CH_PLACEHOLDER_4"}

// distributionQuerier defines the database operations needed by the distribution handler.
type distributionQuerier interface {
	GetTeamsWithScoresForDistribution(ctx context.Context, projectID string) ([]*sqlc.GetTeamsWithScoresForDistributionRow, error)
	ClearSuperTeamAssignmentsForProject(ctx context.Context, projectID string) error
	DeleteSuperTeamsByProjectID(ctx context.Context, projectID string) error
	CreateSuperTeam(ctx context.Context, arg sqlc.CreateSuperTeamParams) (*sqlc.SuperTeam, error)
	AssignTeamToSuperTeam(ctx context.Context, arg sqlc.AssignTeamToSuperTeamParams) error
}

// superteamDistributionHandler handles superteam distribution requests.
type superteamDistributionHandler struct {
	db        *database.DB
	cache     *cache.CacheWithRegistry
	jwtConfig config.JWTConfig

	// For testing - when set, overrides the default implementation
	testQuerier distributionQuerier
}

// TeamInfo represents a team with its score for distribution.
type TeamInfo struct {
	TeamID      string `json:"team_id"`
	TeamName    string `json:"team_name"`
	ChurchID    string `json:"church_id"`
	ChurchName  string `json:"church_name"`
	TotalScore  int64  `json:"total_score"`
	MemberCount int32  `json:"member_count"`
}

// SuperteamResult represents a superteam with its assigned teams.
type SuperteamResult struct {
	SuperTeamID string     `json:"super_team_id"`
	Name        string     `json:"name"`
	TotalScore  int64      `json:"total_score"`
	TeamCount   int        `json:"team_count"`
	MemberCount int32      `json:"member_count"`
	Teams       []TeamInfo `json:"teams"`
	Churches    []string   `json:"churches"`
}

// DistributionResponse is the response for preview/execute endpoints.
type DistributionResponse struct {
	Superteams []SuperteamResult `json:"superteams"`
	Variance   float64           `json:"variance"`
}

// DistributeRequest is the request body for the distribute endpoint.
type DistributeRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
}

// allowedDistributionRoles are the roles that can use distribution endpoints.
var allowedDistributionRoles = []string{"admin", "superadmin"}

// preview handles GET requests to preview the distribution without making changes.
func (h *superteamDistributionHandler) preview(c *gin.Context) {
	ctx := c.Request.Context()

	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	if !h.checkAuth(c) {
		return
	}

	querier := h.getQuerier()

	// Get teams with scores
	teams, err := querier.GetTeamsWithScoresForDistribution(ctx, projectID)
	if err != nil {
		slog.Error("superteam_distribution: failed to get teams", "error", err, "project_id", projectID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve teams"})
		return
	}

	if len(teams) == 0 {
		c.JSON(http.StatusOK, DistributionResponse{
			Superteams: []SuperteamResult{},
			Variance:   0,
		})
		return
	}

	// Calculate distribution (preview only, no IDs assigned)
	result := h.calculateDistribution(teams, false)

	c.JSON(http.StatusOK, result)
}

// handle handles POST requests to execute the distribution.
func (h *superteamDistributionHandler) handle(c *gin.Context) {
	ctx := c.Request.Context()

	var req DistributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if !h.checkAuth(c) {
		return
	}

	querier := h.getQuerier()

	// Get teams with scores
	teams, err := querier.GetTeamsWithScoresForDistribution(ctx, req.ProjectID)
	if err != nil {
		slog.Error("superteam_distribution: failed to get teams", "error", err, "project_id", req.ProjectID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve teams"})
		return
	}

	if len(teams) == 0 {
		c.JSON(http.StatusOK, DistributionResponse{
			Superteams: []SuperteamResult{},
			Variance:   0,
		})
		return
	}

	// Calculate distribution with IDs
	result := h.calculateDistribution(teams, true)

	// Execute the distribution in a transaction
	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		slog.Error("superteam_distribution: failed to begin transaction", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to execute distribution"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txQuerier := h.db.Queries.WithTx(tx)

	// Clear existing assignments
	if err := txQuerier.ClearSuperTeamAssignmentsForProject(ctx, req.ProjectID); err != nil {
		slog.Error("superteam_distribution: failed to clear assignments", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to execute distribution"})
		return
	}

	// Delete existing superteams
	if err := txQuerier.DeleteSuperTeamsByProjectID(ctx, req.ProjectID); err != nil {
		slog.Error("superteam_distribution: failed to delete superteams", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to execute distribution"})
		return
	}

	// Create new superteams and assign teams
	for i := range result.Superteams {
		st := &result.Superteams[i]

		// Create the superteam
		_, err := txQuerier.CreateSuperTeam(ctx, sqlc.CreateSuperTeamParams{
			ID:        st.SuperTeamID,
			ProjectID: req.ProjectID,
			Name:      st.Name,
		})
		if err != nil {
			slog.Error("superteam_distribution: failed to create superteam", "error", err, "name", st.Name)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to execute distribution"})
			return
		}

		// Assign teams to this superteam
		for _, team := range st.Teams {
			superTeamID := st.SuperTeamID
			if err := txQuerier.AssignTeamToSuperTeam(ctx, sqlc.AssignTeamToSuperTeamParams{
				SuperTeamID: &superTeamID,
				TeamID:      team.TeamID,
			}); err != nil {
				slog.Error("superteam_distribution: failed to assign team", "error", err, "team_id", team.TeamID)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to execute distribution"})
				return
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		slog.Error("superteam_distribution: failed to commit transaction", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to execute distribution"})
		return
	}

	// Clear cache to ensure fresh data for leaderboards
	h.cache.Clear()

	slog.Info("superteam_distribution: distribution executed successfully",
		"project_id", req.ProjectID,
		"team_count", len(teams),
	)

	c.JSON(http.StatusOK, result)
}

// churchGroup holds all teams from a single church for batch assignment
type churchGroup struct {
	ChurchID   string
	ChurchName string
	Teams      []TeamInfo
	TotalScore int64
	TeamCount  int
}

// calculateDistribution implements a church-cohesive greedy bin-packing algorithm.
// Churches are assigned as whole units to superteams, considering both score and
// team count balance. This ensures teams from the same church are always together.
func (h *superteamDistributionHandler) calculateDistribution(teams []*sqlc.GetTeamsWithScoresForDistributionRow, generateIDs bool) DistributionResponse {
	// Group teams by church
	churchGroups := make(map[string]*churchGroup)
	var totalPoints int64
	for _, team := range teams {
		totalPoints += team.TotalScore
		if churchGroups[team.ChurchID] == nil {
			churchGroups[team.ChurchID] = &churchGroup{
				ChurchID:   team.ChurchID,
				ChurchName: team.ChurchName,
				Teams:      []TeamInfo{},
			}
		}
		churchGroups[team.ChurchID].Teams = append(churchGroups[team.ChurchID].Teams, TeamInfo{
			TeamID:      team.TeamID,
			TeamName:    team.TeamName,
			ChurchID:    team.ChurchID,
			ChurchName:  team.ChurchName,
			TotalScore:  team.TotalScore,
			MemberCount: team.MemberCount,
		})
		churchGroups[team.ChurchID].TotalScore += team.TotalScore
		churchGroups[team.ChurchID].TeamCount++
	}

	// Convert to slice and sort churches by total score (descending)
	churches := make([]*churchGroup, 0, len(churchGroups))
	for _, cg := range churchGroups {
		churches = append(churches, cg)
	}
	sort.Slice(churches, func(i, j int) bool {
		return churches[i].TotalScore > churches[j].TotalScore
	})

	// Initialize 4 superteam buckets
	buckets := make([]SuperteamResult, 4)
	for i := 0; i < 4; i++ {
		buckets[i] = SuperteamResult{
			Name:       superteamNames[i],
			Teams:      []TeamInfo{},
			Churches:   []string{},
			TotalScore: 0,
		}
		if generateIDs {
			buckets[i].SuperTeamID = ulid.NewSuperTeamID()
		}
	}

	// Track which churches have been assigned
	assignedChurches := make(map[string]bool)

	// Pre-seed priority churches to their designated buckets
	for i, churchID := range priorityChurchIDs {
		if cg, exists := churchGroups[churchID]; exists {
			h.assignChurchToBucket(cg, &buckets[i])
			assignedChurches[churchID] = true
		}
	}

	// Pre-calculate averages for normalization
	avgTeamCount := float64(len(teams)) / 4.0
	avgScore := float64(totalPoints) / 4.0

	// Assign remaining churches, considering both score and team count balance
	for _, cg := range churches {
		if assignedChurches[cg.ChurchID] {
			continue
		}

		// Find bucket with best combined balance after adding this church
		targetIdx := 0
		bestImbalance := float64(1 << 62)
		for i := 0; i < 4; i++ {
			// Calculate what the bucket would look like after adding this church
			newScore := buckets[i].TotalScore + cg.TotalScore
			newTeamCount := len(buckets[i].Teams) + cg.TeamCount

			// Normalize relative to averages
			scoreImbalance := float64(newScore) / avgScore
			teamCountImbalance := float64(newTeamCount) / avgTeamCount

			// Combined imbalance metric - lower is better
			combinedImbalance := scoreImbalance + teamCountImbalance

			if combinedImbalance < bestImbalance {
				bestImbalance = combinedImbalance
				targetIdx = i
			}
		}

		h.assignChurchToBucket(cg, &buckets[targetIdx])
		assignedChurches[cg.ChurchID] = true
	}

	// Calculate team counts
	for i := range buckets {
		buckets[i].TeamCount = len(buckets[i].Teams)
	}

	// Recalculate churches list for consistency
	for i := range buckets {
		buckets[i].Churches = h.getUniqueChurches(buckets[i].Teams)
	}

	// Calculate variance for balance metric
	variance := h.calculateVariance(buckets)

	// Log the distribution result for visibility
	bucketScores := make([]int64, len(buckets))
	bucketTeamCounts := make([]int, len(buckets))
	for i := range buckets {
		bucketScores[i] = buckets[i].TotalScore
		bucketTeamCounts[i] = buckets[i].TeamCount
	}
	slog.Info("superteam_distribution: distribution calculated",
		"total_points", totalPoints,
		"bucket_scores", bucketScores,
		"bucket_team_counts", bucketTeamCounts,
		"variance", variance,
	)

	return DistributionResponse{
		Superteams: buckets,
		Variance:   variance,
	}
}

// assignChurchToBucket adds all teams from a church to a bucket
func (h *superteamDistributionHandler) assignChurchToBucket(cg *churchGroup, bucket *SuperteamResult) {
	bucket.Teams = append(bucket.Teams, cg.Teams...)
	bucket.TotalScore += cg.TotalScore
	for _, team := range cg.Teams {
		bucket.MemberCount += team.MemberCount
	}
	if !containsString(bucket.Churches, cg.ChurchID) {
		bucket.Churches = append(bucket.Churches, cg.ChurchID)
	}
}

// getUniqueChurches returns a list of unique church IDs from a list of teams
func (h *superteamDistributionHandler) getUniqueChurches(teams []TeamInfo) []string {
	seen := make(map[string]bool)
	var churches []string
	for _, team := range teams {
		if team.ChurchID != "" && !seen[team.ChurchID] {
			seen[team.ChurchID] = true
			churches = append(churches, team.ChurchID)
		}
	}
	return churches
}

// calculateVariance calculates the variance of scores across superteams.
func (h *superteamDistributionHandler) calculateVariance(buckets []SuperteamResult) float64 {
	if len(buckets) == 0 {
		return 0
	}

	// Calculate mean
	var total int64
	for _, b := range buckets {
		total += b.TotalScore
	}
	mean := float64(total) / float64(len(buckets))

	// Calculate variance
	var sumSquaredDiff float64
	for _, b := range buckets {
		diff := float64(b.TotalScore) - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(buckets))
}

// checkAuth verifies the user has admin permissions.
func (h *superteamDistributionHandler) checkAuth(c *gin.Context) bool {
	userIDValue, exists := c.Get("user_id")
	if !exists || userIDValue == nil {
		slog.Warn("superteam_distribution: user not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return false
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		slog.Warn("superteam_distribution: invalid user_id in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return false
	}

	userRolesValue, exists := c.Get("user_roles")
	if !exists || userRolesValue == nil {
		slog.Warn("superteam_distribution: no roles found for user", "user_id", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return false
	}
	userRoles, ok := userRolesValue.([]string)
	if !ok {
		slog.Warn("superteam_distribution: invalid user_roles in context", "user_id", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return false
	}

	if !hasDistributionRole(userRoles) {
		slog.Warn("superteam_distribution: user lacks required role",
			"user_id", userID,
			"user_roles", userRoles,
			"required_any", allowedDistributionRoles,
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return false
	}

	return true
}

// getQuerier returns the querier to use (test mock or real implementation).
func (h *superteamDistributionHandler) getQuerier() distributionQuerier {
	if h.testQuerier != nil {
		return h.testQuerier
	}
	return h.db.Queries
}

// hasDistributionRole checks if the user has at least one of the allowed roles.
func hasDistributionRole(userRoles []string) bool {
	for _, userRole := range userRoles {
		for _, allowedRole := range allowedDistributionRoles {
			if userRole == allowedRole {
				return true
			}
		}
	}
	return false
}

// containsString checks if a string slice contains a specific string.
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
