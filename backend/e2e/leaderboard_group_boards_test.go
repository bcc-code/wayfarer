package e2e

import (
	"context"
	"encoding/json"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// groupBoardRow is the row every group leaderboard query returns
type groupBoardRow struct {
	EntityID    string             `json:"entity_id"`
	Name        string             `json:"name"`
	Image       *string            `json:"image"`
	Score       int64              `json:"score"`
	Rank        int64              `json:"rank"`
	LastScoreAt pgtype.Timestamptz `json:"last_score_at"`
}

// toGroupBoardRows converts the rows of any group leaderboard query, sorted by
// rank and ID: rows tied on rank, last score and name have no fixed order
func toGroupBoardRows[T any](t *testing.T, rows []*T, err error) []groupBoardRow {
	t.Helper()
	require.NoError(t, err)
	raw, err := json.Marshal(rows)
	require.NoError(t, err)
	var out []groupBoardRow
	require.NoError(t, json.Unmarshal(raw, &out))
	slices.SortFunc(out, func(a, b groupBoardRow) int {
		if a.Rank != b.Rank {
			return int(a.Rank - b.Rank)
		}
		return strings.Compare(a.EntityID, b.EntityID)
	})
	return out
}

func findGroupBoardRow(rows []groupBoardRow, id string) *groupBoardRow {
	for i := range rows {
		if rows[i].EntityID == id {
			return &rows[i]
		}
	}
	return nil
}

// TestGroupLeaderboardsMatchStoredTables checks that the team, superteam and
// church boards, now averaged from the person totals on read, return exactly
// what the stored group tables give after regenerate_leaderboards(): same
// rows, scores, ranks and last score times, under every filter. The fixture
// covers each member rule of the rebuild: a user in two teams of a superteam,
// a net loss, event points of a team member not enrolled in the event, project
// points of a church user not enrolled in the project, an excluded team and a
// team and superteam nobody in has scored.
func TestGroupLeaderboardsMatchStoredTables(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.NoError(t, dbMgr.DrainLeaderboardApplyQueue(ctx))
	q := dbMgr.DB.Queries

	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]

	churchID := ulid.NewChurchID()
	require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Group Board Church", "NO", "S"))

	superTeamID := ulid.NewSuperTeamID()
	teamAID := ulid.NewTeamID()
	teamBID := ulid.NewTeamID()
	excludedTeamID := ulid.NewTeamID()
	idleTeamID := ulid.NewTeamID()
	idleSuperTeamID := ulid.NewSuperTeamID()
	require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, superTeamID, "Group Board Super Team", projectID))
	require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, teamAID, "Group Board Team A", projectID, superTeamID))
	require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, teamBID, "Group Board Team B", projectID, superTeamID))
	require.NoError(t, dbMgr.CreateTestTeam(ctx, excludedTeamID, "Group Board Excluded Team", projectID))
	require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, idleSuperTeamID, "Group Board Idle Super Team", projectID))
	require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, idleTeamID, "Group Board Idle Team", projectID, idleSuperTeamID))
	_, err = dbMgr.DB.Pool.Exec(ctx, `UPDATE teams SET leaderboard_excluded = true WHERE id = $1`, excludedTeamID)
	require.NoError(t, err)

	multiTeamUserID := ulid.NewUserID()
	teammateID := ulid.NewUserID()
	loserID := ulid.NewUserID()
	notInEventID := ulid.NewUserID()
	excludedMemberID := ulid.NewUserID()
	idleMemberID := ulid.NewUserID()
	notInProjectID := ulid.NewUserID()
	birthdate := time.Now().AddDate(-25, 0, 0)
	for _, m := range []struct {
		userID    string
		teams     []string
		inProject bool
		inEvent   bool
	}{
		{multiTeamUserID, []string{teamAID, teamBID}, true, true},
		{teammateID, []string{teamAID}, true, true},
		{loserID, []string{teamBID}, true, true},
		{notInEventID, []string{teamAID}, true, false},
		{excludedMemberID, []string{excludedTeamID}, true, true},
		{idleMemberID, []string{idleTeamID}, true, true},
		{notInProjectID, nil, false, false},
	} {
		require.NoError(t, dbMgr.CreateTestUser(ctx, m.userID, "Group Board User", "MALE", birthdate, churchID))
		if m.inProject {
			require.NoError(t, dbMgr.EnrollUserInProject(ctx, m.userID, projectID))
		}
		if m.inEvent {
			require.NoError(t, dbMgr.EnrollUserInEvent(ctx, m.userID, eventID))
		}
		for _, teamID := range m.teams {
			require.NoError(t, dbMgr.AddUserToTeam(ctx, m.userID, teamID))
		}
	}

	reason := "group board test"
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
		{loserID, &eventID, -50}, // lost stake: net -30
		{notInEventID, &eventID, 30},
		{excludedMemberID, nil, 500},
		{notInProjectID, nil, 60},
	} {
		_, err := q.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
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
	require.NoError(t, q.RegenerateLeaderboards(ctx))

	// The stored tables count users blocked by a required consent; the boards
	// do not. The comparison holds only while nobody is blocked.
	blocked, err := q.GetConsentBlockedUsersInProjects(ctx, []string{projectID})
	require.NoError(t, err)
	require.Empty(t, blocked, "the fixture has no blocked users")

	// The default filter (no min/max score) is what the service passes when
	// the client sets none
	filters := []struct {
		name               string
		minScore, maxScore int32
		churchID           string
	}{
		{"no filter", math.MinInt32, math.MaxInt32, ""},
		{"minScore 1", 1, math.MaxInt32, ""},
		{"minScore 0", 0, math.MaxInt32, ""},
		{"maxScore 50", math.MinInt32, 50, ""},
		{"churchId", math.MinInt32, math.MaxInt32, churchID},
	}

	boards := map[string]func(minScore, maxScore int32, churchID string) (computed, stored []groupBoardRow){
		"project teams": func(minScore, maxScore int32, churchID string) ([]groupBoardRow, []groupBoardRow) {
			computed, err := q.GetFullProjectTeamLeaderboard(ctx, sqlc.GetFullProjectTeamLeaderboardParams{Projectid: projectID, Minscore: minScore, Maxscore: maxScore, Churchid: churchID})
			stored, storedErr := q.StoredProjectTeamLeaderboard(ctx, sqlc.StoredProjectTeamLeaderboardParams{Projectid: projectID, Minscore: minScore, Maxscore: maxScore, Churchid: churchID})
			return toGroupBoardRows(t, computed, err), toGroupBoardRows(t, stored, storedErr)
		},
		"project superteams": func(minScore, maxScore int32, _ string) ([]groupBoardRow, []groupBoardRow) {
			computed, err := q.GetFullProjectSuperTeamLeaderboard(ctx, sqlc.GetFullProjectSuperTeamLeaderboardParams{Projectid: projectID, Minscore: minScore, Maxscore: maxScore})
			stored, storedErr := q.StoredProjectSuperTeamLeaderboard(ctx, sqlc.StoredProjectSuperTeamLeaderboardParams{Projectid: projectID, Minscore: minScore, Maxscore: maxScore})
			return toGroupBoardRows(t, computed, err), toGroupBoardRows(t, stored, storedErr)
		},
		"project churches": func(minScore, maxScore int32, _ string) ([]groupBoardRow, []groupBoardRow) {
			computed, err := q.GetFullProjectChurchLeaderboard(ctx, sqlc.GetFullProjectChurchLeaderboardParams{Projectid: projectID, Minscore: minScore, Maxscore: maxScore})
			stored, storedErr := q.StoredProjectChurchLeaderboard(ctx, sqlc.StoredProjectChurchLeaderboardParams{Projectid: projectID, Minscore: minScore, Maxscore: maxScore})
			return toGroupBoardRows(t, computed, err), toGroupBoardRows(t, stored, storedErr)
		},
		"event teams": func(minScore, maxScore int32, churchID string) ([]groupBoardRow, []groupBoardRow) {
			computed, err := q.GetFullEventTeamLeaderboard(ctx, sqlc.GetFullEventTeamLeaderboardParams{Eventid: eventID, Minscore: minScore, Maxscore: maxScore, Churchid: churchID})
			stored, storedErr := q.StoredEventTeamLeaderboard(ctx, sqlc.StoredEventTeamLeaderboardParams{Eventid: eventID, Minscore: minScore, Maxscore: maxScore, Churchid: churchID})
			return toGroupBoardRows(t, computed, err), toGroupBoardRows(t, stored, storedErr)
		},
		"event superteams": func(minScore, maxScore int32, _ string) ([]groupBoardRow, []groupBoardRow) {
			computed, err := q.GetFullEventSuperTeamLeaderboard(ctx, sqlc.GetFullEventSuperTeamLeaderboardParams{Eventid: eventID, Minscore: minScore, Maxscore: maxScore})
			stored, storedErr := q.StoredEventSuperTeamLeaderboard(ctx, sqlc.StoredEventSuperTeamLeaderboardParams{Eventid: eventID, Minscore: minScore, Maxscore: maxScore})
			return toGroupBoardRows(t, computed, err), toGroupBoardRows(t, stored, storedErr)
		},
		"event churches": func(minScore, maxScore int32, _ string) ([]groupBoardRow, []groupBoardRow) {
			computed, err := q.GetFullEventChurchLeaderboard(ctx, sqlc.GetFullEventChurchLeaderboardParams{Eventid: eventID, Minscore: minScore, Maxscore: maxScore})
			stored, storedErr := q.StoredEventChurchLeaderboard(ctx, sqlc.StoredEventChurchLeaderboardParams{Eventid: eventID, Minscore: minScore, Maxscore: maxScore})
			return toGroupBoardRows(t, computed, err), toGroupBoardRows(t, stored, storedErr)
		},
	}

	for boardName, board := range boards {
		for _, f := range filters {
			t.Run(boardName+"/"+f.name, func(t *testing.T) {
				computed, stored := board(f.minScore, f.maxScore, f.churchID)
				if f.name == "no filter" {
					require.NotEmpty(t, stored, "the fixture puts rows on this board")
				}
				assert.Equal(t, stored, computed)
			})
		}
	}

	// Pin the fixture's rules down, so the comparison can't pass on boards
	// that both miss them
	t.Run("member rules", func(t *testing.T) {
		projectTeams, _ := boards["project teams"](math.MinInt32, math.MaxInt32, "")
		eventTeams, _ := boards["event teams"](math.MinInt32, math.MaxInt32, "")
		projectSuperTeams, _ := boards["project superteams"](math.MinInt32, math.MaxInt32, "")
		projectChurches, _ := boards["project churches"](math.MinInt32, math.MaxInt32, "")

		// 140 + 10 + 30 over 3 members
		if row := findGroupBoardRow(projectTeams, teamAID); assert.NotNil(t, row) {
			assert.EqualValues(t, 60, row.Score)
		}
		// all members' event points (40 + 10 + 30) over the 2 in the event
		if row := findGroupBoardRow(eventTeams, teamAID); assert.NotNil(t, row) {
			assert.EqualValues(t, 40, row.Score)
		}
		// 140 + 10 - 30 + 30, the multi-team user once, over 4 members
		if row := findGroupBoardRow(projectSuperTeams, superTeamID); assert.NotNil(t, row) {
			assert.EqualValues(t, 38, row.Score)
		}
		// 140 + 10 - 30 + 30 + 500 + 60 (incl. the user outside the project)
		// over the 6 enrolled users: 118.3
		if row := findGroupBoardRow(projectChurches, churchID); assert.NotNil(t, row) {
			assert.EqualValues(t, 118, row.Score)
		}
		assert.Nil(t, findGroupBoardRow(projectTeams, excludedTeamID), "excluded team is not listed")
		assert.Nil(t, findGroupBoardRow(projectTeams, idleTeamID), "team nobody in has scored is not listed")
		assert.Nil(t, findGroupBoardRow(projectSuperTeams, idleSuperTeamID), "superteam nobody in has scored is not listed")
	})
}

