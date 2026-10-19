package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/ulid"
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
	require.NoError(t, setCurrentProject(ctx, dbMgr, cacheTestProjectID))

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

// TestLeaderboardCacheRefreshedAfterAsyncApply covers the gap between a score
// write and its asynchronous leaderboard apply (migration 00101): a ranking
// read in between is cached with the old score, and the write-time project
// invalidation has already happened. The apply worker must drop it once the
// delta is applied, or the stale ranking lives for the full cache TTL (seen
// after batched bet settlement: ~1 min of queued deltas).
func TestLeaderboardCacheRefreshedAfterAsyncApply(t *testing.T) {
	const (
		churchID  = "CH01TEST0000000000000ASYNC01"
		projectID = "PR01TEST0000000000000ASYNC01"
		userID    = "US01TEST0000000000000ASYNC01"
	)
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))

	require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Async Church", "NO", "S"))
	require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Async Project"))
	require.NoError(t, setCurrentProject(ctx, dbMgr, projectID))
	require.NoError(t, dbMgr.CreateTestUser(ctx, userID, "AsyncUser", "MALE", time.Now().AddDate(-25, 0, 0), churchID))
	require.NoError(t, dbMgr.EnrollUserInProject(ctx, userID, projectID))
	require.NoError(t, dbMgr.AddLeaderboardConsent(ctx, userID))
	require.NoError(t, dbMgr.AddScoreForUser(ctx, userID, projectID, 100))

	router, testCache, cleanup, err := testutil.SetupTestServerWithCache(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()
	client := testutil.NewGraphQLClient(router)
	defer client.Close()
	token, err := testutil.GenerateAdminToken(userID)
	require.NoError(t, err)

	score := func(t *testing.T) int {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, `query($projectId: ID!) {
			project(id: $projectId) { leaderboard(entityType: PERSONS, first: 10) { edges { node { id score } } totalCount } }
		}`, map[string]any{"projectId": projectID})
		require.False(t, resp.HasErrors(), resp.ErrorMessage())
		var res leaderboardResult
		require.NoError(t, resp.UnmarshalData(&res))
		require.Equal(t, 1, res.Project.Leaderboard.TotalCount)
		return res.Project.Leaderboard.Edges[0].Node.Score
	}

	// The ranking is cached with the old score
	require.Equal(t, 100, score(t))

	// A score write: the journal row commits (its delta is queued, not yet
	// applied) and the write invalidates the project, as every score write does
	reason := "async"
	_, err = dbMgr.DB.Queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
		ID: ulid.NewScoreJournalID(), ProjectID: projectID, UserID: userID,
		Points: 50, SourceType: "MANUAL", Reason: &reason,
	})
	require.NoError(t, err)
	testCache.InvalidateProject(projectID)

	// A ranking read lands before the apply. The test router drains the queue
	// before each request; holding the queued row's lock makes that drain skip
	// it (SKIP LOCKED), exactly like a read that beats the production worker.
	lock, err := dbMgr.DB.Pool.Begin(ctx)
	require.NoError(t, err)
	queued, err := dbMgr.DB.Queries.WithTx(lock).LockLeaderboardApplyQueue(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, queued, "the delta is waiting for the apply worker")
	require.Equal(t, 100, score(t), "read before the apply: the old score, now cached")
	require.NoError(t, lock.Rollback(ctx))

	// The production apply worker applies the delta and refreshes the caches
	worker := services.NewLeaderboardApplyWorker(dbMgr.DB.Queries, testCache.InvalidateLeaderboardScores)
	worker.Start(ctx)
	defer worker.Stop()
	require.Eventually(t, func() bool {
		n, err := dbMgr.DB.Queries.CountLeaderboardApplyQueue(ctx)
		return err == nil && n == 0
	}, 5*time.Second, 20*time.Millisecond)

	assert.Eventually(t, func() bool { return score(t) == 150 }, 3*time.Second, 50*time.Millisecond,
		"the ranking must show the applied score, not the one cached before the apply")
}

// setCurrentProject points the current_project_id setting at projectID
func setCurrentProject(ctx context.Context, dbMgr *testutil.TestDBManager, projectID string) error {
	n, err := dbMgr.DB.Queries.UpdateSettingText(ctx, sqlc.UpdateSettingTextParams{Key: "current_project_id", ValueText: projectID})
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("current_project_id setting not found")
	}
	return nil
}
