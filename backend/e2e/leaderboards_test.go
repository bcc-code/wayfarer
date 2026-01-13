package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test entity IDs with known ages
const (
	testChurchID  = "CH01TEST00000000000000000001"
	testProjectID = "PR01TEST00000000000000000001"
	testTeamID    = "TM01TEST00000000000000000001"

	// Users with various ages for filtering tests
	userYoung16ID   = "US01TEST0000000000000YOUNG16"
	userYoung17ID   = "US01TEST0000000000000YOUNG17"
	userAdult20ID   = "US01TEST0000000000000ADULT20"
	userAdult22ID   = "US01TEST0000000000000ADULT22"
	userAdult25ID   = "US01TEST0000000000000ADULT25"
	userSenior30ID  = "US01TEST000000000000SENIOR30"
	userSenior40ID  = "US01TEST000000000000SENIOR40"
	userElder50ID   = "US01TEST0000000000000ELDER50"
	userNoConsentID = "US01TEST00000000000NOCONSENT"
	userNoScoreID   = "US01TEST000000000000NOSCORE0"

	// Special user for year-based age calculation test
	// Born December 27, 1988 - will be 38 by year calculation (2026-1988) but 37 by AGE function
	userBornDec1988ID = "US01TEST0000000000000DEC1988"
)

// leaderboardResult is the common response type for leaderboard queries
type leaderboardResult struct {
	Project struct {
		Leaderboard struct {
			Edges []struct {
				Node struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Score int    `json:"score"`
				} `json:"node"`
			} `json:"edges"`
			TotalCount int `json:"totalCount"`
		} `json:"leaderboard"`
	} `json:"project"`
}

// getUserIDs extracts user IDs from leaderboard result
func (r *leaderboardResult) getUserIDs() []string {
	ids := make([]string, len(r.Project.Leaderboard.Edges))
	for i, e := range r.Project.Leaderboard.Edges {
		ids[i] = e.Node.ID
	}
	return ids
}

// leaderboardMeResult includes the "me" field for testing
type leaderboardMeResult struct {
	Project struct {
		Leaderboard struct {
			Edges []struct {
				Node struct {
					ID    string   `json:"id"`
					Name  string   `json:"name"`
					Score int      `json:"score"`
					Rank  *int     `json:"rank"`
					Tags  []string `json:"tags"`
				} `json:"node"`
			} `json:"edges"`
			TotalCount int `json:"totalCount"`
			Me         *struct {
				ID    string   `json:"id"`
				Name  string   `json:"name"`
				Score int      `json:"score"`
				Rank  *int     `json:"rank"`
				Tags  []string `json:"tags"`
			} `json:"me"`
		} `json:"leaderboard"`
	} `json:"project"`
}

