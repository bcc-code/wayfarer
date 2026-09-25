package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAchievementProgress(t *testing.T) {
	for _, kind := range []string{"content", "streak"} {
		for _, tc := range []struct {
			name      string
			total     int
			completed []string
			userID    string
			want      int
		}{
			{name: "empty", userID: "US1"},
			{name: "not started", total: 13, userID: "US1"},
			{name: "three of thirteen", total: 13, completed: []string{"EC0", "EC1", "EC2"}, userID: "US1", want: 3},
			{name: "removed items", total: 2, completed: []string{"EC0", "EC1", "removed"}, userID: "US1", want: 2},
			{name: "replaced item with same total", total: 3, completed: []string{"EC0", "EC1", "removed"}, userID: "US1", want: 2},
			{name: "anonymous", total: 3, completed: []string{"EC0"}},
			{name: "another user", total: 3, completed: []string{"EC0"}, userID: "US2"},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				c, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
				require.NoError(t, err)
				defer c.Close()
				items := make([]*model.ContentItem, tc.total)
				for i := range items {
					items[i] = &model.ContentItem{ID: fmt.Sprintf("item%d", i), ExternalContentID: fmt.Sprintf("EC%d", i)}
				}
				contentProgress := []*sqlc.UserContentProgress{}
				streakProgress := []*sqlc.UserStreakProgress{}
				for _, id := range tc.completed {
					contentProgress = append(contentProgress, &sqlc.UserContentProgress{ExternalContentID: id})
					streakProgress = append(streakProgress, &sqlc.UserStreakProgress{ExternalContentID: id})
				}
				c.Set(cache.ContentItemsByAchievementKey("AC1"), items)
				c.Set(cache.StreakItemsByAchievementKey("AC1"), items)
				c.Set(cache.UserContentProgressKey("US1", "AC1"), contentProgress)
				c.Set(cache.UserStreakProgressKey("US1", "AC1"), streakProgress)
				c.Set(cache.UserContentProgressKey("US2", "AC1"), []*sqlc.UserContentProgress{})
				c.Set(cache.UserStreakProgressKey("US2", "AC1"), []*sqlc.UserStreakProgress{})
				c.Wait()
				r := &Resolver{Loaders: loaders.NewLoaders(nil, c)}
				ctx := context.WithValue(context.Background(), middleware.UserIDKey, tc.userID)
				var total, completed int
				var completedItems []model.ContentItem
				if kind == "content" {
					obj := &model.ContentAchievement{ID: "AC1", TotalItems: 999}
					total, err = r.ContentAchievement().TotalItems(ctx, obj)
					require.NoError(t, err)
					completed, err = r.ContentAchievement().CompletedItemCount(ctx, obj)
					require.NoError(t, err)
					completedItems, err = r.ContentAchievement().UserCompletedItems(ctx, obj)
				} else {
					obj := &model.StreakAchievement{ID: "AC1", TotalItems: 999}
					total, err = r.StreakAchievement().TotalItems(ctx, obj)
					require.NoError(t, err)
					completed, err = r.StreakAchievement().CompletedItemCount(ctx, obj)
					require.NoError(t, err)
					completedItems, err = r.StreakAchievement().UserCompletedItems(ctx, obj)
				}
				require.NoError(t, err)
				assert.Equal(t, tc.total, total)
				assert.Equal(t, tc.want, completed)
				assert.Len(t, completedItems, completed)
			})
		}
	}
}
