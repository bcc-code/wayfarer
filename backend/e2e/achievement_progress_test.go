package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Exercise the profile's project loader as well as direct achievement reads,
// with warm caches and the same item edits that used to corrupt award counts.
func TestAchievementProgress(t *testing.T) {
	ctx := context.Background()
	db, _ := GetTestEnv()
	for _, kind := range []string{"Content", "Streak"} {
		t.Run(kind, func(t *testing.T) {
			require.NoError(t, db.Clean(ctx))
			churchID, projectID := ulid.NewChurchID(), ulid.NewProjectID()
			userID, otherID := ulid.NewUserID(), ulid.NewUserID()
			achievementID := ulid.NewAchievementID()
			require.NoError(t, db.CreateTestChurch(ctx, churchID, "Progress Church", "NO", "L"))
			require.NoError(t, db.CreateTestProject(ctx, projectID, "Progress Project"))
			require.NoError(t, db.SetCurrentProject(ctx, projectID))
			for _, id := range []string{userID, otherID} {
				require.NoError(t, db.CreateUserWithPersonUUID(ctx, id, "Progress User", churchID, uuid.New()))
				require.NoError(t, db.EnrollUserInProject(ctx, id, projectID))
			}
			require.NoError(t, db.AssignRole(ctx, userID, testutil.RoleAdmin))
			contentIDs := make([]string, 4)
			for i := range contentIDs {
				contentIDs[i] = ulid.NewExternalContentID()
				require.NoError(t, db.CreateExternalContentWithDeadline(ctx, contentIDs[i], "progress-plan", fmt.Sprintf("progress-task-%d", i), "media_episode", "ssf", time.Now().Add(24*time.Hour)))
			}
			if kind == "Content" {
				require.NoError(t, db.CreateContentAchievement(ctx, achievementID, projectID, "Progress", 0, contentIDs[:3], false))
			} else {
				require.NoError(t, db.CreateStreakAchievement(ctx, achievementID, projectID, "Progress", 0, contentIDs[:3]))
			}
			router, c, cleanup, err := testutil.SetupTestServerWithCache(ctx, db)
			require.NoError(t, err)
			defer cleanup()
			client := testutil.NewGraphQLClient(router)
			defer client.Close()
			token, err := testutil.GenerateAdminToken(userID)
			require.NoError(t, err)
			otherToken, err := testutil.GenerateUserToken(otherID)
			require.NoError(t, err)
			m2mToken, err := testutil.GenerateM2MToken()
			require.NoError(t, err)

			type progress struct {
				ID                 string
				TotalItems         int
				CompletedItemCount int
				AchievedAt         *string
				UserCompletedItems []struct{ ID string }
			}
			fields := fmt.Sprintf(`id achievedAt ... on %sAchievement { totalItems completedItemCount userCompletedItems { id } }`, kind)
			check := func(auth string, done, total int, awarded bool) {
				t.Helper()
				resp := client.WithAuth(auth).MustExecute(t, `query($id: ID!) {
					myCurrentProject { achievements { `+fields+` } }
					achievement(id: $id) { `+fields+` }
				}`, map[string]any{"id": achievementID})
				require.False(t, resp.HasErrors(), resp.ErrorMessage())
				var result struct {
					MyCurrentProject struct{ Achievements []progress }
					Achievement      progress
				}
				require.NoError(t, resp.UnmarshalData(&result))
				require.Len(t, result.MyCurrentProject.Achievements, 1)
				for _, value := range []progress{result.Achievement, result.MyCurrentProject.Achievements[0]} {
					assert.Equal(t, achievementID, value.ID)
					assert.Equal(t, done, value.CompletedItemCount)
					assert.Equal(t, total, value.TotalItems)
					assert.Len(t, value.UserCompletedItems, done)
					assert.Equal(t, awarded, value.AchievedAt != nil)
				}
				c.Wait()
			}
			mark := func(action string, index int) {
				t.Helper()
				field := fmt.Sprintf("%s%sItemCompleted", action, kind)
				resp := client.WithAuth(m2mToken).MustExecute(t, `mutation($user: ID!, $content: ID!) {
					`+field+`(userId: $user, externalContentId: $content) { id totalItems }
				}`, map[string]any{"user": userID, "content": contentIDs[index]})
				require.False(t, resp.HasErrors(), resp.ErrorMessage())
			}
			edit := func(indices ...int) {
				t.Helper()
				items := make([]map[string]any, 0, len(indices))
				for _, i := range indices {
					items = append(items, map[string]any{"externalContentId": contentIDs[i]})
				}
				resp := client.WithAuth(token).MustExecute(t, fmt.Sprintf(`mutation($id: ID!, $input: Update%sAchievementInput!) {
					update%sAchievement(id: $id, input: $input) { id totalItems }
				}`, kind, kind), map[string]any{"id": achievementID, "input": map[string]any{"items": items}})
				require.False(t, resp.HasErrors(), resp.ErrorMessage())
				var result map[string]struct{ TotalItems int }
				require.NoError(t, resp.UnmarshalData(&result))
				assert.Equal(t, len(indices), result["update"+kind+"Achievement"].TotalItems)
			}

			check(token, 0, 3, false)
			check(otherToken, 0, 3, false)
			mark("mark", 0)
			mark("mark", 1)
			mark("mark", 1) // Duplicate events must not increase progress.
			check(token, 2, 3, false)
			check(otherToken, 0, 3, false)
			edit(1, 2, 3) // Replace completed A with D, preserving the total.
			check(token, 1, 3, false)
			mark("mark", 2) // Three raw rows, but only B and C still count.
			check(token, 2, 3, false)
			mark("unmark", 2)
			check(token, 1, 3, false)
			mark("mark", 2)
			edit(1, 2)                            // Removed A still has historical progress; D is no longer required.
			check(token, 2, 2, kind == "Content") // Content edits award without another completion event.
			check(otherToken, 0, 2, false)

			// The event-processing path must use the same intersection as the API.
			service := &services.ContentAchievementService{DB: db.DB, Cache: c, Loaders: loaders.NewLoaders(db.DB, c)}
			service.ProcessContentEvent(ctx, userID, "progress-task-2")
			check(token, 2, 2, true)
			edit(0, 1, 2, 3)
			check(token, 3, 4, true) // Requirements can change without revoking an award.
			edit()
			check(token, 0, 0, true)
		})
	}
}
