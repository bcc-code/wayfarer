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

	t.Run("teams with same rank are sorted by most recent score first, then alphabetically", func(t *testing.T) {
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

		// Expected order: sorted by rank, then by most recent score (last_score_at DESC), then alphabetically
		// Since teams are created in order (Alpha, Beta, Gamma, Delta, Epsilon), the later ones have more recent scores
		// Rank 1 (100 pts): TeamBeta (most recent), TeamAlpha
		// Rank 2 (50 pts): TeamDelta (most recent), TeamGamma
		// Rank 3 (25 pts): TeamEpsilon
		expected := []string{"TeamBeta", "TeamAlpha", "TeamDelta", "TeamGamma", "TeamEpsilon"}
		assert.Equal(t, expected, names, "leaderboard should be sorted by rank, then by most recent score, then alphabetically")
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

// ==================== TEAMS Leaderboard Church Filtering Tests ====================

// Test entity IDs for church filtering tests (must be exactly 28 chars: 2-char prefix + 26-char ULID)
const (
	testChurch2ID      = "CH01TEST00000000000000000002"
	teamChurch1OnlyID  = "TM01TEST000000000000CHURCH1A"
	teamChurch2OnlyID  = "TM01TEST000000000000CHURCH2A"
	teamMixedID        = "TM01TEST000000000000000MIXED"
	userChurch1AID     = "US01TEST000000000000CHURCH1A"
	userChurch1BID     = "US01TEST000000000000CHURCH1B"
	userChurch2AID     = "US01TEST000000000000CHURCH2A"
	userChurch2BID     = "US01TEST000000000000CHURCH2B"
	userMixedChurch1ID = "US01TEST00000000000MIXEDCH1A"
	userMixedChurch2ID = "US01TEST00000000000MIXEDCH2A"
)

// teamLeaderboardResult is the response type for team leaderboard queries
type teamLeaderboardResult struct {
	Project struct {
		Leaderboard struct {
			Edges []struct {
				Node struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Score int    `json:"score"`
					Rank  *int   `json:"rank"`
				} `json:"node"`
			} `json:"edges"`
			TotalCount int `json:"totalCount"`
			Me         *struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Score int    `json:"score"`
				Rank  *int   `json:"rank"`
			} `json:"me"`
		} `json:"leaderboard"`
	} `json:"project"`
}

// getTeamIDs extracts team IDs from team leaderboard result
func (r *teamLeaderboardResult) getTeamIDs() []string {
	ids := make([]string, len(r.Project.Leaderboard.Edges))
	for i, e := range r.Project.Leaderboard.Edges {
		ids[i] = e.Node.ID
	}
	return ids
}

