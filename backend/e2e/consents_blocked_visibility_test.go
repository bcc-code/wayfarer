package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequiredConsentHidesUserFromOthers covers consent_blocked_user_ids
// (migration 00116): a user who has not accepted a required consent of the
// project is missing from the person leaderboards and the unit view, while
// admins still see them in Team.members.
func TestRequiredConsentHidesUserFromOthers(t *testing.T) {
	const (
		churchID       = "CH01TEST0000000000000BLOCK01"
		projectID      = "PR01TEST0000000000000BLOCK01"
		otherProjectID = "PR01TEST0000000000000BLOCK02"
		teamID         = "TM01TEST0000000000000BLOCK01"
		eventID        = "EV01TEST0000000000000BLOCK01"

		acceptedID   = "US01TEST0000000000000BLKACPT" // accepted -> visible
		unansweredID = "US01TEST0000000000000BLKNONE" // never answered -> hidden
		rejectedID   = "US01TEST0000000000000BLKREJT" // rejected -> hidden
		childID      = "US01TEST0000000000000BLKKID1" // under 13 -> exempt, visible
		teenID       = "US01TEST0000000000000BLKTEEN" // turns 13 this year -> not exempt, hidden
	)

	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))

	adult := time.Now().AddDate(-25, 0, 0)
	child := time.Now().AddDate(-10, 0, 0)
	// Age is the year difference (calculateAge), so this is 13 all year
	teen := time.Date(time.Now().Year()-13, time.December, 31, 0, 0, 0, 0, time.UTC)

	require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Block Church", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Block Project"))
	require.NoError(t, dbMgr.CreateTestProject(ctx, otherProjectID, "Other Project"))
	require.NoError(t, setCurrentProject(ctx, dbMgr, projectID))
	require.NoError(t, dbMgr.CreateTestTeam(ctx, teamID, "Block Team", projectID))
	require.NoError(t, dbMgr.CreateTestEvent(ctx, eventID, "Block Event", projectID))

	// Distinct scores, so a hidden user left in the ranking shows up as a gap
	users := []struct {
		id        string
		birthdate time.Time
		points    int
	}{
		{acceptedID, adult, 500},
		{unansweredID, adult, 400},
		{rejectedID, adult, 300},
		{teenID, teen, 200},
		{childID, child, 100},
	}
	for _, u := range users {
		require.NoError(t, dbMgr.CreateTestUser(ctx, u.id, u.id, "MALE", u.birthdate, churchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, u.id, projectID))
		require.NoError(t, dbMgr.EnrollUserInEvent(ctx, u.id, eventID))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, u.id, teamID))
		// The persons board separately requires leaderboard_consent; give it
		// to everyone so only the required consent decides visibility
		require.NoError(t, dbMgr.AddLeaderboardConsent(ctx, u.id))
		require.NoError(t, dbMgr.AddScoreForUserEvent(ctx, u.id, projectID, eventID, u.points))
	}
	insertConsent := func(t *testing.T, key string, version int, projectID *string, isRequired, isRemote bool) string {
		t.Helper()
		id := ulid.NewConsentID()
		_, err := dbMgr.DB.Pool.Exec(ctx, `
			INSERT INTO consents (id, key, version, title, short_text, body, published_at, is_remote, project_id, is_required)
			VALUES ($1, $2, $3, 'T', 'S', 'B', now() - interval '1 minute', $4, $5, $6)
		`, id, key, version, isRemote, projectID, isRequired)
		require.NoError(t, err)
		return id
	}
	setAction := func(t *testing.T, userID, consentID, key, action string) {
		t.Helper()
		_, err := dbMgr.DB.Pool.Exec(ctx, `
			INSERT INTO user_consent_history (id, user_id, consent_id, consent_key, action, occurred_at)
			VALUES ($1, $2, $3, $4, $5, now())
		`, ulid.NewUserConsentHistoryID(), userID, consentID, key, action)
		require.NoError(t, err)
	}

	pid := projectID
	participation := insertConsent(t, "participation", 1, &pid, true, false)
	setAction(t, acceptedID, participation, "participation", "ACCEPTED")
	// Rejected after accepting: the latest action decides
	setAction(t, rejectedID, participation, "participation", "ACCEPTED")
	time.Sleep(10 * time.Millisecond)
	setAction(t, rejectedID, participation, "participation", "REJECTED")

	router, testCache, cleanup, err := testutil.SetupTestServerWithCache(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	// The admin is not a project member; Team.members checks roles in the DB
	const adminID = "US01TEST0000000000000BLKADMN"
	require.NoError(t, dbMgr.CreateTestUser(ctx, adminID, "Admin", "MALE", adult, churchID))
	require.NoError(t, dbMgr.AssignRole(ctx, adminID, testutil.RoleAdmin))
	adminToken, err := testutil.GenerateAdminToken(adminID)
	require.NoError(t, err)

	// The church admin of the members' church, who places them into units
	const churchAdminID = "US01TEST0000000000000BLKCHAD"
	require.NoError(t, dbMgr.CreateTestUser(ctx, churchAdminID, "Church Admin", "MALE", adult, churchID))
	cid := churchID
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, churchAdminID, testutil.RoleChurchAdmin, &cid, nil, nil))
	churchAdminToken, err := testutil.GenerateChurchAdminToken(churchAdminID)
	require.NoError(t, err)

	// The viewer leads the unit: a participant, so hidden members stay hidden
	tid := teamID
	require.NoError(t, dbMgr.AssignRoleWithScope(ctx, acceptedID, testutil.RoleTeamLead, nil, nil, &tid))
	viewerToken, err := testutil.GenerateUserToken(acceptedID)
	require.NoError(t, err)

	type ranked struct {
		ID   string `json:"id"`
		Rank *int   `json:"rank"`
	}
	ranksByID := func(entries []ranked) map[string]int {
		out := make(map[string]int, len(entries))
		for _, e := range entries {
			require.NotNil(t, e.Rank, "entry %s has no rank", e.ID)
			out[e.ID] = *e.Rank
		}
		return out
	}

	projectBoard := func(t *testing.T) (map[string]int, int) {
		t.Helper()
		resp := client.WithAuth(viewerToken).MustExecute(t, `
			query ($id: ID!) {
				project(id: $id) {
					leaderboard(entityType: PERSONS, first: 100) {
						edges { node { id rank } }
						totalCount
					}
				}
			}
		`, map[string]any{"id": projectID})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
		var result struct {
			Project struct {
				Leaderboard struct {
					Edges []struct {
						Node ranked `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"leaderboard"`
			} `json:"project"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		nodes := make([]ranked, len(result.Project.Leaderboard.Edges))
		for i, e := range result.Project.Leaderboard.Edges {
			nodes[i] = e.Node
		}
		return ranksByID(nodes), result.Project.Leaderboard.TotalCount
	}

	eventBoard := func(t *testing.T) map[string]int {
		t.Helper()
		resp := client.WithAuth(viewerToken).MustExecute(t, `
			query ($id: ID!) {
				event(id: $id) {
					leaderboard(entityType: PERSONS, first: 100) {
						edges { node { id rank } }
					}
				}
			}
		`, map[string]any{"id": eventID})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
		var result struct {
			Event struct {
				Leaderboard struct {
					Edges []struct {
						Node ranked `json:"node"`
					} `json:"edges"`
				} `json:"leaderboard"`
			} `json:"event"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		nodes := make([]ranked, len(result.Event.Leaderboard.Edges))
		for i, e := range result.Event.Leaderboard.Edges {
			nodes[i] = e.Node
		}
		return ranksByID(nodes)
	}

	unit := func(t *testing.T, token string) (map[string]int, []string) {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, `
			query ($id: ID!) {
				team(id: $id) {
					memberLeaderboard { id rank }
				}
			}
		`, map[string]any{"id": teamID})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
		var result struct {
			Team struct {
				MemberLeaderboard []ranked `json:"memberLeaderboard"`
			} `json:"team"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		ids := make([]string, len(result.Team.MemberLeaderboard))
		for i, e := range result.Team.MemberLeaderboard {
			ids[i] = e.ID
		}
		return ranksByID(result.Team.MemberLeaderboard), ids
	}

	// Only the accepted adult and the exempt child are visible, ranked 1 and 2
	// with no gap where the hidden users scored in between
	visible := map[string]int{acceptedID: 1, childID: 2}

	t.Run("person leaderboard hides users who have not accepted", func(t *testing.T) {
		ranks, total := projectBoard(t)
		assert.Equal(t, visible, ranks)
		assert.Equal(t, 2, total, "totalCount must not count hidden users")
	})

	t.Run("event person leaderboard hides them too", func(t *testing.T) {
		assert.Equal(t, visible, eventBoard(t))
	})

	t.Run("unit view hides them, for admins as well", func(t *testing.T) {
		ranks, _ := unit(t, viewerToken)
		assert.Equal(t, visible, ranks)

		adminRanks, _ := unit(t, adminToken)
		assert.Equal(t, visible, adminRanks)
	})

	teamMembers := func(t *testing.T, token string) []string {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, `
			query ($id: ID!) {
				team(id: $id) { members { id } }
			}
		`, map[string]any{"id": teamID})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
		var result struct {
			Team struct {
				Members []struct {
					ID string `json:"id"`
				} `json:"members"`
			} `json:"team"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		ids := make([]string, len(result.Team.Members))
		for i, m := range result.Team.Members {
			ids[i] = m.ID
		}
		return ids
	}

	usersInTeam := func(t *testing.T, token string) ([]string, int) {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, `
			query ($teamId: ID!) {
				users(filter: { teamId: $teamId }, first: 100) {
					edges { node { id } }
					totalCount
				}
			}
		`, map[string]any{"teamId": teamID})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())
		var result struct {
			Users struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"users"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		ids := make([]string, len(result.Users.Edges))
		for i, e := range result.Users.Edges {
			ids[i] = e.Node.ID
		}
		return ids, result.Users.TotalCount
	}

	allMembers := []string{acceptedID, unansweredID, rejectedID, childID, teenID}
	visibleMembers := []string{acceptedID, childID}

	t.Run("admins still see every member in Team.members", func(t *testing.T) {
		assert.ElementsMatch(t, allMembers, teamMembers(t, adminToken))
	})

	t.Run("the church admin who placed them sees every member", func(t *testing.T) {
		assert.ElementsMatch(t, allMembers, teamMembers(t, churchAdminToken))

		ids, total := usersInTeam(t, churchAdminToken)
		assert.ElementsMatch(t, allMembers, ids)
		assert.Equal(t, len(allMembers), total)
	})

	t.Run("the unit leader does not see hidden members", func(t *testing.T) {
		assert.ElementsMatch(t, visibleMembers, teamMembers(t, viewerToken))

		ids, total := usersInTeam(t, viewerToken)
		assert.ElementsMatch(t, visibleMembers, ids)
		assert.Equal(t, len(visibleMembers), total, "totalCount must not count hidden members")
	})

	t.Run("a required consent of another project does not hide anyone here", func(t *testing.T) {
		other := otherProjectID
		insertConsent(t, "other_participation", 1, &other, true, false)
		testCache.InvalidateLeaderboardScores()

		ranks, _ := projectBoard(t)
		assert.Equal(t, visible, ranks)
	})

	t.Run("a required remote consent does not hide anyone", func(t *testing.T) {
		insertConsent(t, "remote_required", 1, nil, true, true)
		testCache.InvalidateLeaderboardScores()

		ranks, _ := projectBoard(t)
		assert.Equal(t, visible, ranks)
	})

	t.Run("only the latest published version decides whether a consent is required", func(t *testing.T) {
		insertConsent(t, "relaxed", 1, nil, true, false)
		insertConsent(t, "relaxed", 2, nil, false, false)
		testCache.InvalidateLeaderboardScores()

		ranks, _ := projectBoard(t)
		assert.Equal(t, visible, ranks)
	})

	t.Run("accepting makes the user visible straight away", func(t *testing.T) {
		// Warm the caches with the user hidden
		ranks, _ := projectBoard(t)
		require.NotContains(t, ranks, unansweredID)
		unitRanks, _ := unit(t, viewerToken)
		require.NotContains(t, unitRanks, unansweredID)

		unansweredToken, err := testutil.GenerateUserToken(unansweredID)
		require.NoError(t, err)
		resp := client.WithAuth(unansweredToken).MustExecute(t, `
			mutation ($id: ID!) { acceptConsent(consentId: $id) { id } }
		`, map[string]any{"id": participation})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		want := map[string]int{acceptedID: 1, unansweredID: 2, childID: 3}
		ranks, total := projectBoard(t)
		assert.Equal(t, want, ranks)
		assert.Equal(t, 3, total)

		unitRanks, _ = unit(t, viewerToken)
		assert.Equal(t, want, unitRanks)
	})

	// Answers are per version (00117): a user has not consented to a version
	// they have not seen
	var participationV2 string
	t.Run("a new version hides users who accepted the previous one", func(t *testing.T) {
		participationV2 = insertConsent(t, "participation", 2, &pid, true, false)
		testCache.InvalidateLeaderboardScores()

		ranks, total := projectBoard(t)
		assert.Equal(t, map[string]int{childID: 1}, ranks)
		assert.Equal(t, 1, total)
	})

	t.Run("answering the previous version is refused", func(t *testing.T) {
		resp := client.WithAuth(viewerToken).MustExecute(t, `
			mutation ($id: ID!) { acceptConsent(consentId: $id) { id } }
		`, map[string]any{"id": participation})
		require.True(t, resp.HasErrors())
		assert.Contains(t, resp.ErrorMessage(), "latest version")
	})

	t.Run("accepting the new version shows them again", func(t *testing.T) {
		require.NotEmpty(t, participationV2)
		resp := client.WithAuth(viewerToken).MustExecute(t, `
			mutation ($id: ID!) { acceptConsent(consentId: $id) { id } }
		`, map[string]any{"id": participationV2})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		ranks, _ := projectBoard(t)
		assert.Equal(t, map[string]int{acceptedID: 1, childID: 2}, ranks)
	})
}
