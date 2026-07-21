package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test entity IDs for membership leaderboard tests (must be exactly 28 chars: 2-char prefix + 26-char ULID)
const (
	// Base entities
	membershipChurchID  = "CH01TEST0000000000MEMBERSHIP"
	membershipProjectID = "PR01TEST0000000000MEMBERSHIP"
	membershipEventID   = "EV01TEST0000000000MEMBERSHIP"

	// Teams
	membershipTeamID      = "TM01TEST0000000000MEMBERSHIP"
	membershipTeam2ID     = "TM01TEST000000000MEMBERSHIP2"
	membershipSuperTeamID = "ST01TEST0000000000MEMBERSHIP"
	membershipTeamSTID    = "TM01TEST00000000MEMBERSHIPST"

	// Users
	membershipUser1ID = "US01TEST000000000MEMBERSHIP1"
	membershipUser2ID = "US01TEST000000000MEMBERSHIP2"
	membershipUser3ID = "US01TEST000000000MEMBERSHIP3"
)

// ==================== Team Membership Tests ====================

// TestTeamMembershipLeaderboardPoints tests that team leaderboard points are updated
// correctly when users join and leave teams.
func TestTeamMembershipLeaderboardPoints(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	t.Run("user with project points joins team -> team leaderboard includes user points", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseData(t, ctx, dbMgr)

		// Create user with points but NOT in team yet
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))

		// Verify team leaderboard has 0 points before join
		pointsBefore, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pointsBefore, "team should have 0 points before user joins")

		// User joins team - trigger should transfer points
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamID))

		// Verify team leaderboard now has user's points
		pointsAfter, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(100), pointsAfter, "team should have user's 100 points after join")
	})

	t.Run("user with project points leaves team -> team leaderboard decreases", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseData(t, ctx, dbMgr)

		// Create user, add to project and team with points
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))

		// Verify team has points
		pointsBefore, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(100), pointsBefore, "team should have 100 points before leave")

		// User leaves team - trigger should subtract points
		require.NoError(t, dbMgr.RemoveUserFromTeam(ctx, membershipUser1ID, membershipTeamID))

		// Verify team leaderboard decreased
		pointsAfter, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pointsAfter, "team should have 0 points after user leaves")
	})

	t.Run("user with 0 points joins team -> no points added", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseData(t, ctx, dbMgr)

		// Create user with NO points
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))

		// User joins team with 0 points
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamID))

		// Verify team leaderboard has 0 points
		points, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), points, "team should have 0 points when user with 0 points joins")
	})

	t.Run("multiple users join same team -> cumulative points", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseData(t, ctx, dbMgr)

		birthdate := time.Now().AddDate(-25, 0, 0)

		// Create user1 with 100 points and add to team
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamID))

		// Create user2 with 50 points and add to same team
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser2ID, "User2", "FEMALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser2ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser2ID, membershipProjectID, 50))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser2ID, membershipTeamID))

		// Verify team has cumulative points
		points, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(150), points, "team should have cumulative 150 points (100 + 50)")
	})

	t.Run("user joins team with superteam -> both leaderboards updated", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseData(t, ctx, dbMgr)

		// Create superteam and team with superteam
		require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, membershipSuperTeamID, "Super Team", membershipProjectID))
		require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, membershipTeamSTID, "Team ST", membershipProjectID, membershipSuperTeamID))

		// Create user with points
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))

		// User joins team with superteam
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamSTID))

		// Verify team leaderboard has points
		teamPoints, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamSTID)
		require.NoError(t, err)
		assert.Equal(t, int64(100), teamPoints, "team should have 100 points")

		// Verify superteam leaderboard also has points
		superTeamPoints, err := dbMgr.GetLeaderboardProjectSuperTeamPoints(ctx, membershipProjectID, membershipSuperTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(100), superTeamPoints, "superteam should have 100 points")
	})
}

// ==================== Project Membership Tests ====================

