package ladder_to_heaven

import (
	"context"
	"fmt"
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
var priorityChurchIDs = []string{
	"CH01KC9E89Q9K4Q1M4MNJC99CQKJ", // Oslo Follo -> Purple
	"CH01KC9E885FZ3DXF7NDQCJJT4C9", // Exter -> Green
	"CH01KC9E8BHK3C9VSD3ABGFPDWZF", // Sveits -> Red
	"CH01KC9E8D1FVXY513F3K59H2D47", // København -> Yellow
}

// distributionQuerier defines the database operations needed by the distribution handler.
type distributionQuerier interface {
	GetTeamsWithScoresForDistribution(ctx context.Context, projectID string) ([]*sqlc.GetTeamsWithScoresForDistributionRow, error)
	GetTeamsWithScoresAndAttendingForDistribution(ctx context.Context, arg sqlc.GetTeamsWithScoresAndAttendingForDistributionParams) ([]*sqlc.GetTeamsWithScoresAndAttendingForDistributionRow, error)
	GetUserIDsByEventID(ctx context.Context, eventID string) ([]string, error)
	GetTeamsByIDs(ctx context.Context, ids []string) ([]*sqlc.GetTeamsByIDsRow, error)
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
	TeamID         string `json:"team_id"`
	TeamName       string `json:"team_name"`
	ChurchID       string `json:"church_id"`
	ChurchName     string `json:"church_name"`
	TotalScore     int64  `json:"total_score"`
	MemberCount    int32  `json:"member_count"`
	AttendingCount int32  `json:"attending_count"`
}

// SuperteamResult represents a superteam with its assigned teams.
type SuperteamResult struct {
	SuperTeamID    string     `json:"super_team_id"`
	Name           string     `json:"name"`
	TotalScore     int64      `json:"total_score"`
	TeamCount      int        `json:"team_count"`
	MemberCount    int32      `json:"member_count"`
	AttendingCount int32      `json:"attending_count"`
	Teams          []TeamInfo `json:"teams"`
	Churches       []string   `json:"churches"`
}

// DistributionResponse is the response for preview/execute endpoints.
type DistributionResponse struct {
	Superteams []SuperteamResult `json:"superteams"`
	Variance   float64           `json:"variance"`
}

// DistributeRequest is the request body for the distribute endpoint.
type DistributeRequest struct {
	ProjectID  string                     `json:"project_id" binding:"required"`
	EventID    string                     `json:"event_id"`   // Optional: if provided, fetch attending users from user_events
	Superteams []SuperteamAssignmentInput `json:"superteams"` // Optional: if provided, use this distribution instead of recalculating
}

// SuperteamAssignmentInput represents a superteam assignment from the frontend.
type SuperteamAssignmentInput struct {
	Name    string   `json:"name" binding:"required"`
	TeamIDs []string `json:"team_ids" binding:"required"`
}

// DistributionWeights configures the relative importance of balancing factors.
// Higher weight means more importance in the distribution algorithm.
type DistributionWeights struct {
	Attending float64 // Weight for attending member balance (default 0.6)
	Score     float64 // Weight for score balance (default 0.3)
	TeamCount float64 // Weight for team count balance (default 0.1)
}

// DefaultDistributionWeights returns the default weights for distribution.
func DefaultDistributionWeights() DistributionWeights {
	return DistributionWeights{
		Attending: 0.6,
		Score:     0.3,
		TeamCount: 0.1,
	}
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

	eventID := c.Query("event_id") // Optional

	if !h.checkAuth(c) {
		return
	}

	querier := h.getQuerier()

	var result DistributionResponse

	if eventID != "" {
		// Use attending-aware distribution
		attendingUserIDs, err := querier.GetUserIDsByEventID(ctx, eventID)
		if err != nil {
			slog.Error("superteam_distribution: failed to get attending users", "error", err, "event_id", eventID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve attending users"})
			return
		}

		teams, err := querier.GetTeamsWithScoresAndAttendingForDistribution(ctx, sqlc.GetTeamsWithScoresAndAttendingForDistributionParams{
			AttendingUserIds: attendingUserIDs,
			ProjectID:        projectID,
		})
		if err != nil {
			slog.Error("superteam_distribution: failed to get teams with attending", "error", err, "project_id", projectID)
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

		if len(attendingUserIDs) == 0 {
			slog.Warn("superteam_distribution: no attending users found, falling back to score-based distribution", "event_id", eventID)
		}

		result = h.calculateDistributionWithAttending(teams, false)
	} else {
		// Fall back to current score-based distribution
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

		result = h.calculateDistribution(teams, false)
	}

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

	var result DistributionResponse
	var err error

	// If superteams are provided, use them directly instead of recalculating
	if len(req.Superteams) > 0 {
		result, err = h.buildDistributionFromInput(ctx, querier, req.Superteams, req.ProjectID)
		if err != nil {
			slog.Error("superteam_distribution: failed to build distribution from input", "error", err, "project_id", req.ProjectID)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else if req.EventID != "" {
		// Use attending-aware distribution
		attendingUserIDs, err := querier.GetUserIDsByEventID(ctx, req.EventID)
		if err != nil {
			slog.Error("superteam_distribution: failed to get attending users", "error", err, "event_id", req.EventID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve attending users"})
			return
		}

		teams, err := querier.GetTeamsWithScoresAndAttendingForDistribution(ctx, sqlc.GetTeamsWithScoresAndAttendingForDistributionParams{
			AttendingUserIds: attendingUserIDs,
			ProjectID:        req.ProjectID,
		})
		if err != nil {
			slog.Error("superteam_distribution: failed to get teams with attending", "error", err, "project_id", req.ProjectID)
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

		if len(attendingUserIDs) == 0 {
			slog.Warn("superteam_distribution: no attending users found, falling back to score-based distribution", "event_id", req.EventID)
		}

		result = h.calculateDistributionWithAttending(teams, true)
	} else {
		// Fall back to current score-based distribution
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

		result = h.calculateDistribution(teams, true)
	}

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

	// Calculate total team count from result
	totalTeamCount := 0
	for _, st := range result.Superteams {
		totalTeamCount += st.TeamCount
	}
	slog.Info("superteam_distribution: distribution executed successfully",
		"project_id", req.ProjectID,
		"team_count", totalTeamCount,
	)

	c.JSON(http.StatusOK, result)
}

// churchGroup holds all teams from a single church for batch assignment
type churchGroup struct {
	ChurchID       string
	ChurchName     string
	Teams          []TeamInfo
	TotalScore     int64
	TeamCount      int
	AttendingCount int32
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

	// Refinement phase: try swapping churches to improve score balance
	h.refineDistributionByScore(buckets, churchGroups)

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

// calculateDistributionWithAttending implements an attending-aware church-cohesive distribution.
// It prioritizes balancing attending members across superteams while maintaining the hard
// constraint that churches are never split.
func (h *superteamDistributionHandler) calculateDistributionWithAttending(teams []*sqlc.GetTeamsWithScoresAndAttendingForDistributionRow, generateIDs bool) DistributionResponse {
	// Group teams by church
	churchGroups := make(map[string]*churchGroup)
	var totalPoints int64
	var totalAttending int32
	for _, team := range teams {
		totalPoints += team.TotalScore
		totalAttending += team.AttendingCount
		if churchGroups[team.ChurchID] == nil {
			churchGroups[team.ChurchID] = &churchGroup{
				ChurchID:   team.ChurchID,
				ChurchName: team.ChurchName,
				Teams:      []TeamInfo{},
			}
		}
		churchGroups[team.ChurchID].Teams = append(churchGroups[team.ChurchID].Teams, TeamInfo{
			TeamID:         team.TeamID,
			TeamName:       team.TeamName,
			ChurchID:       team.ChurchID,
			ChurchName:     team.ChurchName,
			TotalScore:     team.TotalScore,
			MemberCount:    team.MemberCount,
			AttendingCount: team.AttendingCount,
		})
		churchGroups[team.ChurchID].TotalScore += team.TotalScore
		churchGroups[team.ChurchID].TeamCount++
		churchGroups[team.ChurchID].AttendingCount += team.AttendingCount
	}

	// Convert to slice and sort churches by attending count (descending)
	// This ensures larger churches are placed first
	churches := make([]*churchGroup, 0, len(churchGroups))
	for _, cg := range churchGroups {
		churches = append(churches, cg)
	}
	sort.Slice(churches, func(i, j int) bool {
		return churches[i].AttendingCount > churches[j].AttendingCount
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
	avgAttending := float64(totalAttending) / 4.0

	// Get distribution weights
	weights := DefaultDistributionWeights()

	// Check if we have any attending users
	hasAttending := totalAttending > 0

	// Assign remaining churches using weighted multi-objective balancing
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
			newAttending := buckets[i].AttendingCount + cg.AttendingCount

			var combinedImbalance float64

			if hasAttending {
				// Normalize relative to averages and apply weights
				scoreImbalance := float64(newScore) / avgScore
				teamCountImbalance := float64(newTeamCount) / avgTeamCount
				attendingImbalance := float64(newAttending) / avgAttending

				// Weighted combined imbalance - lower is better
				combinedImbalance = weights.Attending*attendingImbalance +
					weights.Score*scoreImbalance +
					weights.TeamCount*teamCountImbalance
			} else {
				// Fall back to score + team count when no attending data
				scoreImbalance := float64(newScore) / avgScore
				teamCountImbalance := float64(newTeamCount) / avgTeamCount
				combinedImbalance = scoreImbalance + teamCountImbalance
			}

			if combinedImbalance < bestImbalance {
				bestImbalance = combinedImbalance
				targetIdx = i
			}
		}

		h.assignChurchToBucket(cg, &buckets[targetIdx])
		assignedChurches[cg.ChurchID] = true
	}

	// Refinement phase: try swapping churches to improve score balance
	h.refineDistributionByScore(buckets, churchGroups)

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
	bucketAttendingCounts := make([]int32, len(buckets))
	for i := range buckets {
		bucketScores[i] = buckets[i].TotalScore
		bucketTeamCounts[i] = buckets[i].TeamCount
		bucketAttendingCounts[i] = buckets[i].AttendingCount
	}
	slog.Info("superteam_distribution: attending-aware distribution calculated",
		"total_points", totalPoints,
		"total_attending", totalAttending,
		"bucket_scores", bucketScores,
		"bucket_team_counts", bucketTeamCounts,
		"bucket_attending_counts", bucketAttendingCounts,
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
	bucket.AttendingCount += cg.AttendingCount
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

// isPriorityChurchInAssignedBucket checks if a church is a priority church in its designated bucket.
func isPriorityChurchInAssignedBucket(churchID string, bucketIdx int) bool {
	for i, priorityID := range priorityChurchIDs {
		if priorityID == churchID && i == bucketIdx {
			return true
		}
	}
	return false
}

// abs64 returns the absolute value of an int64.
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// removeString removes a string from a slice and returns the new slice.
func removeString(slice []string, str string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != str {
			result = append(result, s)
		}
	}
	return result
}

// removeChurchTeams removes all teams belonging to a church from a team slice.
func removeChurchTeams(teams []TeamInfo, churchID string) []TeamInfo {
	result := make([]TeamInfo, 0, len(teams))
	for _, team := range teams {
		if team.ChurchID != churchID {
			result = append(result, team)
		}
	}
	return result
}

// swapChurchesBetweenBuckets swaps two churches between their respective buckets.
func (h *superteamDistributionHandler) swapChurchesBetweenBuckets(bucket1, bucket2 *SuperteamResult, church1, church2 *churchGroup) {
	// Remove church1 from bucket1
	bucket1.Teams = removeChurchTeams(bucket1.Teams, church1.ChurchID)
	bucket1.TotalScore -= church1.TotalScore
	bucket1.AttendingCount -= church1.AttendingCount
	bucket1.Churches = removeString(bucket1.Churches, church1.ChurchID)
	for _, team := range church1.Teams {
		bucket1.MemberCount -= team.MemberCount
	}

	// Remove church2 from bucket2
	bucket2.Teams = removeChurchTeams(bucket2.Teams, church2.ChurchID)
	bucket2.TotalScore -= church2.TotalScore
	bucket2.AttendingCount -= church2.AttendingCount
	bucket2.Churches = removeString(bucket2.Churches, church2.ChurchID)
	for _, team := range church2.Teams {
		bucket2.MemberCount -= team.MemberCount
	}

	// Add church1 to bucket2
	bucket2.Teams = append(bucket2.Teams, church1.Teams...)
	bucket2.TotalScore += church1.TotalScore
	bucket2.AttendingCount += church1.AttendingCount
	bucket2.Churches = append(bucket2.Churches, church1.ChurchID)
	for _, team := range church1.Teams {
		bucket2.MemberCount += team.MemberCount
	}

	// Add church2 to bucket1
	bucket1.Teams = append(bucket1.Teams, church2.Teams...)
	bucket1.TotalScore += church2.TotalScore
	bucket1.AttendingCount += church2.AttendingCount
	bucket1.Churches = append(bucket1.Churches, church2.ChurchID)
	for _, team := range church2.Teams {
		bucket1.MemberCount += team.MemberCount
	}
}

// buildDistributionFromInput creates a DistributionResponse from user-provided superteam assignments.
// It validates that all team IDs exist and belong to the specified project.
func (h *superteamDistributionHandler) buildDistributionFromInput(ctx context.Context, querier distributionQuerier, assignments []SuperteamAssignmentInput, projectID string) (DistributionResponse, error) {
	// Validate we have exactly 4 superteams
	if len(assignments) != 4 {
		return DistributionResponse{}, fmt.Errorf("expected 4 superteams, got %d", len(assignments))
	}

	// Validate superteam names
	nameSet := make(map[string]bool)
	for _, name := range superteamNames {
		nameSet[name] = true
	}
	for _, assignment := range assignments {
		if !nameSet[assignment.Name] {
			return DistributionResponse{}, fmt.Errorf("invalid superteam name: %s", assignment.Name)
		}
	}

	// Collect all team IDs and check for duplicates
	allTeamIDs := make([]string, 0)
	seenTeamIDs := make(map[string]bool)
	for _, assignment := range assignments {
		for _, teamID := range assignment.TeamIDs {
			if seenTeamIDs[teamID] {
				return DistributionResponse{}, fmt.Errorf("team %s appears in multiple superteams", teamID)
			}
			seenTeamIDs[teamID] = true
			allTeamIDs = append(allTeamIDs, teamID)
		}
	}

	// Fetch all teams to validate they exist and belong to this project
	teams, err := querier.GetTeamsByIDs(ctx, allTeamIDs)
	if err != nil {
		return DistributionResponse{}, fmt.Errorf("failed to fetch teams: %w", err)
	}

	// Create a map for quick lookup
	teamMap := make(map[string]*sqlc.GetTeamsByIDsRow)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	// Validate all teams exist and belong to this project
	for _, teamID := range allTeamIDs {
		team, exists := teamMap[teamID]
		if !exists {
			return DistributionResponse{}, fmt.Errorf("team %s not found", teamID)
		}
		if team.ProjectID != projectID {
			return DistributionResponse{}, fmt.Errorf("team %s does not belong to project %s", teamID, projectID)
		}
	}

	// Build the result
	buckets := make([]SuperteamResult, 4)
	for i, assignment := range assignments {
		buckets[i] = SuperteamResult{
			SuperTeamID: ulid.NewSuperTeamID(),
			Name:        assignment.Name,
			Teams:       make([]TeamInfo, 0, len(assignment.TeamIDs)),
			Churches:    []string{},
		}

		for _, teamID := range assignment.TeamIDs {
			team := teamMap[teamID]
			buckets[i].Teams = append(buckets[i].Teams, TeamInfo{
				TeamID:   team.ID,
				TeamName: team.Name,
				// Note: church info not available from GetTeamsByIDs, but not needed for storage
			})
		}

		buckets[i].TeamCount = len(buckets[i].Teams)
	}

	// Note: Variance is 0 since we don't have score info - not needed for execute
	return DistributionResponse{
		Superteams: buckets,
		Variance:   0,
	}, nil
}

// refineDistributionByScore attempts to improve score balance by swapping churches between buckets.
// It iteratively finds the highest and lowest score buckets and tries swapping churches to reduce variance.
func (h *superteamDistributionHandler) refineDistributionByScore(buckets []SuperteamResult, churchGroups map[string]*churchGroup) {
	const maxIterations = 10

	for iter := 0; iter < maxIterations; iter++ {
		// Find highest and lowest score buckets
		highIdx, lowIdx := 0, 0
		for i := 1; i < 4; i++ {
			if buckets[i].TotalScore > buckets[highIdx].TotalScore {
				highIdx = i
			}
			if buckets[i].TotalScore < buckets[lowIdx].TotalScore {
				lowIdx = i
			}
		}

		if highIdx == lowIdx {
			break // Already balanced
		}

		currentDiff := buckets[highIdx].TotalScore - buckets[lowIdx].TotalScore
		improved := false

		// Try swapping churches to reduce imbalance
		for _, highChurchID := range buckets[highIdx].Churches {
			if isPriorityChurchInAssignedBucket(highChurchID, highIdx) {
				continue
			}
			highChurch := churchGroups[highChurchID]

			for _, lowChurchID := range buckets[lowIdx].Churches {
				if isPriorityChurchInAssignedBucket(lowChurchID, lowIdx) {
					continue
				}
				lowChurch := churchGroups[lowChurchID]

				// Calculate new scores after swap
				newHighScore := buckets[highIdx].TotalScore - highChurch.TotalScore + lowChurch.TotalScore
				newLowScore := buckets[lowIdx].TotalScore - lowChurch.TotalScore + highChurch.TotalScore
				newDiff := abs64(newHighScore - newLowScore)

				if newDiff < currentDiff {
					// Perform the swap
					h.swapChurchesBetweenBuckets(&buckets[highIdx], &buckets[lowIdx], highChurch, lowChurch)
					improved = true
					break
				}
			}
			if improved {
				break
			}
		}

		if !improved {
			break
		}
	}
}