// TestTeamLeaderboardChurchFiltering tests that team leaderboards can be filtered by church
// Filter logic: include team if ANY member is from the specified church
func TestTeamLeaderboardChurchFiltering(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for fresh test data
	require.NoError(t, dbMgr.Clean(ctx))

	// Setup test data with multiple churches and teams
	setupTeamChurchFilterTestData(t, ctx, dbMgr)

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	// Generate admin token for queries
	adminToken, err := testutil.GenerateAdminToken(userChurch1AID)
	require.NoError(t, err)

	const teamLeaderboardQuery = `
		query GetTeamLeaderboard($projectId: ID!, $filter: LeaderboardFilter) {
			project(id: $projectId) {
				leaderboard(entityType: TEAMS, filter: $filter, first: 100) {
					edges {
						node {
							id
							name
							score
							rank
						}
					}
					totalCount
					me {
						id
						name
						score
						rank
					}
				}
			}
		}
	`

	t.Run("filter by church 1 returns only teams with church 1 members", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurchID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Should include teamChurch1Only (all members from church 1) and teamMixed (has member from church 1)
		teamIDs := result.getTeamIDs()
		assert.Equal(t, 2, result.Project.Leaderboard.TotalCount)
		assert.Contains(t, teamIDs, teamChurch1OnlyID)
		assert.Contains(t, teamIDs, teamMixedID)
		assert.NotContains(t, teamIDs, teamChurch2OnlyID)
	})

	t.Run("filter by church 2 returns only teams with church 2 members", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurch2ID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Should include teamChurch2Only (all members from church 2) and teamMixed (has member from church 2)
		teamIDs := result.getTeamIDs()
		assert.Equal(t, 2, result.Project.Leaderboard.TotalCount)
		assert.Contains(t, teamIDs, teamChurch2OnlyID)
		assert.Contains(t, teamIDs, teamMixedID)
		assert.NotContains(t, teamIDs, teamChurch1OnlyID)
	})

	t.Run("team with members from both churches appears in both filters", func(t *testing.T) {
		// Filter by church 1
		resp1 := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurchID,
			},
		})
		require.False(t, resp1.HasErrors(), "unexpected error: %s", resp1.ErrorMessage())

		var result1 teamLeaderboardResult
		require.NoError(t, resp1.UnmarshalData(&result1))
		teamIDs1 := result1.getTeamIDs()
		assert.Contains(t, teamIDs1, teamMixedID, "mixed team should appear in church 1 filter")

		// Filter by church 2
		resp2 := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurch2ID,
			},
		})
		require.False(t, resp2.HasErrors(), "unexpected error: %s", resp2.ErrorMessage())

		var result2 teamLeaderboardResult
		require.NoError(t, resp2.UnmarshalData(&result2))
		teamIDs2 := result2.getTeamIDs()
		assert.Contains(t, teamIDs2, teamMixedID, "mixed team should appear in church 2 filter")
	})

	t.Run("filter with non-existent church returns empty", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": "CH01NONEXISTENT000000000000",
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		assert.Equal(t, 0, result.Project.Leaderboard.TotalCount)
		assert.Empty(t, result.Project.Leaderboard.Edges)
	})

	t.Run("me field populated when user's team has member from filtered church", func(t *testing.T) {
		// Query as user from church 1 who is on teamChurch1Only
		userToken, err := testutil.GenerateUserToken(userChurch1AID)
		require.NoError(t, err)

		resp := client.WithAuth(userToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurchID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		require.NotNil(t, result.Project.Leaderboard.Me, "me should be populated")
		assert.Equal(t, teamChurch1OnlyID, result.Project.Leaderboard.Me.ID)
	})

	t.Run("me field nil when user's team excluded by church filter", func(t *testing.T) {
		// Query as user from church 1 who is on teamChurch1Only, but filter by church 2
		userToken, err := testutil.GenerateUserToken(userChurch1AID)
		require.NoError(t, err)

		resp := client.WithAuth(userToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurch2ID,
			},
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// me should be nil because user's team (teamChurch1Only) has no members from church 2
		assert.Nil(t, result.Project.Leaderboard.Me, "me should be nil when user's team is excluded by filter")
		// But leaderboard should still have results
		assert.Greater(t, result.Project.Leaderboard.TotalCount, 0)
	})

	t.Run("no filter returns all teams", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})

		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Should include all 3 teams
		teamIDs := result.getTeamIDs()
		assert.Equal(t, 3, result.Project.Leaderboard.TotalCount)
		assert.Contains(t, teamIDs, teamChurch1OnlyID)
		assert.Contains(t, teamIDs, teamChurch2OnlyID)
		assert.Contains(t, teamIDs, teamMixedID)
	})
}