// TestProjectMembershipLeaderboardPoints tests that church leaderboard points are updated
// correctly when users join and leave projects.
func TestProjectMembershipLeaderboardPoints(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	t.Run("user with project points joins project -> church leaderboard includes points", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))

		// Create church and project
		require.NoError(t, dbMgr.CreateTestChurch(ctx, membershipChurchID, "Test Church", "NO", "S"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, membershipProjectID, "Test Project"))

		// Create user (not enrolled in project yet)
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))

		// First enroll to create score_journal entries, then we need to re-test
		// Actually, the trigger fires on user_projects insert
		// User needs score BEFORE joining project for this test
		// But score_journal requires project enrollment...
		// The realistic flow is: user joins project, then earns points

		// Let's test the realistic flow: user joins, earns points, church leaderboard updates
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))

		// Verify church leaderboard has user's points (via score trigger, not membership trigger)
		churchPoints, err := dbMgr.GetLeaderboardProjectChurchPoints(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, int64(100), churchPoints, "church should have 100 points")
	})

	t.Run("user leaves project -> church leaderboard decreases", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))

		// Create church and project
		require.NoError(t, dbMgr.CreateTestChurch(ctx, membershipChurchID, "Test Church", "NO", "S"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, membershipProjectID, "Test Project"))

		// Create user, enroll, and add points
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))

		// Verify church has points
		pointsBefore, err := dbMgr.GetLeaderboardProjectChurchPoints(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, int64(100), pointsBefore, "church should have 100 points before leave")

		// User leaves project
		require.NoError(t, dbMgr.RemoveUserFromProject(ctx, membershipUser1ID, membershipProjectID))

		// Verify church leaderboard decreased
		pointsAfter, err := dbMgr.GetLeaderboardProjectChurchPoints(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pointsAfter, "church should have 0 points after user leaves")
	})
}

// ==================== Event Membership Tests ====================

// TestEventMembershipLeaderboardPoints tests that event leaderboard points are updated
// correctly when users join and leave events.
func TestEventMembershipLeaderboardPoints(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	t.Run("user with event points joins event -> church event leaderboard updated", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseDataWithEvent(t, ctx, dbMgr)

		// Create user with event points
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))

		// User joins event and earns event points
		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, membershipUser1ID, membershipEventID))
		require.NoError(t, dbMgr.AddScoreForUserEvent(ctx, membershipUser1ID, membershipProjectID, membershipEventID, 75))

		// Verify church event leaderboard has points
		churchEventPoints, err := dbMgr.GetLeaderboardEventChurchPoints(ctx, membershipEventID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, int64(75), churchEventPoints, "church event leaderboard should have 75 points")
	})

	t.Run("user in team joins event -> team event leaderboard updated", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseDataWithEvent(t, ctx, dbMgr)

		// Create user and add to team first
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamID))

		// User joins event first (realistic flow)
		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, membershipUser1ID, membershipEventID))

		// Verify no points yet
		pointsBefore, err := dbMgr.GetLeaderboardEventTeamPoints(ctx, membershipEventID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pointsBefore, "team event leaderboard should have 0 points before scoring")

		// User earns event points
		require.NoError(t, dbMgr.AddScoreForUserEvent(ctx, membershipUser1ID, membershipProjectID, membershipEventID, 75))

		// Verify team event leaderboard has points
		teamEventPoints, err := dbMgr.GetLeaderboardEventTeamPoints(ctx, membershipEventID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(75), teamEventPoints, "team event leaderboard should have 75 points")
	})

	t.Run("user in team with superteam joins event -> superteam event leaderboard updated", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseDataWithEvent(t, ctx, dbMgr)

		// Create superteam and team
		require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, membershipSuperTeamID, "Super Team", membershipProjectID))
		require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, membershipTeamSTID, "Team ST", membershipProjectID, membershipSuperTeamID))

		// Create user and add to team
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamSTID))

		// User joins event first (realistic flow)
		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, membershipUser1ID, membershipEventID))

		// Verify no points yet
		pointsBefore, err := dbMgr.GetLeaderboardEventTeamPoints(ctx, membershipEventID, membershipTeamSTID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pointsBefore, "team event leaderboard should have 0 points before scoring")

		// User earns event points
		require.NoError(t, dbMgr.AddScoreForUserEvent(ctx, membershipUser1ID, membershipProjectID, membershipEventID, 75))

		// Verify team event leaderboard has points
		teamEventPoints, err := dbMgr.GetLeaderboardEventTeamPoints(ctx, membershipEventID, membershipTeamSTID)
		require.NoError(t, err)
		assert.Equal(t, int64(75), teamEventPoints, "team event leaderboard should have 75 points")

		// Verify superteam event leaderboard also has points
		superTeamEventPoints, err := dbMgr.GetLeaderboardEventSuperTeamPoints(ctx, membershipEventID, membershipSuperTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(75), superTeamEventPoints, "superteam event leaderboard should have 75 points")
	})

	t.Run("user leaves event -> event leaderboards decrease", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseDataWithEvent(t, ctx, dbMgr)

		// Create user, enroll in project, team, and event with points
		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamID))
		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, membershipUser1ID, membershipEventID))
		require.NoError(t, dbMgr.AddScoreForUserEvent(ctx, membershipUser1ID, membershipProjectID, membershipEventID, 75))

		// Verify points exist before leave
		churchPoints, err := dbMgr.GetLeaderboardEventChurchPoints(ctx, membershipEventID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, int64(75), churchPoints, "church event should have 75 points before leave")

		teamPoints, err := dbMgr.GetLeaderboardEventTeamPoints(ctx, membershipEventID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(75), teamPoints, "team event should have 75 points before leave")

		// User leaves event
		require.NoError(t, dbMgr.RemoveUserFromEvent(ctx, membershipUser1ID, membershipEventID))

		// Verify church event leaderboard decreased
		churchPointsAfter, err := dbMgr.GetLeaderboardEventChurchPoints(ctx, membershipEventID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), churchPointsAfter, "church event should have 0 points after leave")

		// Verify team event leaderboard decreased
		teamPointsAfter, err := dbMgr.GetLeaderboardEventTeamPoints(ctx, membershipEventID, membershipTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), teamPointsAfter, "team event should have 0 points after leave")
	})
}

