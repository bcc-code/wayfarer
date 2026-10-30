package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequiredConsentExcludesUserFromCounts covers requirement 6 of required
// consents (notes/required-consents.md): a user who has not accepted a
// required consent of the project counts nowhere, for admins too. Not in the
// team, superteam and church averages (project and event), the team's average
// age and the auto-exclusion it drives, or the member-count filters. Admins
// still see them in Team.members, marked as not counting. Every consent change
// that blocks or unblocks them takes effect at once, and leaves alone the
// teams of users it did not change.
func TestRequiredConsentExcludesUserFromCounts(t *testing.T) {
	const (
		churchID     = "CH01TEST0000000000000CNTS001"
		projectID    = "PR01TEST0000000000000CNTS001"
		eventID      = "EV01TEST0000000000000CNTS001"
		superTeamID  = "ST01TEST0000000000000CNTS001"
		teamID       = "TM01TEST0000000000000CNTS001" // the blocked user's team
		controlTeam  = "TM01TEST0000000000000CNTS002" // excluded by hand
		firstID      = "US01TEST0000000000000CNTSOK1"
		secondID     = "US01TEST0000000000000CNTSOK2"
		blockedID    = "US01TEST0000000000000CNTSBLK" // has not accepted
		controlID    = "US01TEST0000000000000CNTSCTL"
		adminID      = "US01TEST0000000000000CNTSADM"
		consentKey   = "participation"
		minMembers   = 3
		youngAge     = 20
		blockedAge   = 80
		blockedScore = 1000
	)

	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))

	born := func(age int) time.Time { return time.Now().AddDate(-age, 0, 0) }
	require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Counts Church", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Counts Project"))
	require.NoError(t, setCurrentProject(ctx, dbMgr, projectID))
	require.NoError(t, dbMgr.CreateTestEvent(ctx, eventID, "Counts Event", projectID))
	require.NoError(t, dbMgr.CreateTestSuperTeam(ctx, superTeamID, "Counts Super Team", projectID))
	require.NoError(t, dbMgr.CreateTestTeamWithSuperTeam(ctx, teamID, "Counts Team", projectID, superTeamID))
	require.NoError(t, dbMgr.CreateTestTeam(ctx, controlTeam, "Control Team", projectID))
	// Young enough for the age rule to include it: only a hand-set flag
	// excludes it, which a consent change must not reset
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE teams SET leaderboard_excluded = true WHERE id = $1`, controlTeam)
	require.NoError(t, err)

	for _, u := range []struct {
		id     string
		age    int
		team   string
		points int
	}{
		{firstID, youngAge, teamID, 100},
		{secondID, youngAge, teamID, 50},
		{blockedID, blockedAge, "", blockedScore}, // added to the team through the API below
		{controlID, youngAge, controlTeam, 10},
	} {
		require.NoError(t, dbMgr.CreateTestUser(ctx, u.id, u.id, "MALE", born(u.age), churchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, u.id, projectID))
		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, u.id, eventID))
		if u.team != "" {
			require.NoError(t, dbMgr.AddUserToTeam(ctx, u.id, u.team))
		}
		require.NoError(t, dbMgr.AddScoreForUserEvent(ctx, u.id, projectID, eventID, u.points))
	}

	consentID := ulid.NewConsentID()
	_, err = dbMgr.DB.Pool.Exec(ctx, `
		INSERT INTO consents (id, key, version, title, short_text, body, published_at, is_remote, project_id, is_required)
		VALUES ($1, $2, 1, 'T', 'S', 'B', now() - interval '1 minute', false, $3, true)
	`, consentID, consentKey, projectID)
	require.NoError(t, err)
	for _, id := range []string{firstID, secondID, controlID} {
		_, err = dbMgr.DB.Pool.Exec(ctx, `
			INSERT INTO user_consent_history (id, user_id, consent_id, consent_key, action, occurred_at)
			VALUES ($1, $2, $3, $4, 'ACCEPTED', now())
		`, ulid.NewUserConsentHistoryID(), id, consentID, consentKey)
		require.NoError(t, err)
	}

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()
	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	require.NoError(t, dbMgr.CreateTestUser(ctx, adminID, "Admin", "MALE", born(40), churchID))
	require.NoError(t, dbMgr.AssignRole(ctx, adminID, testutil.RoleAdmin))
	adminToken, err := testutil.GenerateAdminToken(adminID)
	require.NoError(t, err)
	blockedToken, err := testutil.GenerateUserToken(blockedID)
	require.NoError(t, err)

	execute := func(t *testing.T, token, query string, vars map[string]any, out any) {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, query, vars)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
		if out != nil {
			require.NoError(t, resp.UnmarshalData(out))
		}
	}

	// Scores by entity on a board; root is "project" or "event"
	board := func(t *testing.T, root, id, entityType string) map[string]int {
		t.Helper()
		var result map[string]struct {
			Leaderboard struct {
				Edges []struct {
					Node struct {
						ID    string `json:"id"`
						Score int    `json:"score"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"leaderboard"`
		}
		execute(t, adminToken, fmt.Sprintf(`
			query ($id: ID!) {
				%s(id: $id) { leaderboard(entityType: %s, first: 100) { edges { node { id score } } } }
			}
		`, root, entityType), map[string]any{"id": id}, &result)
		scores := make(map[string]int)
		for _, e := range result[root].Leaderboard.Edges {
			scores[e.Node.ID] = e.Node.Score
		}
		return scores
	}

	type teamState struct {
		AverageAge          float64 `json:"averageAge"`
		LeaderboardExcluded bool    `json:"leaderboardExcluded"`
		Members             []struct {
			ID                 string `json:"id"`
			CountsTowardScores bool   `json:"countsTowardScores"`
		} `json:"members"`
	}
	team := func(t *testing.T, id string) teamState {
		t.Helper()
		var result struct {
			Team teamState `json:"team"`
		}
		execute(t, adminToken, `
			query ($id: ID!) {
				team(id: $id) { averageAge leaderboardExcluded members { id countsTowardScores } }
			}
		`, map[string]any{"id": id}, &result)
		return result.Team
	}
	countsToward := func(state teamState) map[string]bool {
		out := make(map[string]bool, len(state.Members))
		for _, m := range state.Members {
			out[m.ID] = m.CountsTowardScores
		}
		return out
	}

	// IDs the teams/superteams lists return with at least minMembers members
	withMinMembers := func(t *testing.T, field string) []string {
		t.Helper()
		var result map[string]struct {
			Edges []struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"edges"`
		}
		execute(t, adminToken, fmt.Sprintf(`
			query ($projectId: ID!, $min: Int) {
				%s(filter: { projectId: $projectId, minMembers: $min }, first: 100) { edges { node { id } } }
			}
		`, field), map[string]any{"projectId": projectID, "min": minMembers}, &result)
		ids := make([]string, 0, len(result[field].Edges))
		for _, e := range result[field].Edges {
			ids = append(ids, e.Node.ID)
		}
		return ids
	}

	const (
		// (100 + 50) / 2 members, without the blocked user
		groupWithout = 75
		// (100 + 50 + 1000) / 3 members
		groupWith = 383
		// (100 + 50 + 10) / 3 church members
		churchWithout = 53
		// (100 + 50 + 10 + 1000) / 4
		churchWith = 290
	)

	// excluded asserts the blocked user counts nowhere
	excluded := func(t *testing.T) {
		t.Helper()
		for _, root := range []struct{ name, id string }{{"project", projectID}, {"event", eventID}} {
			assert.Equal(t, groupWithout, board(t, root.name, root.id, "TEAMS")[teamID], "%s team average", root.name)
			assert.Equal(t, groupWithout, board(t, root.name, root.id, "SUPERTEAMS")[superTeamID], "%s superteam average", root.name)
			assert.Equal(t, churchWithout, board(t, root.name, root.id, "CHURCHES")[churchID], "%s church average", root.name)
		}
		state := team(t, teamID)
		assert.InDelta(t, float64(youngAge), state.AverageAge, 0.01, "average age")
		assert.False(t, state.LeaderboardExcluded, "the age rule ignores the blocked user")
		assert.Equal(t, map[string]bool{firstID: true, secondID: true, blockedID: false}, countsToward(state),
			"admins see the blocked member, not counting")
		assert.NotContains(t, withMinMembers(t, "teams"), teamID, "member count without the blocked user")
		assert.NotContains(t, withMinMembers(t, "superteams"), superTeamID, "member count without the blocked user")
	}

	// counted asserts the user counts again everywhere
	counted := func(t *testing.T) {
		t.Helper()
		for _, root := range []struct{ name, id string }{{"project", projectID}, {"event", eventID}} {
			// The team itself is now excluded by the age rule (average 40)
			assert.NotContains(t, board(t, root.name, root.id, "TEAMS"), teamID, "%s team excluded by age", root.name)
			assert.Equal(t, groupWith, board(t, root.name, root.id, "SUPERTEAMS")[superTeamID], "%s superteam average", root.name)
			assert.Equal(t, churchWith, board(t, root.name, root.id, "CHURCHES")[churchID], "%s church average", root.name)
		}
		state := team(t, teamID)
		assert.InDelta(t, float64(youngAge+youngAge+blockedAge)/3, state.AverageAge, 0.01, "average age")
		assert.True(t, state.LeaderboardExcluded, "the age rule counts the user")
		assert.Equal(t, map[string]bool{firstID: true, secondID: true, blockedID: true}, countsToward(state))
		assert.Contains(t, withMinMembers(t, "teams"), teamID)
		assert.Contains(t, withMinMembers(t, "superteams"), superTeamID)
	}

	controlUntouched := func(t *testing.T) {
		t.Helper()
		assert.True(t, team(t, controlTeam).LeaderboardExcluded, "a team no change affected keeps its hand-set flag")
	}

	// Answers are per version (00117): every other member has to accept each
	// new version too, or they would be left out alongside the blocked user
	othersAccept := func(t *testing.T, consentID string) {
		t.Helper()
		for _, id := range []string{firstID, secondID, controlID} {
			token, err := testutil.GenerateUserToken(id)
			require.NoError(t, err)
			execute(t, token, `mutation ($id: ID!) { acceptConsent(consentId: $id) { id } }`,
				map[string]any{"id": consentID}, nil)
		}
	}

	// Placed in the unit by an admin while blocked: the age rule runs on add
	execute(t, adminToken, `
		mutation ($teamId: ID!, $userIds: [ID!]!) { addTeamMembers(teamId: $teamId, userIds: $userIds) { id } }
	`, map[string]any{"teamId": teamID, "userIds": []string{blockedID}}, nil)

	t.Run("a user who has not accepted counts nowhere", func(t *testing.T) {
		excluded(t)
		controlUntouched(t)
	})

	t.Run("accepting counts them at once", func(t *testing.T) {
		execute(t, blockedToken, `mutation ($id: ID!) { acceptConsent(consentId: $id) { id } }`,
			map[string]any{"id": consentID}, nil)
		counted(t)
		controlUntouched(t)
	})

	t.Run("rejecting leaves them out again", func(t *testing.T) {
		execute(t, blockedToken, `mutation ($id: ID!) { rejectConsent(consentId: $id) { id } }`,
			map[string]any{"id": consentID}, nil)
		excluded(t)
		controlUntouched(t)
	})

	var relaxedID string
	// updateConsent stores publishedAt as given, so it is sent again below:
	// leaving it out would unpublish the version
	relaxedPublishedAt := time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)
	t.Run("a new version that is not required counts them", func(t *testing.T) {
		var result struct {
			CreateConsent struct {
				ID string `json:"id"`
			} `json:"createConsent"`
		}
		execute(t, adminToken, `
			mutation ($key: String!, $publishedAt: DateTime) {
				createConsent(key: $key, title: "T", body: "B", publishedAt: $publishedAt, isRequired: false) { id }
			}
		`, map[string]any{
			"key":         consentKey,
			"publishedAt": relaxedPublishedAt,
		}, &result)
		relaxedID = result.CreateConsent.ID
		require.NotEmpty(t, relaxedID)
		othersAccept(t, relaxedID)
		counted(t)
		controlUntouched(t)
	})

	t.Run("making it required again leaves them out", func(t *testing.T) {
		require.NotEmpty(t, relaxedID)
		execute(t, adminToken, `
			mutation ($id: ID!, $publishedAt: DateTime) {
				updateConsent(id: $id, publishedAt: $publishedAt, isRequired: true) { id }
			}
		`, map[string]any{"id": relaxedID, "publishedAt": relaxedPublishedAt}, nil)
		excluded(t)
		controlUntouched(t)
	})

	t.Run("an admin accepting for them counts them", func(t *testing.T) {
		require.NotEmpty(t, relaxedID)
		execute(t, adminToken, `
			mutation ($userId: ID!, $consentId: ID!) {
				adminSetUserConsent(userId: $userId, consentId: $consentId, action: ACCEPTED) { id }
			}
		`, map[string]any{"userId": blockedID, "consentId": relaxedID}, nil)
		counted(t)
		controlUntouched(t)
	})

	// Answers are per version (00117): a new required version leaves out users
	// who accepted the previous one, until they accept it. Publishing it blocks
	// every member at once, so the age rule re-runs for every unit and resets
	// the control team's hand-set flag (accepted, notes/required-consents.md)
	var v3ID string
	t.Run("a new required version leaves them out until they accept it", func(t *testing.T) {
		var result struct {
			CreateConsent struct {
				ID string `json:"id"`
			} `json:"createConsent"`
		}
		execute(t, adminToken, `
			mutation ($key: String!, $publishedAt: DateTime) {
				createConsent(key: $key, title: "T", body: "B", publishedAt: $publishedAt, isRequired: true) { id }
			}
		`, map[string]any{
			"key":         consentKey,
			"publishedAt": time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		}, &result)
		v3ID = result.CreateConsent.ID
		require.NotEmpty(t, v3ID)
		assert.False(t, team(t, controlTeam).LeaderboardExcluded, "publishing re-ran the age rule")
		othersAccept(t, v3ID)
		excluded(t)
	})

	t.Run("answering an older version is refused", func(t *testing.T) {
		resp := client.WithAuth(blockedToken).MustExecute(t,
			`mutation ($id: ID!) { acceptConsent(consentId: $id) { id } }`,
			map[string]any{"id": relaxedID})
		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "latest version")
		excluded(t)
	})

	t.Run("accepting the new version counts them", func(t *testing.T) {
		require.NotEmpty(t, v3ID)
		execute(t, blockedToken, `mutation ($id: ID!) { acceptConsent(consentId: $id) { id } }`,
			map[string]any{"id": v3ID}, nil)
		counted(t)
	})
}