// TestTeamLeaderboardCaching tests cache behavior for team leaderboards
func TestTeamLeaderboardCaching(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for fresh test data
	require.NoError(t, dbMgr.Clean(ctx))

	// Setup test data with multiple churches and teams
	setupTeamChurchFilterTestData(t, ctx, dbMgr)

	// Setup test server with cache access
	router, testCache, cleanup, err := testutil.SetupTestServerWithCache(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	// Generate admin token for queries
	adminToken, err := testutil.GenerateAdminToken(userChurch1AID)
	require.NoError(t, err)

	const teamLeaderboardQuery = `
		query GetTeamLeaderboard($projectId: ID!, $filter: LeaderboardFilter) {
			project(id: $projectId) {
				leaderboard(entityType: TEAMS, filter: $filter, first: 100) {
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

	t.Run("repeated query hits cache", func(t *testing.T) {
		// Clear cache before test
		testCache.Clear()

		initialHits := testCache.Hits()

		// First query - should miss cache
		resp1 := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurchID,
			},
		})
		require.False(t, resp1.HasErrors(), "unexpected error: %s", resp1.ErrorMessage())

		// Second query - should hit cache
		resp2 := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurchID,
			},
		})
		require.False(t, resp2.HasErrors(), "unexpected error: %s", resp2.ErrorMessage())

		hitsAfter := testCache.Hits()
		assert.Greater(t, hitsAfter, initialHits, "cache hits should increase after repeated query")
	})

	t.Run("different church filters create separate cache entries", func(t *testing.T) {
		// Clear cache before test
		testCache.Clear()

		// Query with church 1 filter
		resp1 := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurchID,
			},
		})
		require.False(t, resp1.HasErrors(), "unexpected error: %s", resp1.ErrorMessage())

		var result1 teamLeaderboardResult
		require.NoError(t, resp1.UnmarshalData(&result1))

		// Query with church 2 filter
		resp2 := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurch2ID,
			},
		})
		require.False(t, resp2.HasErrors(), "unexpected error: %s", resp2.ErrorMessage())

		var result2 teamLeaderboardResult
		require.NoError(t, resp2.UnmarshalData(&result2))

		// Results should be different (different teams)
		teamIDs1 := result1.getTeamIDs()
		teamIDs2 := result2.getTeamIDs()

		// Church 1 filter should include teamChurch1Only but not teamChurch2Only
		assert.Contains(t, teamIDs1, teamChurch1OnlyID)
		assert.NotContains(t, teamIDs1, teamChurch2OnlyID)

		// Church 2 filter should include teamChurch2Only but not teamChurch1Only
		assert.Contains(t, teamIDs2, teamChurch2OnlyID)
		assert.NotContains(t, teamIDs2, teamChurch1OnlyID)
	})

	t.Run("cache serves correct data per filter", func(t *testing.T) {
		// Clear cache before test
		testCache.Clear()

		// Pre-populate cache with both queries
		client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurchID,
			},
		})
		client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurch2ID,
			},
		})

		// Query church 1 again (should hit cache)
		resp := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
			"filter": map[string]any{
				"churchId": testChurchID,
			},
		})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Verify we got the correct cached data for church 1
		teamIDs := result.getTeamIDs()
		assert.Contains(t, teamIDs, teamChurch1OnlyID, "cache should serve church 1 data")
		assert.NotContains(t, teamIDs, teamChurch2OnlyID, "cache should not serve church 2 data")
	})

	t.Run("cache cleared returns fresh data", func(t *testing.T) {
		// First query to populate cache
		resp1 := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})
		require.False(t, resp1.HasErrors(), "unexpected error: %s", resp1.ErrorMessage())

		// Clear cache
		testCache.Clear()

		hitsAfterClear := testCache.Hits()

		// Query again - should not hit cache (since we cleared it)
		resp2 := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})
		require.False(t, resp2.HasErrors(), "unexpected error: %s", resp2.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp2.UnmarshalData(&result))

		// Should still get correct data
		assert.Equal(t, 3, result.Project.Leaderboard.TotalCount)

		// The next query should hit cache
		resp3 := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})
		require.False(t, resp3.HasErrors(), "unexpected error: %s", resp3.ErrorMessage())

		hitsAfterThird := testCache.Hits()
		assert.Greater(t, hitsAfterThird, hitsAfterClear, "third query should hit cache")
	})
}

// setupTeamChurchFilterTestData creates test data for team church filtering tests
func setupTeamChurchFilterTestData(t *testing.T, ctx context.Context, dbMgr *testutil.TestDBManager) {
	t.Helper()

	now := time.Now()
	birthdate := now.AddDate(-25, 0, 0)

	// Create 2 churches
	require.NoError(t, dbMgr.CreateTestChurch(ctx, testChurchID, "Test Church 1", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestChurch(ctx, testChurch2ID, "Test Church 2", "NO", "S"))

	// Create test project
	require.NoError(t, dbMgr.CreateTestProject(ctx, testProjectID, "Test Project"))

	// Update settings to point to our test project
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'`, testProjectID)
	require.NoError(t, err)

	// Create 3 teams:
	// - teamChurch1Only: All members from church 1
	// - teamChurch2Only: All members from church 2
	// - teamMixed: Members from both churches
	require.NoError(t, dbMgr.CreateTestTeam(ctx, teamChurch1OnlyID, "Team Church1 Only", testProjectID))
	require.NoError(t, dbMgr.CreateTestTeam(ctx, teamChurch2OnlyID, "Team Church2 Only", testProjectID))
	require.NoError(t, dbMgr.CreateTestTeam(ctx, teamMixedID, "Team Mixed", testProjectID))

	// Create users and assign to teams
	// Team Church1 Only: 2 users from church 1
	require.NoError(t, dbMgr.CreateTestUser(ctx, userChurch1AID, "Church1A", "MALE", birthdate, testChurchID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userChurch1AID, testProjectID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userChurch1AID, teamChurch1OnlyID))
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userChurch1AID, testProjectID, 100))

	require.NoError(t, dbMgr.CreateTestUser(ctx, userChurch1BID, "Church1B", "FEMALE", birthdate, testChurchID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userChurch1BID, testProjectID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userChurch1BID, teamChurch1OnlyID))
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userChurch1BID, testProjectID, 50))

	// Team Church2 Only: 2 users from church 2
	require.NoError(t, dbMgr.CreateTestUser(ctx, userChurch2AID, "Church2A", "MALE", birthdate, testChurch2ID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userChurch2AID, testProjectID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userChurch2AID, teamChurch2OnlyID))
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userChurch2AID, testProjectID, 75))

	require.NoError(t, dbMgr.CreateTestUser(ctx, userChurch2BID, "Church2B", "FEMALE", birthdate, testChurch2ID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userChurch2BID, testProjectID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userChurch2BID, teamChurch2OnlyID))
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userChurch2BID, testProjectID, 25))

	// Team Mixed: 1 user from church 1, 1 user from church 2
	require.NoError(t, dbMgr.CreateTestUser(ctx, userMixedChurch1ID, "MixedChurch1", "MALE", birthdate, testChurchID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userMixedChurch1ID, testProjectID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userMixedChurch1ID, teamMixedID))
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userMixedChurch1ID, testProjectID, 60))

	require.NoError(t, dbMgr.CreateTestUser(ctx, userMixedChurch2ID, "MixedChurch2", "FEMALE", birthdate, testChurch2ID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userMixedChurch2ID, testProjectID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userMixedChurch2ID, teamMixedID))
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userMixedChurch2ID, testProjectID, 40))
}

