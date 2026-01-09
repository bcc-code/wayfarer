package e2e

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentEventsProcessedOnRegistration(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	t.Run("content_event_before_registration_awards_achievement", func(t *testing.T) {
		// Clean database for isolated test
		require.NoError(t, dbMgr.Clean(ctx))

		// Setup: Create required entities
		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		eventID := ulid.NewContentEventID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()
		taskID := "test-task-123"
		planID := "test-plan-456"

		// 1. Create church
		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))

		// 2. Create project
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))

		// 3. Create external content
		require.NoError(t, dbMgr.CreateExternalContent(ctx, contentID, planID, taskID, "media_episode", "test-source"))

		// 4. Create content achievement (not hidden, so it's "published")
		require.NoError(t, dbMgr.CreateContentAchievement(ctx, achievementID, projectID, "Test Content Achievement", 100, []string{contentID}, false))

		// 5. Create content event BEFORE user exists
		progress := float32(1.0)
		require.NoError(t, dbMgr.CreateExternalContentEvent(ctx, eventID, personUUID, taskID, &planID, "test-source", &progress))

		// 6. Create user with the same person_uuid (simulating registration)
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		// 7. Create ContentAchievementService and process pending events
		service, cleanup, err := testutil.NewTestContentAchievementService(dbMgr)
		require.NoError(t, err)
		defer cleanup()

		// Convert uuid.UUID to pgtype.UUID for the service call
		pgUUID := pgtype.UUID{
			Bytes: personUUID,
			Valid: true,
		}
		service.ProcessPendingContentEvents(ctx, userID, pgUUID)

		// 8. Verify: User should have the achievement
		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.Contains(t, achievements, achievementID, "user should have been awarded the achievement")

		// 9. Verify: Score journal entry should exist
		scoreCount, err := dbMgr.GetScoreJournalEntriesCount(ctx, userID, projectID)
		require.NoError(t, err)
		assert.Equal(t, 1, scoreCount, "user should have exactly one score journal entry")

		// 10. Verify: Content progress should be tracked
		progressCount, err := dbMgr.GetUserContentProgress(ctx, userID, achievementID)
		require.NoError(t, err)
		assert.Equal(t, 1, progressCount, "user should have content progress for the achievement")
	})

	t.Run("multiple_content_events_complete_achievement", func(t *testing.T) {
		// Clean database for isolated test
		require.NoError(t, dbMgr.Clean(ctx))

		// Setup
		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID1 := ulid.NewExternalContentID()
		contentID2 := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		eventID1 := ulid.NewContentEventID()
		eventID2 := ulid.NewContentEventID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()
		taskID1 := "test-task-1"
		taskID2 := "test-task-2"
		planID := "test-plan-multi"

		// Create church and project
		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))

		// Create two external content items
		require.NoError(t, dbMgr.CreateExternalContent(ctx, contentID1, planID, taskID1, "media_episode", "test-source"))
		require.NoError(t, dbMgr.CreateExternalContent(ctx, contentID2, planID, taskID2, "media_episode", "test-source"))

		// Create achievement requiring BOTH content items
		require.NoError(t, dbMgr.CreateContentAchievement(ctx, achievementID, projectID, "Multi-Content Achievement", 150, []string{contentID1, contentID2}, false))

		// Create BOTH content events before user exists
		progress := float32(1.0)
		require.NoError(t, dbMgr.CreateExternalContentEvent(ctx, eventID1, personUUID, taskID1, &planID, "test-source", &progress))
		require.NoError(t, dbMgr.CreateExternalContentEvent(ctx, eventID2, personUUID, taskID2, &planID, "test-source", &progress))

		// Create user
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		// Process pending events
		service, cleanup, err := testutil.NewTestContentAchievementService(dbMgr)
		require.NoError(t, err)
		defer cleanup()

		pgUUID := pgtype.UUID{Bytes: personUUID, Valid: true}
		service.ProcessPendingContentEvents(ctx, userID, pgUUID)

		// Verify: Achievement should be awarded (both items completed)
		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.Contains(t, achievements, achievementID, "user should have been awarded the achievement")

		// Verify: Both content items should be tracked
		progressCount, err := dbMgr.GetUserContentProgress(ctx, userID, achievementID)
		require.NoError(t, err)
		assert.Equal(t, 2, progressCount, "user should have progress for both content items")
	})

	t.Run("partial_completion_does_not_award_achievement", func(t *testing.T) {
		// Clean database for isolated test
		require.NoError(t, dbMgr.Clean(ctx))

		// Setup
		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID1 := ulid.NewExternalContentID()
		contentID2 := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		eventID1 := ulid.NewContentEventID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()
		taskID1 := "test-task-partial-1"
		taskID2 := "test-task-partial-2"
		planID := "test-plan-partial"

		// Create church and project
		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))

		// Create two external content items
		require.NoError(t, dbMgr.CreateExternalContent(ctx, contentID1, planID, taskID1, "media_episode", "test-source"))
		require.NoError(t, dbMgr.CreateExternalContent(ctx, contentID2, planID, taskID2, "media_episode", "test-source"))

		// Create achievement requiring BOTH content items
		require.NoError(t, dbMgr.CreateContentAchievement(ctx, achievementID, projectID, "Partial Achievement", 100, []string{contentID1, contentID2}, false))

		// Create only ONE content event (partial completion)
		progress := float32(1.0)
		require.NoError(t, dbMgr.CreateExternalContentEvent(ctx, eventID1, personUUID, taskID1, &planID, "test-source", &progress))

		// Create user
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		// Process pending events
		service, cleanup, err := testutil.NewTestContentAchievementService(dbMgr)
		require.NoError(t, err)
		defer cleanup()

		pgUUID := pgtype.UUID{Bytes: personUUID, Valid: true}
		service.ProcessPendingContentEvents(ctx, userID, pgUUID)

		// Verify: Achievement should NOT be awarded (only partial completion)
		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.NotContains(t, achievements, achievementID, "user should NOT have the achievement with partial completion")

		// Verify: Only one content item should be tracked
		progressCount, err := dbMgr.GetUserContentProgress(ctx, userID, achievementID)
		require.NoError(t, err)
		assert.Equal(t, 1, progressCount, "user should have progress for only one content item")

		// Verify: No score journal entry (no achievement awarded)
		scoreCount, err := dbMgr.GetScoreJournalEntriesCount(ctx, userID, projectID)
		require.NoError(t, err)
		assert.Equal(t, 0, scoreCount, "user should have no score journal entries")
	})

	t.Run("single_content_in_multiple_achievements", func(t *testing.T) {
		// Clean database for isolated test
		require.NoError(t, dbMgr.Clean(ctx))

		// Setup
		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID := ulid.NewExternalContentID()
		achievementID1 := ulid.NewAchievementID()
		achievementID2 := ulid.NewAchievementID()
		eventID := ulid.NewContentEventID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()
		taskID := "test-task-multi-ach"
		planID := "test-plan-multi-ach"

		// Create church and project
		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))

		// Create ONE external content item
		require.NoError(t, dbMgr.CreateExternalContent(ctx, contentID, planID, taskID, "media_episode", "test-source"))

		// Create TWO achievements both using the same content
		require.NoError(t, dbMgr.CreateContentAchievement(ctx, achievementID1, projectID, "Achievement 1", 50, []string{contentID}, false))
		require.NoError(t, dbMgr.CreateContentAchievement(ctx, achievementID2, projectID, "Achievement 2", 75, []string{contentID}, false))

		// Create ONE content event
		progress := float32(1.0)
		require.NoError(t, dbMgr.CreateExternalContentEvent(ctx, eventID, personUUID, taskID, &planID, "test-source", &progress))

		// Create user
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		// Process pending events
		service, cleanup, err := testutil.NewTestContentAchievementService(dbMgr)
		require.NoError(t, err)
		defer cleanup()

		pgUUID := pgtype.UUID{Bytes: personUUID, Valid: true}
		service.ProcessPendingContentEvents(ctx, userID, pgUUID)

		// Verify: BOTH achievements should be awarded
		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.Contains(t, achievements, achievementID1, "user should have achievement 1")
		assert.Contains(t, achievements, achievementID2, "user should have achievement 2")

		// Verify: Two score journal entries (one per achievement)
		scoreCount, err := dbMgr.GetScoreJournalEntriesCount(ctx, userID, projectID)
		require.NoError(t, err)
		assert.Equal(t, 2, scoreCount, "user should have two score journal entries")
	})

	t.Run("idempotent_processing_no_double_award", func(t *testing.T) {
		// Clean database for isolated test
		require.NoError(t, dbMgr.Clean(ctx))

		// Setup
		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		eventID := ulid.NewContentEventID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()
		taskID := "test-task-idempotent"
		planID := "test-plan-idempotent"

		// Create church and project
		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))

		// Create external content
		require.NoError(t, dbMgr.CreateExternalContent(ctx, contentID, planID, taskID, "media_episode", "test-source"))

		// Create content achievement
		require.NoError(t, dbMgr.CreateContentAchievement(ctx, achievementID, projectID, "Idempotent Achievement", 100, []string{contentID}, false))

		// Create content event
		progress := float32(1.0)
		require.NoError(t, dbMgr.CreateExternalContentEvent(ctx, eventID, personUUID, taskID, &planID, "test-source", &progress))

		// Create user
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		// Process pending events TWICE
		service, cleanup, err := testutil.NewTestContentAchievementService(dbMgr)
		require.NoError(t, err)
		defer cleanup()

		pgUUID := pgtype.UUID{Bytes: personUUID, Valid: true}
		service.ProcessPendingContentEvents(ctx, userID, pgUUID)
		service.ProcessPendingContentEvents(ctx, userID, pgUUID) // Second call

		// Verify: Only ONE achievement award
		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		count := 0
		for _, a := range achievements {
			if a == achievementID {
				count++
			}
		}
		assert.Equal(t, 1, count, "achievement should be awarded exactly once")

		// Verify: Only ONE score journal entry
		scoreCount, err := dbMgr.GetScoreJournalEntriesCount(ctx, userID, projectID)
		require.NoError(t, err)
		assert.Equal(t, 1, scoreCount, "user should have exactly one score journal entry")
	})

	t.Run("events_without_matching_content_are_ignored", func(t *testing.T) {
		// Clean database for isolated test
		require.NoError(t, dbMgr.Clean(ctx))

		// Setup
		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		eventID := ulid.NewContentEventID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()
		taskID := "non-existent-task"
		planID := "non-existent-plan"

		// Create church and project (but NO external content)
		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))

		// Create content event for non-existent content
		progress := float32(1.0)
		require.NoError(t, dbMgr.CreateExternalContentEvent(ctx, eventID, personUUID, taskID, &planID, "test-source", &progress))

		// Create user
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		// Process pending events (should not error, just skip)
		service, cleanup, err := testutil.NewTestContentAchievementService(dbMgr)
		require.NoError(t, err)
		defer cleanup()

		pgUUID := pgtype.UUID{Bytes: personUUID, Valid: true}
		service.ProcessPendingContentEvents(ctx, userID, pgUUID)

		// Verify: No achievements awarded
		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, achievements, "no achievements should be awarded for non-existent content")

		// Verify: No score journal entries
		scoreCount, err := dbMgr.GetScoreJournalEntriesCount(ctx, userID, projectID)
		require.NoError(t, err)
		assert.Equal(t, 0, scoreCount, "no score journal entries should exist")
	})

	t.Run("hidden_achievement_not_awarded", func(t *testing.T) {
		// Clean database for isolated test
		require.NoError(t, dbMgr.Clean(ctx))

		// Setup
		churchID := ulid.NewChurchID()
		projectID := ulid.NewProjectID()
		contentID := ulid.NewExternalContentID()
		achievementID := ulid.NewAchievementID()
		eventID := ulid.NewContentEventID()
		userID := ulid.NewUserID()
		personUUID := uuid.New()
		taskID := "test-task-hidden"
		planID := "test-plan-hidden"

		// Create church and project
		require.NoError(t, dbMgr.CreateTestChurch(ctx, churchID, "Test Church", "NO", "L"))
		require.NoError(t, dbMgr.CreateTestProject(ctx, projectID, "Test Project"))

		// Create external content
		require.NoError(t, dbMgr.CreateExternalContent(ctx, contentID, planID, taskID, "media_episode", "test-source"))

		// Create HIDDEN content achievement
		require.NoError(t, dbMgr.CreateContentAchievement(ctx, achievementID, projectID, "Hidden Achievement", 100, []string{contentID}, true))

		// Create content event
		progress := float32(1.0)
		require.NoError(t, dbMgr.CreateExternalContentEvent(ctx, eventID, personUUID, taskID, &planID, "test-source", &progress))

		// Create user
		require.NoError(t, dbMgr.CreateUserWithPersonUUID(ctx, userID, "Test User", churchID, personUUID))

		// Process pending events
		service, cleanup, err := testutil.NewTestContentAchievementService(dbMgr)
		require.NoError(t, err)
		defer cleanup()

		pgUUID := pgtype.UUID{Bytes: personUUID, Valid: true}
		service.ProcessPendingContentEvents(ctx, userID, pgUUID)

		// Verify: Hidden achievement should NOT be awarded
		achievements, err := dbMgr.GetUserAchievements(ctx, userID)
		require.NoError(t, err)
		assert.NotContains(t, achievements, achievementID, "hidden achievement should NOT be awarded")

		// Verify: No score journal entries
		scoreCount, err := dbMgr.GetScoreJournalEntriesCount(ctx, userID, projectID)
		require.NoError(t, err)
		assert.Equal(t, 0, scoreCount, "no score journal entries should exist for hidden achievement")
	})
}