func TestLeaderboardAgeFiltering(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for fresh test data
	require.NoError(t, dbMgr.Clean(ctx))

	// Setup test data with known ages
	setupLeaderboardTestData(t, ctx, dbMgr)

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	// Generate admin token for queries
	adminToken, err := testutil.GenerateAdminToken(userAdult20ID)
	require.NoError(t, err)

	// Standard query used across tests
	const leaderboardQuery = `
		query GetLeaderboard($projectId: ID!, $filter: LeaderboardFilter) {
			project(id: $projectId) {
				leaderboard(entityType: PERSONS, filter: $filter, first: 100) {
					edges {
						node {
							id
							name
							score
						}
					}
					totalCount
				}
			}
		}
	`

	// ==================== A. Consent-based tests ====================

	t.Run("consent-based: no age filter returns all consented users", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// All users with consent should appear (excluding userNoConsentID)
		assert.Equal(t, 8, result.Project.Leaderboard.TotalCount)
		userIDs := result.getUserIDs()
		assert.NotContains(t, userIDs, userNoConsentID, "user without consent should not appear")
	})

	t.Run("consent-based: age filter 18-25 returns only adults in range", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"ageRange": map[string]any{
					"min": 18,
					"max": 25,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Only Adult20, Adult22, Adult25 should appear
		assert.Equal(t, 3, result.Project.Leaderboard.TotalCount)
		userIDs := result.getUserIDs()
		assert.Contains(t, userIDs, userAdult20ID)
		assert.Contains(t, userIDs, userAdult22ID)
		assert.Contains(t, userIDs, userAdult25ID)
		assert.NotContains(t, userIDs, userYoung16ID)
		assert.NotContains(t, userIDs, userYoung17ID)
		assert.NotContains(t, userIDs, userSenior30ID)
	})

	t.Run("consent-based: age filter with no matches returns empty", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"ageRange": map[string]any{
					"min": 60,
					"max": 70,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, 0, result.Project.Leaderboard.TotalCount)
		assert.Empty(t, result.Project.Leaderboard.Edges)
	})

	t.Run("consent-based: user without consent is excluded", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"ageRange": map[string]any{
					"min": 26,
					"max": 40,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		userIDs := result.getUserIDs()
		// userNoConsentID is age 35, so should match age filter but NOT appear due to no consent
		assert.NotContains(t, userIDs, userNoConsentID)
		// Senior30 and Senior40 should appear (they have consent)
		assert.Contains(t, userIDs, userSenior30ID)
		assert.Contains(t, userIDs, userSenior40ID)
	})

	// ==================== B. Team-filtered tests ====================

	t.Run("team-filtered: age filter 18-25 with team filter", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
				"ageRange": map[string]any{
					"min": 18,
					"max": 25,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Adult20, Adult22, Adult25 are in team and in age range
		assert.Equal(t, 3, result.Project.Leaderboard.TotalCount)
		userIDs := result.getUserIDs()
		assert.Contains(t, userIDs, userAdult20ID)
		assert.Contains(t, userIDs, userAdult22ID)
		assert.Contains(t, userIDs, userAdult25ID)
	})

	t.Run("team-filtered: includes non-consented user when using team filter", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
				"ageRange": map[string]any{
					"min": 30,
					"max": 40,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		userIDs := result.getUserIDs()
		// userNoConsentID (age 35) should appear when team-filtered (bypasses consent)
		assert.Contains(t, userIDs, userNoConsentID, "non-consented user should appear when team-filtered")
		assert.Contains(t, userIDs, userSenior30ID)
		assert.Contains(t, userIDs, userSenior40ID)
	})

	t.Run("team-filtered: exact age boundary at max", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
				"ageRange": map[string]any{
					"min": 18,
					"max": 25,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		userIDs := result.getUserIDs()
		// User aged exactly 25 should be included in 18-25 range
		assert.Contains(t, userIDs, userAdult25ID, "user aged exactly 25 should be in 18-25 range")
	})

	// ==================== C. Combined filter tests ====================

	t.Run("combined: team + wider age range filter", func(t *testing.T) {
		// Use team filter to bypass consent requirement
		// Test that wider age range (13-50) includes all adults but not seniors over 50
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
				"ageRange": map[string]any{
					"min": 13,
					"max": 50,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// All team members aged 13-50 should appear (everyone in this test)
		// Young16, Young17, Adult20, Adult22, Adult25, Senior30, NoConsent, Senior40, Elder50
		assert.Equal(t, 9, result.Project.Leaderboard.TotalCount)
	})

	t.Run("combined: team + narrow age range excludes users", func(t *testing.T) {
		// Test narrow age range that should only include a few users
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
				"ageRange": map[string]any{
					"min": 20,
					"max": 22,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Only Adult20 (20) and Adult22 (22) should appear
		assert.Equal(t, 2, result.Project.Leaderboard.TotalCount)
		userIDs := result.getUserIDs()
		assert.Contains(t, userIDs, userAdult20ID)
		assert.Contains(t, userIDs, userAdult22ID)

		// Others should not appear
		assert.NotContains(t, userIDs, userYoung16ID)
		assert.NotContains(t, userIDs, userYoung17ID)
		assert.NotContains(t, userIDs, userAdult25ID)
		assert.NotContains(t, userIDs, userSenior30ID)
	})
}

func TestLeaderboardMeField(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for fresh test data
	require.NoError(t, dbMgr.Clean(ctx))

	// Setup test data with known ages
	setupLeaderboardTestData(t, ctx, dbMgr)

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	// Query that includes the "me" field
	const meQuery = `
		query GetLeaderboard($projectId: ID!, $filter: LeaderboardFilter) {
			project(id: $projectId) {
				leaderboard(entityType: PERSONS, filter: $filter, first: 100) {
					edges {
						node {
							id
							name
							score
							rank
							tags
						}
					}
					totalCount
					me {
						id
						name
						score
						rank
						tags
					}
				}
			}
		}
	`

	t.Run("me field returns current user's entry", func(t *testing.T) {
		// Query as userAdult20 - should see themselves in "me"
		userToken, err := testutil.GenerateUserToken(userAdult20ID)
		require.NoError(t, err)

		resp := client.WithAuth(userToken).MustExecute(t, meQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardMeResult
		require.NoError(t, resp.UnmarshalData(&result))

		// "me" should be populated with userAdult20
		require.NotNil(t, result.Project.Leaderboard.Me, "me field should not be nil")
		assert.Equal(t, userAdult20ID, result.Project.Leaderboard.Me.ID)
		assert.Equal(t, "Adult20", result.Project.Leaderboard.Me.Name)
		assert.Equal(t, 100, result.Project.Leaderboard.Me.Score)
	})

	t.Run("me field includes correct rank", func(t *testing.T) {
		userToken, err := testutil.GenerateUserToken(userAdult20ID)
		require.NoError(t, err)

		resp := client.WithAuth(userToken).MustExecute(t, meQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardMeResult
		require.NoError(t, resp.UnmarshalData(&result))

		require.NotNil(t, result.Project.Leaderboard.Me)
		require.NotNil(t, result.Project.Leaderboard.Me.Rank, "rank should not be nil")
		// Rank should be a valid positive number (all users have same score, so rank depends on name order)
		assert.Greater(t, *result.Project.Leaderboard.Me.Rank, 0)
	})

	t.Run("me field includes ME tag", func(t *testing.T) {
		userToken, err := testutil.GenerateUserToken(userAdult20ID)
		require.NoError(t, err)

		resp := client.WithAuth(userToken).MustExecute(t, meQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardMeResult
		require.NoError(t, resp.UnmarshalData(&result))

		require.NotNil(t, result.Project.Leaderboard.Me)
		assert.Contains(t, result.Project.Leaderboard.Me.Tags, "ME", "me entry should have ME tag")
	})

	t.Run("me field is null when user has no score", func(t *testing.T) {
		// userNoScoreID is enrolled but has no score
		userToken, err := testutil.GenerateUserToken(userNoScoreID)
		require.NoError(t, err)

		resp := client.WithAuth(userToken).MustExecute(t, meQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardMeResult
		require.NoError(t, resp.UnmarshalData(&result))

		// "me" should be null because user has no score
		assert.Nil(t, result.Project.Leaderboard.Me, "me should be nil for user with no score")
	})

	t.Run("me field is null when age filter excludes current user", func(t *testing.T) {
		// Query as userAdult20 (age 20) with age filter 30-40 that excludes them
		userToken, err := testutil.GenerateUserToken(userAdult20ID)
		require.NoError(t, err)

		resp := client.WithAuth(userToken).MustExecute(t, meQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
				"ageRange": map[string]any{
					"min": 30,
					"max": 40,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardMeResult
		require.NoError(t, resp.UnmarshalData(&result))

		// "me" should be null because the age filter excludes userAdult20
		assert.Nil(t, result.Project.Leaderboard.Me, "me should be nil when age filter excludes current user")
		// But the leaderboard should still have results
		assert.Greater(t, result.Project.Leaderboard.TotalCount, 0)
	})

	t.Run("me field available with pagination", func(t *testing.T) {
		// Even with limited pagination, "me" should still be available
		userToken, err := testutil.GenerateUserToken(userElder50ID)
		require.NoError(t, err)

		// Query with very small pagination (first: 2) - Elder50 might not be in first 2
		const paginatedQuery = `
			query GetLeaderboard($projectId: ID!, $filter: LeaderboardFilter) {
				project(id: $projectId) {
					leaderboard(entityType: PERSONS, filter: $filter, first: 2) {
						edges {
							node {
								id
							}
						}
						totalCount
						me {
							id
							name
							rank
						}
					}
				}
			}
		`

		resp := client.WithAuth(userToken).MustExecute(t, paginatedQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardMeResult
		require.NoError(t, resp.UnmarshalData(&result))

		// "me" should still be populated even if not in first 2 edges
		require.NotNil(t, result.Project.Leaderboard.Me, "me should be available even with pagination")
		assert.Equal(t, userElder50ID, result.Project.Leaderboard.Me.ID)
	})
}

// TestLeaderboardAgeCalculationYearBased tests that age is calculated using
// current_year - birth_year, NOT PostgreSQL's AGE function which considers
// whether the birthday has passed this year.
//
// Bug: Person born December 27, 1988 was showing up in results for age 20-37
// because AGE() calculated age as 37 (birthday not yet passed in 2026).
// Expected: Age should be 2026-1988=38, so they should NOT appear in 20-37 range.
func TestLeaderboardAgeCalculationYearBased(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for fresh test data
	require.NoError(t, dbMgr.Clean(ctx))

	// Create test church
	require.NoError(t, dbMgr.CreateTestChurch(ctx, testChurchID, "Test Church", "NO", "S"))

	// Create test project
	require.NoError(t, dbMgr.CreateTestProject(ctx, testProjectID, "Test Project"))

	// Update settings to point to our test project
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'`, testProjectID)
	require.NoError(t, err)

	// Create test team
	require.NoError(t, dbMgr.CreateTestTeam(ctx, testTeamID, "Test Team", testProjectID))

	// Create user born December 27, 1988
	// Year-based age: 2026 - 1988 = 38
	// AGE() function age: 37 (birthday December 27 hasn't occurred yet if test runs before Dec 27)
	birthdate := time.Date(1988, time.December, 27, 0, 0, 0, 0, time.UTC)
	require.NoError(t, dbMgr.CreateTestUser(ctx, userBornDec1988ID, "BornDec1988", "MALE", birthdate, testChurchID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userBornDec1988ID, testProjectID))
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userBornDec1988ID, testProjectID, 100))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userBornDec1988ID, testTeamID))
	require.NoError(t, dbMgr.AddLeaderboardConsent(ctx, userBornDec1988ID))

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(userBornDec1988ID)
	require.NoError(t, err)

	const leaderboardQuery = `
		query GetLeaderboard($projectId: ID!, $filter: LeaderboardFilter) {
			project(id: $projectId) {
				leaderboard(entityType: PERSONS, filter: $filter, first: 100) {
					edges {
						node {
							id
							name
							score
						}
					}
					totalCount
				}
			}
		}
	`

	t.Run("user born Dec 1988 should NOT appear in age range 20-37 (year-based age is 38)", func(t *testing.T) {
		// Filter for ages 20-37
		// Year-based age of user born 1988: 2026 - 1988 = 38
		// If age calculation is correct, user should NOT appear (38 > 37)
		// If AGE() function is used, user might appear (AGE calculates 37 before Dec 27)
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
				"ageRange": map[string]any{
					"min": 20,
					"max": 37,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		userIDs := result.getUserIDs()
		// User born 1988 has year-based age 38, should NOT be in 20-37 range
		assert.NotContains(t, userIDs, userBornDec1988ID,
			"user born Dec 1988 should NOT appear in age 20-37 range; year-based age is 38, not 37")
		assert.Equal(t, 0, result.Project.Leaderboard.TotalCount,
			"no users should match age 20-37 filter")
	})

	t.Run("user born Dec 1988 SHOULD appear in age range 20-38 (year-based age is 38)", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"teamId": testTeamID,
				"ageRange": map[string]any{
					"min": 20,
					"max": 38,
				},
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		userIDs := result.getUserIDs()
		// User born 1988 has year-based age 38, should be in 20-38 range
		assert.Contains(t, userIDs, userBornDec1988ID,
			"user born Dec 1988 SHOULD appear in age 20-38 range; year-based age is exactly 38")
	})
}

// TestLeaderboardDenseRanking tests that entities with the same score share the same rank
// and that ranks are consecutive (no gaps). This tests the DENSE_RANK() behavior.
// Uses TEAMS entity type to avoid PERSONS visibility limit filtering.
func TestLeaderboardDenseRanking(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for fresh test data
	require.NoError(t, dbMgr.Clean(ctx))

	// Create test church
	require.NoError(t, dbMgr.CreateTestChurch(ctx, testChurchID, "Test Church", "NO", "S"))

	// Create test project
	require.NoError(t, dbMgr.CreateTestProject(ctx, testProjectID, "Test Project"))

	// Update settings to point to our test project
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'`, testProjectID)
	require.NoError(t, err)

	now := time.Now()
	birthdate := now.AddDate(-25, 0, 0)

	// Create teams with specific scores to test ranking:
	// TeamAlpha: 100 pts (rank 1)
	// TeamBeta: 100 pts (rank 1 - same as Alpha)
	// TeamGamma: 50 pts (rank 2 - consecutive, no gap)
	// TeamDelta: 50 pts (rank 2 - same as Gamma)
	// TeamEpsilon: 25 pts (rank 3 - consecutive)
	teams := []struct {
		teamID string
		name   string
		userID string
		score  int
	}{
		{"TM01TEST00000000000000ALPHA0", "TeamAlpha", "US01TEST00000000000000ALPHA0", 100},
		{"TM01TEST00000000000000BETA00", "TeamBeta", "US01TEST00000000000000BETA00", 100},
		{"TM01TEST00000000000000GAMMA0", "TeamGamma", "US01TEST00000000000000GAMMA0", 50},
		{"TM01TEST00000000000000DELTA0", "TeamDelta", "US01TEST00000000000000DELTA0", 50},
		{"TM01TEST0000000000000EPSILON", "TeamEpsilon", "US01TEST0000000000000EPSILON", 25},
	}

	for _, team := range teams {
		// Create team
		require.NoError(t, dbMgr.CreateTestTeam(ctx, team.teamID, team.name, testProjectID))

		// Create a user for this team (needed for score calculation)
		require.NoError(t, dbMgr.CreateTestUser(ctx, team.userID, team.name+"User", "MALE", birthdate, testChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, team.userID, testProjectID))
		require.NoError(t, dbMgr.AddUserToTeam(ctx, team.userID, team.teamID))
		require.NoError(t, dbMgr.AddScoreForUser(ctx, team.userID, testProjectID, team.score))
	}

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(teams[0].userID)
	require.NoError(t, err)

	const rankingQuery = `
		query GetLeaderboard($projectId: ID!) {
			project(id: $projectId) {
				leaderboard(entityType: TEAMS, first: 100) {
					edges {
						node {
							id
							name
							score
							rank
						}
					}
					totalCount
				}
			}
		}
	`

	t.Run("teams with same score share the same rank (dense ranking)", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, rankingQuery, map[string]any{
			"projectId": testProjectID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardMeResult
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, 5, result.Project.Leaderboard.TotalCount)

		// Build a map of name -> rank for easier assertions
		ranks := make(map[string]int)
		for _, edge := range result.Project.Leaderboard.Edges {
			require.NotNil(t, edge.Node.Rank, "rank should not be nil for %s", edge.Node.Name)
			ranks[edge.Node.Name] = *edge.Node.Rank
		}

		// TeamAlpha and TeamBeta both have 100 pts, should share rank 1
		assert.Equal(t, 1, ranks["TeamAlpha"], "TeamAlpha with 100 pts should be rank 1")
		assert.Equal(t, 1, ranks["TeamBeta"], "TeamBeta with 100 pts should be rank 1 (same as Alpha)")

		// TeamGamma and TeamDelta both have 50 pts, should share rank 2 (dense - no gap)
		assert.Equal(t, 2, ranks["TeamGamma"], "TeamGamma with 50 pts should be rank 2 (dense ranking)")
		assert.Equal(t, 2, ranks["TeamDelta"], "TeamDelta with 50 pts should be rank 2 (same as Gamma)")

		// TeamEpsilon has 25 pts, should be rank 3 (consecutive after rank 2)
		assert.Equal(t, 3, ranks["TeamEpsilon"], "TeamEpsilon with 25 pts should be rank 3 (consecutive)")
	})

	t.Run("teams with same rank are sorted alphabetically by name", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, rankingQuery, map[string]any{
			"projectId": testProjectID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result leaderboardMeResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Extract names in order
		names := make([]string, len(result.Project.Leaderboard.Edges))
		for i, edge := range result.Project.Leaderboard.Edges {
			names[i] = edge.Node.Name
		}

		// Expected order: TeamAlpha, TeamBeta (both rank 1, alpha), TeamDelta, TeamGamma (both rank 2, alpha), TeamEpsilon
		expected := []string{"TeamAlpha", "TeamBeta", "TeamDelta", "TeamGamma", "TeamEpsilon"}
		assert.Equal(t, expected, names, "leaderboard should be sorted by rank, then alphabetically by name")
	})
}

// setupLeaderboardTestData creates test users with known ages
func setupLeaderboardTestData(t *testing.T, ctx context.Context, dbMgr *testutil.TestDBManager) {
	t.Helper()

	now := time.Now()

	// Create test church
	require.NoError(t, dbMgr.CreateTestChurch(ctx, testChurchID, "Test Church", "NO", "S"))

	// Create test project
	require.NoError(t, dbMgr.CreateTestProject(ctx, testProjectID, "Test Project"))

	// Update settings to point to our test project (settings table is not truncated)
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'`, testProjectID)
	require.NoError(t, err)

	// Create test team
	require.NoError(t, dbMgr.CreateTestTeam(ctx, testTeamID, "Test Team", testProjectID))

	// Create users with specific ages
	users := []struct {
		id        string
		name      string
		gender    string
		age       int
		hasTeam   bool
		noConsent bool
		noScore   bool
	}{
		{userYoung16ID, "Young16", "MALE", 16, true, false, false},
		{userYoung17ID, "Young17", "FEMALE", 17, true, false, false},
		{userAdult20ID, "Adult20", "MALE", 20, true, false, false},
		{userAdult22ID, "Adult22", "FEMALE", 22, true, false, false},
		{userAdult25ID, "Adult25", "MALE", 25, true, false, false},
		{userSenior30ID, "Senior30", "MALE", 30, true, false, false},
		{userSenior40ID, "Senior40", "FEMALE", 40, true, false, false},
		{userElder50ID, "Elder50", "MALE", 50, true, false, false},
		{userNoConsentID, "NoConsent", "MALE", 35, true, true, false},
		{userNoScoreID, "NoScore", "MALE", 25, true, false, true},
	}

	for _, u := range users {
		birthdate := now.AddDate(-u.age, 0, 0)
		require.NoError(t, dbMgr.CreateTestUser(ctx, u.id, u.name, u.gender, birthdate, testChurchID))
		require.NoError(t, dbMgr.EnrollUserInProject(ctx, u.id, testProjectID))

		if !u.noScore {
			require.NoError(t, dbMgr.AddScoreForUser(ctx, u.id, testProjectID, 100))
		}

		if u.hasTeam {
			require.NoError(t, dbMgr.AddUserToTeam(ctx, u.id, testTeamID))
		}

		if !u.noConsent {
			require.NoError(t, dbMgr.AddLeaderboardConsent(ctx, u.id))
		}
	}
}