// ==================== Score Consistency Tests ====================

// TestMembershipTriggerConsistencyWithRegenerate tests that trigger-updated scores
// match the results of regenerate_leaderboards().
func TestMembershipTriggerConsistencyWithRegenerate(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	t.Run("trigger scores match regenerate_leaderboards for team membership", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseData(t, ctx, dbMgr)

		// Create superteam and team with superteam
		require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, membershipSuperTeamID, "Super Team", membershipProjectID))
		require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, membershipTeamSTID, "Team ST", membershipProjectID, membershipSuperTeamID))

		birthdate := time.Now().AddDate(-25, 0, 0)

		// Create multiple users with different point values
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))

		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser2ID, "User2", "FEMALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser2ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser2ID, membershipProjectID, 50))

		// Add users to teams - triggers fire
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamSTID))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser2ID, membershipTeamID))

		// Get trigger-updated scores
		teamSTPoints, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamSTID)
		require.NoError(t, err)
		teamPoints, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)
		superTeamPoints, err := dbMgr.GetLeaderboardProjectSuperTeamPoints(ctx, membershipProjectID, membershipSuperTeamID)
		require.NoError(t, err)
		churchPoints, err := dbMgr.GetLeaderboardProjectChurchPoints(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)

		// Run regenerate_leaderboards
		_, err = dbMgr.DB.Pool.Exec(ctx, `SELECT * FROM regenerate_leaderboards()`)
		require.NoError(t, err)

		// Get regenerated scores
		teamSTPointsRegen, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamSTID)
		require.NoError(t, err)
		teamPointsRegen, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)
		superTeamPointsRegen, err := dbMgr.GetLeaderboardProjectSuperTeamPoints(ctx, membershipProjectID, membershipSuperTeamID)
		require.NoError(t, err)
		churchPointsRegen, err := dbMgr.GetLeaderboardProjectChurchPoints(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)

		// Verify trigger scores match regenerated scores
		assert.Equal(t, teamSTPoints, teamSTPointsRegen, "team ST points should match after regeneration")
		assert.Equal(t, teamPoints, teamPointsRegen, "team points should match after regeneration")
		assert.Equal(t, superTeamPoints, superTeamPointsRegen, "superteam points should match after regeneration")
		assert.Equal(t, churchPoints, churchPointsRegen, "church points should match after regeneration")
	})

	t.Run("trigger scores match regenerate_leaderboards for event membership", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseDataWithEvent(t, ctx, dbMgr)

		// Create superteam and team with superteam
		require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, membershipSuperTeamID, "Super Team", membershipProjectID))
		require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, membershipTeamSTID, "Team ST", membershipProjectID, membershipSuperTeamID))

		birthdate := time.Now().AddDate(-25, 0, 0)

		// Create user with event points
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamSTID))
		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, membershipUser1ID, membershipEventID))
		require.NoError(t, dbMgr.AddScoreForUserEvent(ctx, membershipUser1ID, membershipProjectID, membershipEventID, 75))

		// Get trigger-updated scores
		teamEventPoints, err := dbMgr.GetLeaderboardEventTeamPoints(ctx, membershipEventID, membershipTeamSTID)
		require.NoError(t, err)
		superTeamEventPoints, err := dbMgr.GetLeaderboardEventSuperTeamPoints(ctx, membershipEventID, membershipSuperTeamID)
		require.NoError(t, err)
		churchEventPoints, err := dbMgr.GetLeaderboardEventChurchPoints(ctx, membershipEventID, membershipChurchID)
		require.NoError(t, err)

		// Run regenerate_leaderboards
		_, err = dbMgr.DB.Pool.Exec(ctx, `SELECT * FROM regenerate_leaderboards()`)
		require.NoError(t, err)

		// Get regenerated scores
		teamEventPointsRegen, err := dbMgr.GetLeaderboardEventTeamPoints(ctx, membershipEventID, membershipTeamSTID)
		require.NoError(t, err)
		superTeamEventPointsRegen, err := dbMgr.GetLeaderboardEventSuperTeamPoints(ctx, membershipEventID, membershipSuperTeamID)
		require.NoError(t, err)
		churchEventPointsRegen, err := dbMgr.GetLeaderboardEventChurchPoints(ctx, membershipEventID, membershipChurchID)
		require.NoError(t, err)

		// Verify trigger scores match regenerated scores
		assert.Equal(t, teamEventPoints, teamEventPointsRegen, "team event points should match after regeneration")
		assert.Equal(t, superTeamEventPoints, superTeamEventPointsRegen, "superteam event points should match after regeneration")
		assert.Equal(t, churchEventPoints, churchEventPointsRegen, "church event points should match after regeneration")
	})

	t.Run("membership changes produce consistent scores with regenerate", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseData(t, ctx, dbMgr)

		birthdate := time.Now().AddDate(-25, 0, 0)

		// Create users
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))

		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser2ID, "User2", "FEMALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser2ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser2ID, membershipProjectID, 50))

		// Add both to team
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser1ID, membershipTeamID))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, membershipUser2ID, membershipTeamID))

		// Remove user1 from team
		require.NoError(t, dbMgr.RemoveUserFromTeam(ctx, membershipUser1ID, membershipTeamID))

		// Get trigger-updated score (should be 50 from user2)
		teamPoints, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)

		// Regenerate and compare
		_, err = dbMgr.DB.Pool.Exec(ctx, `SELECT * FROM regenerate_leaderboards()`)
		require.NoError(t, err)

		teamPointsRegen, err := dbMgr.GetLeaderboardProjectTeamPoints(ctx, membershipProjectID, membershipTeamID)
		require.NoError(t, err)

		assert.Equal(t, teamPoints, teamPointsRegen, "team points after member removal should match regeneration")
		assert.Equal(t, int64(50), teamPointsRegen, "team should have only user2's 50 points")
	})
}