// ==================== Multi-Team User Scoring Tests ====================

// Test entity IDs for multi-team tests (must be exactly 28 chars: 2-char prefix + 26-char ULID)
const (
	multiTeamUserID  = "US01TEST00000000000MULTITEAM"
	multiTeamTeam1ID = "TM01TEST0000000000MULTITEAM1"
	multiTeamTeam2ID = "TM01TEST0000000000MULTITEAM2"
)

// TestLeaderboardMultiTeamUserScoring tests that when a user is a member of multiple teams
// in the same project, adding a score for that user updates ALL team leaderboards correctly.
// This tests the fix for the bug where the trigger only updated one random team.
func TestLeaderboardMultiTeamUserScoring(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for fresh test data
	require.NoError(t, dbMgr.Clean(ctx))

	now := time.Now()
	birthdate := now.AddDate(-25, 0, 0)

	// Create test church
	require.NoError(t, dbMgr.CreateTestChurch(ctx, testChurchID, "Test Church", "NO", "S"))

	// Create test project
	require.NoError(t, dbMgr.CreateTestProject(ctx, testProjectID, "Test Project"))

	// Update settings to point to our test project
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'`, testProjectID)
	require.NoError(t, err)

	// Create 2 teams in the same project
	require.NoError(t, dbMgr.CreateTestTeam(ctx, multiTeamTeam1ID, "Multi Team 1", testProjectID))
	require.NoError(t, dbMgr.CreateTestTeam(ctx, multiTeamTeam2ID, "Multi Team 2", testProjectID))

	// Create user who will be a member of BOTH teams
	require.NoError(t, dbMgr.CreateTestUser(ctx, multiTeamUserID, "MultiTeamUser", "MALE", birthdate, testChurchID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, multiTeamUserID, testProjectID))

	// Add user to BOTH teams
	require.NoError(t, dbMgr.AddUserToTeam(ctx, multiTeamUserID, multiTeamTeam1ID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, multiTeamUserID, multiTeamTeam2ID))

	// Add score for the user - this should update BOTH team leaderboards via the trigger
	require.NoError(t, dbMgr.AddScoreForUser(ctx, multiTeamUserID, testProjectID, 100))

	// Setup test server with cache access
	router, testCache, cleanup, err := testutil.SetupTestServerWithCache(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(multiTeamUserID)
	require.NoError(t, err)

	const teamLeaderboardQuery = `
		query GetTeamLeaderboard($projectId: ID!) {
			project(id: $projectId) {
				leaderboard(entityType: TEAMS, first: 100) {
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

	t.Run("trigger updates both teams when user is in multiple teams", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Both teams should appear in the leaderboard with the same score
		require.Equal(t, 2, result.Project.Leaderboard.TotalCount, "Both teams should appear in leaderboard")
		require.Len(t, result.Project.Leaderboard.Edges, 2)

		// Build a map of team scores for easier verification
		teamScores := make(map[string]int)
		for _, edge := range result.Project.Leaderboard.Edges {
			teamScores[edge.Node.ID] = edge.Node.Score
		}

		// Both teams should have score 100 from the user's score
		assert.Equal(t, 100, teamScores[multiTeamTeam1ID], "Team 1 should have score of 100")
		assert.Equal(t, 100, teamScores[multiTeamTeam2ID], "Team 2 should have score of 100")
	})

	t.Run("adding more scores updates both teams", func(t *testing.T) {
		// Add another score
		require.NoError(t, dbMgr.AddScoreForUser(ctx, multiTeamUserID, testProjectID, 50))

		// Clear cache to see fresh data
		testCache.Clear()

		resp := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})
		require.False(t, resp.HasErrors())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		// Build a map of team scores
		teamScores := make(map[string]int)
		for _, edge := range result.Project.Leaderboard.Edges {
			teamScores[edge.Node.ID] = edge.Node.Score
		}

		// Both teams should have cumulative score 150
		assert.Equal(t, 150, teamScores[multiTeamTeam1ID], "Team 1 should have cumulative score of 150")
		assert.Equal(t, 150, teamScores[multiTeamTeam2ID], "Team 2 should have cumulative score of 150")
	})

	t.Run("regenerate_leaderboards produces same results as trigger", func(t *testing.T) {
		// Clear cache to ensure we get fresh data
		testCache.Clear()

		// Get current scores from trigger-updated leaderboards
		resp := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})
		require.False(t, resp.HasErrors())

		var beforeRegen teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&beforeRegen))

		// Build map of scores before regeneration
		scoresBefore := make(map[string]int)
		for _, edge := range beforeRegen.Project.Leaderboard.Edges {
			scoresBefore[edge.Node.ID] = edge.Node.Score
		}

		// Call regenerate_leaderboards
		_, err := dbMgr.DB.Pool.Exec(ctx, `SELECT * FROM regenerate_leaderboards()`)
		require.NoError(t, err)

		// Clear cache after regeneration
		testCache.Clear()

		// Query again after regeneration
		respAfter := client.WithAuth(adminToken).MustExecute(t, teamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})
		require.False(t, respAfter.HasErrors())

		var afterRegen teamLeaderboardResult
		require.NoError(t, respAfter.UnmarshalData(&afterRegen))

		// Build map of scores after regeneration
		scoresAfter := make(map[string]int)
		for _, edge := range afterRegen.Project.Leaderboard.Edges {
			scoresAfter[edge.Node.ID] = edge.Node.Score
		}

		// Scores should match before and after regeneration
		assert.Equal(t, scoresBefore[multiTeamTeam1ID], scoresAfter[multiTeamTeam1ID],
			"Team 1 score should match before and after regeneration")
		assert.Equal(t, scoresBefore[multiTeamTeam2ID], scoresAfter[multiTeamTeam2ID],
			"Team 2 score should match before and after regeneration")
	})
}

// TestLeaderboardMultiTeamWithSuperTeam tests that when a user is in multiple teams
// that belong to the same super_team, the super_team score is only updated once per score entry.
func TestLeaderboardMultiTeamWithSuperTeam(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for fresh test data
	require.NoError(t, dbMgr.Clean(ctx))

	now := time.Now()
	birthdate := now.AddDate(-25, 0, 0)

	// Test IDs (must be exactly 28 chars: 2-char prefix + 26-char ULID)
	const (
		superTeamID        = "ST01TEST0000000000000SUPER01"
		teamInSuperTeam1ID = "TM01TEST0000000000SUPERTEAM1"
		teamInSuperTeam2ID = "TM01TEST0000000000SUPERTEAM2"
		userInBothTeamsID  = "US01TEST000000000INBOTHTEAMS"
	)

	// Create test church
	require.NoError(t, dbMgr.CreateTestChurch(ctx, testChurchID, "Test Church", "NO", "S"))

	// Create test project
	require.NoError(t, dbMgr.CreateTestProject(ctx, testProjectID, "Test Project"))

	// Update settings
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'`, testProjectID)
	require.NoError(t, err)

	// Create super team
	_, err = dbMgr.DB.Pool.Exec(ctx, `
		INSERT INTO super_teams (id, project_id, name)
		VALUES ($1, $2, 'Test Super Team')
	`, superTeamID, testProjectID)
	require.NoError(t, err)

	// Create 2 teams that both belong to the same super_team
	_, err = dbMgr.DB.Pool.Exec(ctx, `
		INSERT INTO teams (id, project_id, name, join_code, super_team_id)
		VALUES ($1, $2, 'Team In Super 1', 'TIST1', $3)
	`, teamInSuperTeam1ID, testProjectID, superTeamID)
	require.NoError(t, err)

	_, err = dbMgr.DB.Pool.Exec(ctx, `
		INSERT INTO teams (id, project_id, name, join_code, super_team_id)
		VALUES ($1, $2, 'Team In Super 2', 'TIST2', $3)
	`, teamInSuperTeam2ID, testProjectID, superTeamID)
	require.NoError(t, err)

	// Create user in both teams
	require.NoError(t, dbMgr.CreateTestUser(ctx, userInBothTeamsID, "UserInBothTeams", "MALE", birthdate, testChurchID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userInBothTeamsID, testProjectID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userInBothTeamsID, teamInSuperTeam1ID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, userInBothTeamsID, teamInSuperTeam2ID))

	// Add score for the user
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userInBothTeamsID, testProjectID, 100))

	// Setup test server
	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(userInBothTeamsID)
	require.NoError(t, err)

	const superTeamLeaderboardQuery = `
		query GetSuperTeamLeaderboard($projectId: ID!) {
			project(id: $projectId) {
				leaderboard(entityType: SUPERTEAMS, first: 100) {
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

	t.Run("super_team score is only counted once when user is in multiple teams of same super_team", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, superTeamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		require.Equal(t, 1, result.Project.Leaderboard.TotalCount, "Should have exactly 1 super_team")
		require.Len(t, result.Project.Leaderboard.Edges, 1)
		assert.Equal(t, superTeamID, result.Project.Leaderboard.Edges[0].Node.ID)
		// Score should be 100, NOT 200 (which would happen if super_team was updated twice)
		assert.Equal(t, 100, result.Project.Leaderboard.Edges[0].Node.Score,
			"Super team score should be 100, not 200 (should only count once)")
	})

	t.Run("regenerate produces same super_team score", func(t *testing.T) {
		// Call regenerate_leaderboards
		_, err := dbMgr.DB.Pool.Exec(ctx, `SELECT * FROM regenerate_leaderboards()`)
		require.NoError(t, err)

		resp := client.WithAuth(adminToken).MustExecute(t, superTeamLeaderboardQuery, map[string]any{
			"projectId": testProjectID,
		})
		require.False(t, resp.HasErrors())

		var result teamLeaderboardResult
		require.NoError(t, resp.UnmarshalData(&result))

		require.Len(t, result.Project.Leaderboard.Edges, 1)
		assert.Equal(t, 100, result.Project.Leaderboard.Edges[0].Node.Score,
			"Super team score after regeneration should still be 100")
	})
}
