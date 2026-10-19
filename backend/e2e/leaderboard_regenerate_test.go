package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findLeaderboardRow returns the snapshot row of table whose column equals
// id, decoded, or nil if there is none
func findLeaderboardRow(t *testing.T, snap map[string][]string, table, column, id string) map[string]any {
	t.Helper()
	for _, raw := range snap[table] {
		var row map[string]any
		require.NoError(t, json.Unmarshal([]byte(raw), &row))
		if row[column] == id {
			return row
		}
	}
	return nil
}

// TestRegenerateLeaderboardsMatchesIncrementalApply checks that the full
// rebuild (regenerate_leaderboards) ends in exactly the state the incremental
// apply leaves: on seed data, for a user in two teams of one superteam (the
// superteam counts them once), and for net scores of zero or below (a lost
// bet stake keeps the row).
func TestRegenerateLeaderboardsMatchesIncrementalApply(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.NoError(t, dbMgr.DrainLeaderboardApplyQueue(ctx))

	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]
	churchID := data.ChurchIDs[0]

	superTeamID := ulid.NewSuperTeamID()
	teamAID := ulid.NewTeamID()
	teamBID := ulid.NewTeamID()
	require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, superTeamID, "Regenerate Super Team", projectID))
	require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, teamAID, "Regenerate Team A", projectID, superTeamID))
	require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, teamBID, "Regenerate Team B", projectID, superTeamID))

	multiTeamUserID := ulid.NewUserID()
	teammateID := ulid.NewUserID()
	loserID := ulid.NewUserID()
	birthdate := time.Now().AddDate(-25, 0, 0)
	for _, m := range []struct {
		userID string
		teams  []string
	}{
		{multiTeamUserID, []string{teamAID, teamBID}},
		{teammateID, []string{teamAID}},
		{loserID, []string{teamBID}},
	} {
		require.NoError(t, dbMgr.CreateTestUser(ctx, m.userID, "Regenerate User", "MALE", birthdate, churchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, m.userID, projectID))
		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, m.userID, eventID))
		for _, teamID := range m.teams {
			require.NoError(t, dbMgr.AddUserToTeam(ctx, m.userID, teamID))
		}
	}

	// Scored after joining the teams and the event
	reason := "regenerate test"
	at := time.Now().Add(-time.Hour).Truncate(time.Millisecond)
	for i, e := range []struct {
		userID string
		event  *string
		points int32
	}{
		{multiTeamUserID, nil, 100},
		{multiTeamUserID, &eventID, 40},
		{teammateID, &eventID, 10},
		{loserID, &eventID, 20},
		{loserID, &eventID, -50}, // lost stake: net -30 in the event and the project
	} {
		_, err := dbMgr.DB.Queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
			ID:         ulid.NewScoreJournalID(),
			ProjectID:  projectID,
			UserID:     e.userID,
			EventID:    e.event,
			Points:     e.points,
			SourceType: "MANUAL",
			Reason:     &reason,
			CreatedAt:  pgtype.Timestamptz{Time: at.Add(time.Duration(i) * time.Second), Valid: true},
		})
		require.NoError(t, err)
	}
	require.NoError(t, dbMgr.DrainLeaderboardApplyQueue(ctx))
	incremental := snapshotLeaderboards(t, ctx, dbMgr.DB.Queries)

	require.NoError(t, dbMgr.DB.Queries.RegenerateLeaderboards(ctx))
	regenerated := snapshotLeaderboards(t, ctx, dbMgr.DB.Queries)

	for _, table := range leaderboardTables {
		assert.NotEmpty(t, incremental[table], "%s has rows", table)
		assert.Equal(t, incremental[table], regenerated[table], "%s: regenerate equals the incremental apply", table)
	}

	t.Run("superteam counts a user in two of its teams once", func(t *testing.T) {
		// 100 + 40 (multi-team user) + 10 (teammate) - 30 (loser)
		projectPoints, err := dbMgr.GetLeaderboardProjectSuperTeamPoints(ctx, projectID, superTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(120), projectPoints)

		// 40 (multi-team user) + 10 (teammate) - 30 (loser)
		eventPoints, err := dbMgr.GetLeaderboardEventSuperTeamPoints(ctx, eventID, superTeamID)
		require.NoError(t, err)
		assert.Equal(t, int64(20), eventPoints)
	})

	t.Run("net score below zero keeps the row", func(t *testing.T) {
		for _, table := range []string{"leaderboard_project_persons", "leaderboard_event_persons"} {
			row := findLeaderboardRow(t, regenerated, table, "user_id", loserID)
			require.NotNil(t, row, "%s keeps the loser's row", table)
			assert.EqualValues(t, -30, row["score"], "%s score", table)
		}
	})
}
