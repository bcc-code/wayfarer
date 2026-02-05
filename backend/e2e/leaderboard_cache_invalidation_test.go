package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test entity IDs for cache invalidation test (exactly 28 chars: 2-char prefix + 26-char ULID)
const (
	cacheTestChurchID  = "CH01TEST0000000000000CACHE01"
	cacheTestProjectID = "PR01TEST0000000000000CACHE01"
	cacheTestTeamID    = "TM01TEST0000000000000CACHE01"
	cacheTestUserID    = "US01TEST0000000000000CACHE01"
)

// TestLeaderboardCacheInvalidationAfterScoreChange verifies that after scores
// are added and InvalidateProject is called, the leaderboard cache is properly
// invalidated so subsequent queries return fresh data.
//
// This is the regression test for the bug where LeaderboardService used *cache.Cache
// (base type) instead of *cache.CacheWithRegistry, causing SetWithTTL to bypass
// the KeyRegistry. InvalidateProject's DeletePrefix then found no keys to delete,
// leaving stale leaderboard entries until TTL expiry.
func TestLeaderboardCacheInvalidationAfterScoreChange(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	// Clean database for fresh test data
	require.NoError(t, dbMgr.Clean(ctx))

	now := time.Now()
	birthdate := now.AddDate(-25, 0, 0)

	// Create test church
	require.NoError(t, dbMgr.CreateTestChurch(ctx, cacheTestChurchID, "Cache Test Church", "NO", "S"))

	// Create test project
	require.NoError(t, dbMgr.CreateTestProject(ctx, cacheTestProjectID, "Cache Test Project"))

	// Update settings to point to our test project
	_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'`, cacheTestProjectID)
	require.NoError(t, err)

	// Create test team and user
	require.NoError(t, dbMgr.CreateTestTeam(ctx, cacheTestTeamID, "Cache Test Team", cacheTestProjectID))
	require.NoError(t, dbMgr.CreateTestUser(ctx, cacheTestUserID, "CacheTestUser", "MALE", birthdate, cacheTestChurchID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, cacheTestUserID, cacheTestProjectID))
	require.NoError(t, dbMgr.AddUserToTeam(ctx, cacheTestUserID, cacheTestTeamID))
	require.NoError(t, dbMgr.AddLeaderboardConsent(ctx, cacheTestUserID))

	// Award initial 100 points
	require.NoError(t, dbMgr.AddScoreForUser(ctx, cacheTestUserID, cacheTestProjectID, 100))

	// Setup test server with cache access
	router, testCache, cleanup, err := testutil.SetupTestServerWithCache(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(cacheTestUserID)
	require.NoError(t, err)

	const leaderboardQuery = `
		query GetLeaderboard($projectId: ID!) {
			project(id: $projectId) {
				leaderboard(entityType: PERSONS, first: 100) {
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

	// Step 1: Query leaderboard to populate cache (should see score=100)
	resp := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
		"projectId": cacheTestProjectID,
	})
	require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

	var result1 leaderboardResult
	require.NoError(t, resp.UnmarshalData(&result1))
	require.Equal(t, 1, result1.Project.Leaderboard.TotalCount)
	assert.Equal(t, 100, result1.Project.Leaderboard.Edges[0].Node.Score, "initial score should be 100")

	// Step 2: Award 50 more points (total should become 150)
	require.NoError(t, dbMgr.AddScoreForUser(ctx, cacheTestUserID, cacheTestProjectID, 50))

	// Step 3: Invalidate project cache (simulates what plugin handlers do)
	testCache.InvalidateProject(cacheTestProjectID)

	// Step 4: Query leaderboard again - should see updated score=150
	resp2 := client.WithAuth(adminToken).MustExecute(t, leaderboardQuery, map[string]any{
		"projectId": cacheTestProjectID,
	})
	require.False(t, resp2.HasErrors(), "unexpected error: %s", resp2.ErrorMessage())

	var result2 leaderboardResult
	require.NoError(t, resp2.UnmarshalData(&result2))
	require.Equal(t, 1, result2.Project.Leaderboard.TotalCount)
	assert.Equal(t, 150, result2.Project.Leaderboard.Edges[0].Node.Score,
		"score should be 150 after invalidation, not stale 100")
}