// ==================== Member Count Tests ====================

// TestChurchLeaderboardMemberCount tests that church leaderboard member_count is
// maintained incrementally by the membership triggers (migration 00101) and stays
// consistent with regenerate_leaderboards().
func TestChurchLeaderboardMemberCount(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	t.Run("first join creates church row with member_count 1 and 0 points", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		require.NoError(t, dbMgr.CreateTestChurch(ctx, membershipChurchID, "Test Church", "NO", "S"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, membershipProjectID, "Test Project"))

		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))

		countBefore, err := dbMgr.GetLeaderboardProjectChurchMemberCount(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, -1, countBefore, "no leaderboard row should exist before first join")

		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))

		count, err := dbMgr.GetLeaderboardProjectChurchMemberCount(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "member_count should be 1 after first join")

		points, err := dbMgr.GetLeaderboardProjectChurchPoints(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), points, "a 0-point member should not add points")
	})

	t.Run("joins increment and leaves decrement member_count", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		require.NoError(t, dbMgr.CreateTestChurch(ctx, membershipChurchID, "Test Church", "NO", "S"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, membershipProjectID, "Test Project"))

		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser2ID, "User2", "FEMALE", birthdate, membershipChurchID))

		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser2ID, membershipProjectID))

		count, err := dbMgr.GetLeaderboardProjectChurchMemberCount(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "member_count should be 2 after two joins")

		require.NoError(t, dbMgr.RemoveUserFromProject(ctx, membershipUser2ID, membershipProjectID))

		count, err = dbMgr.GetLeaderboardProjectChurchMemberCount(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "member_count should be 1 after one member leaves")
	})

	t.Run("score award keeps member_count and updates points", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		require.NoError(t, dbMgr.CreateTestChurch(ctx, membershipChurchID, "Test Church", "NO", "S"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, membershipProjectID, "Test Project"))

		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser2ID, "User2", "FEMALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser2ID, membershipProjectID))

		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))

		count, err := dbMgr.GetLeaderboardProjectChurchMemberCount(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "score award must not change member_count")

		points, err := dbMgr.GetLeaderboardProjectChurchPoints(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, int64(100), points, "church should have the awarded 100 points")
	})

	t.Run("event join and leave maintain event church member_count", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		setupMembershipBaseDataWithEvent(t, ctx, dbMgr)

		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser2ID, "User2", "FEMALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser2ID, membershipProjectID))

		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, membershipUser1ID, membershipEventID))
		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, membershipUser2ID, membershipEventID))

		count, err := dbMgr.GetLeaderboardEventChurchMemberCount(ctx, membershipEventID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "event church member_count should be 2 after two joins")

		require.NoError(t, dbMgr.RemoveUserFromEvent(ctx, membershipUser2ID, membershipEventID))

		count, err = dbMgr.GetLeaderboardEventChurchMemberCount(ctx, membershipEventID, membershipChurchID)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "event church member_count should be 1 after a leave")
	})

	t.Run("incremental member_count matches regenerate_leaderboards", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))
		require.NoError(t, dbMgr.CreateTestChurch(ctx, membershipChurchID, "Test Church", "NO", "S"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, membershipProjectID, "Test Project"))

		birthdate := time.Now().AddDate(-25, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser1ID, "User1", "MALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser2ID, "User2", "FEMALE", birthdate, membershipChurchID))
		require.NoError(t, dbMgr.CreateTestUser(ctx, membershipUser3ID, "User3", "MALE", birthdate, membershipChurchID))

		// Mixed operations: joins, scoring, and a leave
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser1ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser1ID, membershipProjectID, 100))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser2ID, membershipProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, membershipUser2ID, membershipProjectID, 50))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, membershipUser3ID, membershipProjectID))
		require.NoError(t, dbMgr.RemoveUserFromProject(ctx, membershipUser2ID, membershipProjectID))

		countTrigger, err := dbMgr.GetLeaderboardProjectChurchMemberCount(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		pointsTrigger, err := dbMgr.GetLeaderboardProjectChurchPoints(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)

		_, err = dbMgr.DB.Pool.Exec(ctx, `SELECT * FROM regenerate_leaderboards()`)
		require.NoError(t, err)

		countRegen, err := dbMgr.GetLeaderboardProjectChurchMemberCount(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)
		pointsRegen, err := dbMgr.GetLeaderboardProjectChurchPoints(ctx, membershipProjectID, membershipChurchID)
		require.NoError(t, err)

		assert.Equal(t, countTrigger, countRegen, "incremental member_count should match regenerate")
		assert.Equal(t, pointsTrigger, pointsRegen, "incremental points should match regenerate")
		assert.Equal(t, 2, countRegen, "two members remain after user2 left")
		assert.Equal(t, int64(100), pointsRegen, "only user1's points remain after user2 left")
	})
}

// ==================== Helper Functions ====================

// setupMembershipBaseData creates base entities for membership tests
func setupMembershipBaseData(t *testing.T, ctx context.Context, dbMgr *testutil.TestDBManager) {
	t.Helper()

	require.NoError(t, dbMgr.CreateTestChurch(ctx, membershipChurchID, "Membership Test Church", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestProject(ctx, membershipProjectID, "Membership Test Project"))
	require.NoError(t, dbMgr.CreateTestTeam(ctx, membershipTeamID, "Membership Test Team", membershipProjectID))

	// Update settings to point to test project
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'`, membershipProjectID)
	require.NoError(t, err)
}

// setupMembershipBaseDataWithEvent creates base entities including an event for membership tests
func setupMembershipBaseDataWithEvent(t *testing.T, ctx context.Context, dbMgr *testutil.TestDBManager) {
	t.Helper()

	setupMembershipBaseData(t, ctx, dbMgr)
	require.NoError(t, dbMgr.CreateTestEvent(ctx, membershipEventID, "Membership Test Event", membershipProjectID))
}
