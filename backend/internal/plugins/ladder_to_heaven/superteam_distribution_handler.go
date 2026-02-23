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
var superteamNames = []string{"Blue", "Green", "Red", "Yellow"}

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

// calculateDistribution implements a church-preferring greedy bin-packing algorithm.
// Teams are assigned to superteams with a preference for keeping churches together,
// but churches will be split if keeping them together would cause deviation > 10% of the target score.
func (h *superteamDistributionHandler) calculateDistribution(teams []*sqlc.GetTeamsWithScoresForDistributionRow, generateIDs bool) DistributionResponse {
	// Calculate total points and deviation threshold
	var totalPoints int64
	for _, team := range teams {
		totalPoints += team.TotalScore
	}
	targetScore := totalPoints / 4
	deviationThreshold := float64(totalPoints) * 0.05 / 4 // (total/4) * 0.05

	// Sort teams by score (descending) for greedy assignment
	sortedTeams := make([]*sqlc.GetTeamsWithScoresForDistributionRow, len(teams))
	copy(sortedTeams, teams)
	sort.Slice(sortedTeams, func(i, j int) bool {
		return sortedTeams[i].TotalScore > sortedTeams[j].TotalScore
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

	// Track which bucket each church is primarily assigned to
	churchToBucket := make(map[string]int)

	// Greedy assignment with church preference
	for _, team := range sortedTeams {
		targetIdx := -1

		// Check if this church already has teams in a bucket
		if team.ChurchID != "" {
			if bucketIdx, exists := churchToBucket[team.ChurchID]; exists {
				// Check if adding to the church's bucket would exceed deviation threshold
				newScore := buckets[bucketIdx].TotalScore + team.TotalScore
				deviation := float64(newScore) - float64(targetScore)
				if deviation < 0 {
					deviation = -deviation
				}
				if deviation <= deviationThreshold {
					targetIdx = bucketIdx
				}
			}
		}

		// If no church preference or deviation too high, use bucket with lowest score
		if targetIdx == -1 {
			targetIdx = 0
			for i := 1; i < 4; i++ {
				if buckets[i].TotalScore < buckets[targetIdx].TotalScore {
					targetIdx = i
				}
			}
		}

		buckets[targetIdx].Teams = append(buckets[targetIdx].Teams, TeamInfo{
			TeamID:      team.TeamID,
			TeamName:    team.TeamName,
			ChurchID:    team.ChurchID,
			ChurchName:  team.ChurchName,
			TotalScore:  team.TotalScore,
			MemberCount: team.MemberCount,
		})
		buckets[targetIdx].TotalScore += team.TotalScore
		buckets[targetIdx].MemberCount += team.MemberCount

		// Track unique churches and update church-to-bucket mapping
		if team.ChurchID != "" {
			if !containsString(buckets[targetIdx].Churches, team.ChurchID) {
				buckets[targetIdx].Churches = append(buckets[targetIdx].Churches, team.ChurchID)
			}
			// First team from a church determines the preferred bucket
			if _, exists := churchToBucket[team.ChurchID]; !exists {
				churchToBucket[team.ChurchID] = targetIdx
			}
		}
	}

	// Refinement phase: try to reunite split churches by swapping teams
	// Run multiple iterations as new swap opportunities may appear after each pass
	// Stop early if no swaps were made in an iteration
	for range 10 {
		swapsMade := h.refineSplitChurches(buckets, deviationThreshold)
		if swapsMade == 0 {
			break
		}
	}

	// Calculate team counts
	for i := range buckets {
		buckets[i].TeamCount = len(buckets[i].Teams)
	}

	// Recalculate churches list after swaps
	for i := range buckets {
		buckets[i].Churches = h.getUniqueChurches(buckets[i].Teams)
	}

	// Calculate variance for balance metric
	variance := h.calculateVariance(buckets)

	return DistributionResponse{
		Superteams: buckets,
		Variance:   variance,
	}
}

// refineSplitChurches attempts to reunite churches that were split across superteams
// by swapping teams with similar scores. Returns the number of swaps made.
func (h *superteamDistributionHandler) refineSplitChurches(buckets []SuperteamResult, deviationThreshold float64) int {
	swapsMade := 0
	// Build a map of church -> list of (bucketIdx, teamIdx) for all teams
	type teamLocation struct {
		bucketIdx int
		teamIdx   int
		team      *TeamInfo
	}
	churchLocations := make(map[string][]teamLocation)

	for bi := range buckets {
		for ti := range buckets[bi].Teams {
			team := &buckets[bi].Teams[ti]
			if team.ChurchID != "" {
				churchLocations[team.ChurchID] = append(churchLocations[team.ChurchID], teamLocation{
					bucketIdx: bi,
					teamIdx:   ti,
					team:      team,
				})
			}
		}
	}

	// Sort church IDs for deterministic processing order
	// Process smaller churches first (more likely to be reunitable)
	churchIDs := make([]string, 0, len(churchLocations))
	for churchID := range churchLocations {
		churchIDs = append(churchIDs, churchID)
	}
	sort.Slice(churchIDs, func(i, j int) bool {
		// Sort by number of teams (ascending), then by church ID for stability
		if len(churchLocations[churchIDs[i]]) != len(churchLocations[churchIDs[j]]) {
			return len(churchLocations[churchIDs[i]]) < len(churchLocations[churchIDs[j]])
		}
		return churchIDs[i] < churchIDs[j]
	})

	// Find split churches (teams in more than one bucket)
	for _, churchID := range churchIDs {
		locations := churchLocations[churchID]
		if len(locations) <= 1 {
			continue
		}

		// Find which buckets this church is in
		bucketSet := make(map[int][]int) // bucketIdx -> teamIndices
		for _, loc := range locations {
			bucketSet[loc.bucketIdx] = append(bucketSet[loc.bucketIdx], loc.teamIdx)
		}

		if len(bucketSet) <= 1 {
			continue // All teams already in same bucket
		}

		// Try to consolidate: find the bucket with most teams from this church
		var targetBucket int
		maxTeams := 0
		for bi, indices := range bucketSet {
			if len(indices) > maxTeams {
				maxTeams = len(indices)
				targetBucket = bi
			}
		}

		// Try to move teams from other buckets to target bucket via swaps
		for sourceBucket, teamIndices := range bucketSet {
			if sourceBucket == targetBucket {
				continue
			}

			for _, sourceTeamIdx := range teamIndices {
				sourceTeam := &buckets[sourceBucket].Teams[sourceTeamIdx]

				// Find a team in targetBucket (from a different church) to swap with
				bestSwapIdx := -1
				bestScoreDiff := int64(1<<62 - 1)

				for ti := range buckets[targetBucket].Teams {
					candidate := &buckets[targetBucket].Teams[ti]

					// Don't swap with teams from the same church we're trying to consolidate
					if candidate.ChurchID == churchID {
						continue
					}

					// Calculate score difference
					scoreDiff := sourceTeam.TotalScore - candidate.TotalScore
					if scoreDiff < 0 {
						scoreDiff = -scoreDiff
					}

					// Check if this swap would keep both buckets within threshold
					newSourceScore := buckets[sourceBucket].TotalScore - sourceTeam.TotalScore + candidate.TotalScore
					newTargetScore := buckets[targetBucket].TotalScore - candidate.TotalScore + sourceTeam.TotalScore

					// Calculate total for threshold check
					var totalScore int64
					for _, b := range buckets {
						totalScore += b.TotalScore
					}
					targetScore := totalScore / 4

					sourceDeviation := float64(newSourceScore) - float64(targetScore)
					if sourceDeviation < 0 {
						sourceDeviation = -sourceDeviation
					}
					targetDeviation := float64(newTargetScore) - float64(targetScore)
					if targetDeviation < 0 {
						targetDeviation = -targetDeviation
					}

					// Only consider swap if it keeps both buckets within threshold
					if sourceDeviation <= deviationThreshold && targetDeviation <= deviationThreshold {
						if scoreDiff < bestScoreDiff {
							bestScoreDiff = scoreDiff
							bestSwapIdx = ti
						}
					}
				}

				// Perform the swap if we found a valid candidate
				if bestSwapIdx >= 0 {
					h.swapTeams(buckets, sourceBucket, sourceTeamIdx, targetBucket, bestSwapIdx)
					swapsMade++
					// After swap, indices may have shifted - break and let next iteration handle remaining
					break
				}
			}
		}
	}

	return swapsMade
}

// swapTeams swaps two teams between buckets
func (h *superteamDistributionHandler) swapTeams(buckets []SuperteamResult, bucket1, idx1, bucket2, idx2 int) {
	team1 := buckets[bucket1].Teams[idx1]
	team2 := buckets[bucket2].Teams[idx2]

	// Update scores
	buckets[bucket1].TotalScore = buckets[bucket1].TotalScore - team1.TotalScore + team2.TotalScore
	buckets[bucket2].TotalScore = buckets[bucket2].TotalScore - team2.TotalScore + team1.TotalScore

	// Update member counts
	buckets[bucket1].MemberCount = buckets[bucket1].MemberCount - team1.MemberCount + team2.MemberCount
	buckets[bucket2].MemberCount = buckets[bucket2].MemberCount - team2.MemberCount + team1.MemberCount

	// Swap the teams
	buckets[bucket1].Teams[idx1] = team2
	buckets[bucket2].Teams[idx2] = team1
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
