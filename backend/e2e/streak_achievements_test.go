package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreakAchievements(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	t.Run("m2m_mark_streak_item_completed_records_progress", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))

		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()

		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))
		require.NoError(t, dbMgr.SetCurrentProject(ctx, projectID))
		deadline := time.Now().Add(24 * time.Hour)
		require.NoError(t, dbMgr.CreateExternalContentWithDeadline(ctx, contentID, "test-plan", "task-1", "media_episode", "ssf", deadline))
		require.NoError(t, dbMgr.CreateStreakAchievement(ctx, achievementID, projectID, "Streak Test", 100, []string{contentID}))
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
		require.NoError(t, err)
		defer cleanup()

		client := testutil.NewGraphQLClient(router)
		defer client.Close()

		m2mToken, err := testutil.GenerateM2MToken()
		require.NoError(t, err)

		resp := client.WithAuth(m2mToken).MustExecute(t, `
			mutation MarkStreakItemCompleted($userId: ID!, $externalContentId: ID!) {
				markStreakItemCompleted(userId: $userId, externalContentId: $externalContentId) {
					id
					name
					totalItems
				}
			}
		`, map[string]any{
			"userId":            userID,
			"externalContentId": contentID,
		})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		var result struct {
			MarkStreakItemCompleted []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				TotalItems int    `json:"totalItems"`
			} `json:"markStreakItemCompleted"`
		}
		require.NoError(t, resp.UnmarshalData(&result))
		require.Len(t, result.MarkStreakItemCompleted, 1)
		assert.Equal(t, achievementID, result.MarkStreakItemCompleted[0].ID)

		// Verify progress recorded
		progressCount, err := dbMgr.GetUserStreakProgress(ctx, userID, achievementID)
		require.NoError(t, err)
		assert.Equal(t, 1, progressCount)

		// Verify auto-award (single item, all completed)
		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.Contains(t, achievements, achievementID)
	})

	t.Run("deadline_enforcement_rejects_late_completion", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))

		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()

		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))
		require.NoError(t, dbMgr.SetCurrentProject(ctx, projectID))
		deadline := time.Now().Add(-24 * time.Hour) // past
		require.NoError(t, dbMgr.CreateExternalContentWithDeadline(ctx, contentID, "test-plan", "task-late", "media_episode", "ssf", deadline))
		require.NoError(t, dbMgr.CreateStreakAchievement(ctx, achievementID, projectID, "Late Streak", 100, []string{contentID}))
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
		require.NoError(t, err)
		defer cleanup()

		client := testutil.NewGraphQLClient(router)
		defer client.Close()

		m2mToken, err := testutil.GenerateM2MToken()
		require.NoError(t, err)

		resp := client.WithAuth(m2mToken).MustExecute(t, `
			mutation MarkStreakItemCompleted($userId: ID!, $externalContentId: ID!) {
				markStreakItemCompleted(userId: $userId, externalContentId: $externalContentId) { id }
			}
		`, map[string]any{
			"userId":            userID,
			"externalContentId": contentID,
		})
		assert.True(t, resp.HasErrors(), "should reject completion past deadline")

		progressCount, err := dbMgr.GetUserStreakProgress(ctx, userID, achievementID)
		require.NoError(t, err)
		assert.Equal(t, 0, progressCount)
	})

	t.Run("force_flag_bypasses_deadline", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))

		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()

		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))
		require.NoError(t, dbMgr.SetCurrentProject(ctx, projectID))
		deadline := time.Now().Add(-24 * time.Hour) // past
		require.NoError(t, dbMgr.CreateExternalContentWithDeadline(ctx, contentID, "test-plan", "task-force", "media_episode", "ssf", deadline))
		require.NoError(t, dbMgr.CreateStreakAchievement(ctx, achievementID, projectID, "Force Streak", 100, []string{contentID}))
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
		require.NoError(t, err)
		defer cleanup()

		client := testutil.NewGraphQLClient(router)
		defer client.Close()

		m2mToken, err := testutil.GenerateM2MToken()
		require.NoError(t, err)

		resp := client.WithAuth(m2mToken).MustExecute(t, `
			mutation MarkStreakItemCompleted($userId: ID!, $externalContentId: ID!, $force: Boolean) {
				markStreakItemCompleted(userId: $userId, externalContentId: $externalContentId, force: $force) { id }
			}
		`, map[string]any{
			"userId":            userID,
			"externalContentId": contentID,
			"force":             true,
		})
		require.False(t, resp.HasErrors(), "force should bypass deadline: %s", resp.ErrorMessage())

		progressCount, err := dbMgr.GetUserStreakProgress(ctx, userID, achievementID)
		require.NoError(t, err)
		assert.Equal(t, 1, progressCount)

		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.Contains(t, achievements, achievementID)
	})

	t.Run("multi_item_awards_only_when_all_completed", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))

		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID1 := ulid.NewExternalContentID()
		contentID2 := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()

		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))
		require.NoError(t, dbMgr.SetCurrentProject(ctx, projectID))
		deadline := time.Now().Add(24 * time.Hour)
		require.NoError(t, dbMgr.CreateExternalContentWithDeadline(ctx, contentID1, "test-plan", "task-a", "media_episode", "ssf", deadline))
		require.NoError(t, dbMgr.CreateExternalContentWithDeadline(ctx, contentID2, "test-plan", "task-b", "media_episode", "ssf", deadline))
		require.NoError(t, dbMgr.CreateStreakAchievement(ctx, achievementID, projectID, "Multi Streak", 200, []string{contentID1, contentID2}))
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
		require.NoError(t, err)
		defer cleanup()

		client := testutil.NewGraphQLClient(router)
		defer client.Close()

		m2mToken, err := testutil.GenerateM2MToken()
		require.NoError(t, err)

		markMutation := `
			mutation MarkStreakItemCompleted($userId: ID!, $externalContentId: ID!) {
				markStreakItemCompleted(userId: $userId, externalContentId: $externalContentId) { id }
			}
		`

		// Complete first item
		resp := client.WithAuth(m2mToken).MustExecute(t, markMutation, map[string]any{
			"userId":            userID,
			"externalContentId": contentID1,
		})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		// Not yet awarded
		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.NotContains(t, achievements, achievementID, "should not award with partial completion")

		// Complete second item
		resp = client.WithAuth(m2mToken).MustExecute(t, markMutation, map[string]any{
			"userId":            userID,
			"externalContentId": contentID2,
		})
		require.False(t, resp.HasErrors(), "unexpected error: %s", resp.ErrorMessage())

		// Now awarded
		achievements, err = dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.Contains(t, achievements, achievementID, "should auto-award when all items completed")
	})

	t.Run("webhook_processes_streak_achievements", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))

		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()
		taskID := "webhook-streak-task"

		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))
		require.NoError(t, dbMgr.SetCurrentProject(ctx, projectID))
		deadline := time.Now().Add(24 * time.Hour)
		require.NoError(t, dbMgr.CreateExternalContentWithDeadline(ctx, contentID, "test-plan", taskID, "media_episode", "ssf", deadline))
		require.NoError(t, dbMgr.CreateStreakAchievement(ctx, achievementID, projectID, "Webhook Streak", 100, []string{contentID}))
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Webhook User", churchID, personUUID))

		service, serviceCleanup, err := testutil.NewTestContentAchievementService(dbMgr)
		require.NoError(t, err)
		defer serviceCleanup()

		pgUUID := pgtype.UUID{Bytes: personUUID, Valid: true}
		consumedAt := time.Now()
		err = service.StoreAndProcessContentEvent(ctx, userID, pgUUID, taskID, "test-plan", nil, consumedAt, "test-source", false)
		require.NoError(t, err)

		progressCount, err := dbMgr.GetUserStreakProgress(ctx, userID, achievementID)
		require.NoError(t, err)
		assert.Equal(t, 1, progressCount, "webhook should create streak progress")

		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.Contains(t, achievements, achievementID, "webhook should auto-award streak achievement")
	})

	t.Run("webhook_deadline_uses_consumed_at_not_now", func(t *testing.T) {
		require.NoError(t, dbMgr.Clean(ctx))

		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()
		taskID := "webhook-deadline-task"

		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))
		// Deadline was 2 days ago
		deadline := time.Now().Add(-48 * time.Hour)
		require.NoError(t, dbMgr.CreateExternalContentWithDeadline(ctx, contentID, "test-plan", taskID, "media_episode", "ssf", deadline))
		require.NoError(t, dbMgr.CreateStreakAchievement(ctx, achievementID, projectID, "Deadline Streak", 100, []string{contentID}))
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Deadline User", churchID, personUUID))

		service, serviceCleanup, err := testutil.NewTestContentAchievementService(dbMgr)
		require.NoError(t, err)
		defer serviceCleanup()

		pgUUID := pgtype.UUID{Bytes: personUUID, Valid: true}

		// consumedAt is 3 days ago (BEFORE deadline) - should succeed even though "now" is after deadline
		consumedAt := time.Now().Add(-72 * time.Hour)
		err = service.StoreAndProcessContentEvent(ctx, userID, pgUUID, taskID, "test-plan", nil, consumedAt, "test-source", false)
		require.NoError(t, err)

		progressCount, err := dbMgr.GetUserStreakProgress(ctx, userID, achievementID)
		require.NoError(t, err)
		assert.Equal(t, 1, progressCount, "consumed before deadline should record progress even if now is after deadline")

		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.Contains(t, achievements, achievementID)
	})
}