// TestGroupLeaderboardsFollowSuperTeamAndChurchChanges checks that moving a
// team to another superteam and moving a user to another church carry the
// points along at once: nothing maintains stored group totals for either change
func TestGroupLeaderboardsFollowSuperTeamAndChurchChanges(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)
	require.NoError(t, dbMgr.DrainLeaderboardApplyQueue(ctx))
	q := dbMgr.DB.Queries
	projectID := data.ProjectIDs[0]

	oldChurchID := ulid.NewChurchID()
	newChurchID := ulid.NewChurchID()
	require.NoError(t, dbMgr.CreateTestChurch(ctx, oldChurchID, "Group Board Old Church", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestChurch(ctx, newChurchID, "Group Board New Church", "NO", "S"))
	oldSuperTeamID := ulid.NewSuperTeamID()
	newSuperTeamID := ulid.NewSuperTeamID()
	teamID := ulid.NewTeamID()
	require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, oldSuperTeamID, "Group Board Old Super Team", projectID))
	require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, newSuperTeamID, "Group Board New Super Team", projectID))
	require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, teamID, "Group Board Moving Team", projectID, oldSuperTeamID))

	userID := ulid.NewUserID()
	require.NoError(t, dbMgr.CreateTestUser(ctx, userID, "Group Board Mover", "MALE", time.Now().AddDate(-25, 0, 0), oldChurchID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userID, projectID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userID, teamID))
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userID, projectID, 100))
	require.NoError(t, dbMgr.DrainLeaderboardApplyQueue(ctx))

	require.NoError(t, q.AssignTeamToSuperTeam(ctx, sqlc.AssignTeamToSuperTeamParams{SuperTeamID: &newSuperTeamID, TeamID: teamID}))
	_, err = dbMgr.DB.Pool.Exec(ctx, `UPDATE users SET church_id = $1 WHERE id = $2`, newChurchID, userID)
	require.NoError(t, err)

	superTeamRows, err := q.GetFullProjectSuperTeamLeaderboard(ctx, sqlc.GetFullProjectSuperTeamLeaderboardParams{Projectid: projectID, Minscore: math.MinInt32, Maxscore: math.MaxInt32})
	superTeams := toGroupBoardRows(t, superTeamRows, err)
	if row := findGroupBoardRow(superTeams, newSuperTeamID); assert.NotNil(t, row, "new superteam is listed") {
		assert.EqualValues(t, 100, row.Score)
	}
	assert.Nil(t, findGroupBoardRow(superTeams, oldSuperTeamID), "old superteam has no members left")

	churchRows, err := q.GetFullProjectChurchLeaderboard(ctx, sqlc.GetFullProjectChurchLeaderboardParams{Projectid: projectID, Minscore: math.MinInt32, Maxscore: math.MaxInt32})
	churches := toGroupBoardRows(t, churchRows, err)
	if row := findGroupBoardRow(churches, newChurchID); assert.NotNil(t, row, "new church is listed") {
		assert.EqualValues(t, 100, row.Score)
	}
	assert.Nil(t, findGroupBoardRow(churches, oldChurchID), "old church has no members left")
}