// TestRequiredConsentCountsEdgeCases covers the edges of "counts nowhere":
// under-13 members stay counted (as in consent_blocked_user_ids), another
// project's required consent changes nothing here, a unit leader sees the
// same numbers as an admin, the team board's church filter ignores a blocked
// member, and the age rule leaves them out when another member is removed.
func TestRequiredConsentCountsEdgeCases(t *testing.T) {
	const (
		churchID       = "CH01TEST0000000000000EDGE001"
		otherChurchID  = "CH01TEST0000000000000EDGE002" // only the blocked user's
		projectID      = "PR01TEST0000000000000EDGE001"
		otherProjectID = "PR01TEST0000000000000EDGE002"
		teamID         = "TM01TEST0000000000000EDGE001"
		leadID         = "US01TEST0000000000000EDGELD1" // accepted here, not the other project's consent
		childID        = "US01TEST0000000000000EDGEKID" // under 13, never answered
		blockedID      = "US01TEST0000000000000EDGEBLK" // never answered
		adminID        = "US01TEST0000000000000EDGEADM"
		leadAge        = 20
		childAge       = 10
		blockedAge     = 80
	)

	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))

	born := func(age int) time.Time { return time.Now().AddDate(-age, 0, 0) }
	require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Edge Church", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestChurch(ctx, otherChurchID, "Other Church", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Edge Project"))
	require.NoError(t, dbMgr.CreateTestProject(ctx, otherProjectID, "Other Project"))
	require.NoError(t, setCurrentProject(ctx, dbMgr, projectID))
	require.NoError(t, dbMgr.CreateTestTeam(ctx, teamID, "Edge Team", projectID))

	for _, u := range []struct {
		id     string
		age    int
		church string
		points int
	}{
		{leadID, leadAge, churchID, 100},
		{childID, childAge, churchID, 50},
		{blockedID, blockedAge, otherChurchID, 1000},
	} {
		require.NoError(t, dbMgr.CreateTestUser(ctx, u.id, u.id, "MALE", born(u.age), u.church))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, u.id, projectID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, u.id, otherProjectID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, u.id, projectID, u.points))
	}
	for _, id := range []string{leadID, childID} {
		require.NoError(t, dbMgr.AddUserToTeam(ctx, id, teamID))
	}

	insertConsent := func(key string, projectID string) string {
		id := ulid.NewConsentID()
		_, err := dbMgr.DB.Pool.Exec(ctx, `
			INSERT INTO consents (id, key, version, title, short_text, body, published_at, is_remote, project_id, is_required)
			VALUES ($1, $2, 1, 'T', 'S', 'B', now() - interval '1 minute', false, $3, true)
		`, id, key, projectID)
		require.NoError(t, err)
		return id
	}
	participation := insertConsent("participation", projectID)
	// Nobody accepts the other project's required consent
	insertConsent("other-participation", otherProjectID)
	_, err := dbMgr.DB.Pool.Exec(ctx, `
		INSERT INTO user_consent_history (id, user_id, consent_id, consent_key, action, occurred_at)
		VALUES ($1, $2, $3, 'participation', 'ACCEPTED', now())
	`, ulid.NewUserConsentHistoryID(), leadID, participation)
	require.NoError(t, err)

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()
	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	require.NoError(t, dbMgr.CreateTestUser(ctx, adminID, "Admin", "MALE", born(40), churchID))
	require.NoError(t, dbMgr.AssignRole(ctx, adminID, testutil.RoleAdmin))
	adminToken, err := testutil.GenerateAdminToken(adminID)
	require.NoError(t, err)
	tid := teamID
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, leadID, testutil.RoleTeamLead, nil, nil, &tid))
	leadToken, err := testutil.GenerateUserToken(leadID)
	require.NoError(t, err)

	execute := func(t *testing.T, token, query string, vars map[string]any, out any) {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, query, vars)
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
		if out != nil {
			require.NoError(t, resp.UnmarshalData(out))
		}
	}

	// Team scores on the project's team board, optionally for one church
	teamBoard := func(t *testing.T, churchFilter string) map[string]int {
		t.Helper()
		var result struct {
			Project struct {
				Leaderboard struct {
					Edges []struct {
						Node struct {
							ID    string `json:"id"`
							Score int    `json:"score"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"leaderboard"`
			} `json:"project"`
		}
		filter := map[string]any{}
		if churchFilter != "" {
			filter["churchId"] = churchFilter
		}
		execute(t, adminToken, `
			query ($id: ID!, $filter: LeaderboardFilter) {
				project(id: $id) { leaderboard(entityType: TEAMS, filter: $filter, first: 100) { edges { node { id score } } } }
			}
		`, map[string]any{"id": projectID, "filter": filter}, &result)
		scores := make(map[string]int)
		for _, e := range result.Project.Leaderboard.Edges {
			scores[e.Node.ID] = e.Node.Score
		}
		return scores
	}

	type teamState struct {
		AverageAge          float64 `json:"averageAge"`
		LeaderboardExcluded bool    `json:"leaderboardExcluded"`
		Members             []struct {
			ID                 string `json:"id"`
			CountsTowardScores bool   `json:"countsTowardScores"`
		} `json:"members"`
	}
	team := func(t *testing.T, token string) teamState {
		t.Helper()
		var result struct {
			Team teamState `json:"team"`
		}
		execute(t, token, `
			query ($id: ID!) {
				team(id: $id) { averageAge leaderboardExcluded members { id countsTowardScores } }
			}
		`, map[string]any{"id": teamID}, &result)
		return result.Team
	}
	countsToward := func(state teamState) map[string]bool {
		out := make(map[string]bool, len(state.Members))
		for _, m := range state.Members {
			out[m.ID] = m.CountsTowardScores
		}
		return out
	}
	teamsWithMinMembers := func(t *testing.T, min int) []string {
		t.Helper()
		var result struct {
			Teams struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"teams"`
		}
		execute(t, adminToken, `
			query ($projectId: ID!, $min: Int) {
				teams(filter: { projectId: $projectId, minMembers: $min }, first: 100) { edges { node { id } } }
			}
		`, map[string]any{"projectId": projectID, "min": min}, &result)
		ids := make([]string, 0, len(result.Teams.Edges))
		for _, e := range result.Teams.Edges {
			ids = append(ids, e.Node.ID)
		}
		return ids
	}

	// Placed in the unit by an admin while blocked
	execute(t, adminToken, `
		mutation ($teamId: ID!, $userIds: [ID!]!) { addTeamMembers(teamId: $teamId, userIds: $userIds) { id } }
	`, map[string]any{"teamId": teamID, "userIds": []string{blockedID}}, nil)

	t.Run("an under-13 member who never answered still counts", func(t *testing.T) {
		// (100 + 50) / 2: the child counts, the blocked user does not
		assert.Equal(t, 75, teamBoard(t, "")[teamID], "team average")
		state := team(t, adminToken)
		assert.InDelta(t, float64(leadAge+childAge)/2, state.AverageAge, 0.01, "average age")
		assert.False(t, state.LeaderboardExcluded)
		assert.Equal(t, map[string]bool{leadID: true, childID: true, blockedID: false}, countsToward(state))
		assert.Contains(t, teamsWithMinMembers(t, 2), teamID, "the child is in the member count")
		assert.NotContains(t, teamsWithMinMembers(t, 3), teamID, "the blocked user is not")
	})

	t.Run("another project's required consent changes nothing here", func(t *testing.T) {
		// The lead and the child have not accepted it either; were it applied
		// here, the team would have no counted member and drop off the board
		assert.Contains(t, teamBoard(t, ""), teamID)
		assert.True(t, countsToward(team(t, adminToken))[leadID])
	})

	t.Run("the unit leader sees the same numbers, without the blocked member", func(t *testing.T) {
		admin := team(t, adminToken)
		lead := team(t, leadToken)
		assert.InDelta(t, admin.AverageAge, lead.AverageAge, 0.01)
		assert.Equal(t, map[string]bool{leadID: true, childID: true}, countsToward(lead))
	})

	t.Run("the church filter ignores a blocked member", func(t *testing.T) {
		assert.Contains(t, teamBoard(t, churchID), teamID)
		assert.NotContains(t, teamBoard(t, otherChurchID), teamID,
			"the blocked user is the team's only member from that church")
	})

	t.Run("removing a member re-runs the age rule without the blocked user", func(t *testing.T) {
		// Left: the lead (20) and the blocked user (80). Counted, the average
		// would be 50 and the rule would exclude the team
		execute(t, adminToken, `
			mutation ($teamId: ID!, $userIds: [ID!]!) { removeTeamMembers(teamId: $teamId, userIds: $userIds) { id } }
		`, map[string]any{"teamId": teamID, "userIds": []string{childID}}, nil)
		state := team(t, adminToken)
		assert.InDelta(t, float64(leadAge), state.AverageAge, 0.01, "average age")
		assert.False(t, state.LeaderboardExcluded, "the age rule ignores the blocked user")
		assert.Equal(t, 100, teamBoard(t, "")[teamID], "team average")
	})
}
