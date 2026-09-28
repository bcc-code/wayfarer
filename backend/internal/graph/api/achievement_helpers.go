package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/push"
)

// checkProjectFinished loads the project by ID and returns an error if it is finished
// (archived or past end_date). Skips the check when force is true.
func checkProjectFinished(ctx context.Context, loadersInstance *loaders.Loaders, projectID string, force bool) error {
	if force {
		return nil
	}
	projectThunk := loadersInstance.ProjectByIDLoader.Load(ctx, projectID)
	project, err := projectThunk()
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}
	return services.IsProjectFinished(project)
}

// isAchievementAwardable checks if an achievement can be awarded based on awardable_from timestamp.
func isAchievementAwardable(awardableFrom *scalars.DateTime) error {
	if awardableFrom != nil && awardableFrom.After(time.Now()) {
		return fmt.Errorf("achievement is not yet available for awarding")
	}
	return nil
}

func getAchievementAwardableFrom(achievement model.Achievement) *scalars.DateTime {
	switch a := achievement.(type) {
	case *model.SimpleAchievement:
		return a.AwardableFrom
	case *model.ContentAchievement:
		return a.AwardableFrom
	case *model.StreakAchievement:
		return a.AwardableFrom
	case *model.QuizAchievement:
		return a.AwardableFrom
	default:
		return nil
	}
}

func getAchievementProjectID(achievement model.Achievement) string {
	switch a := achievement.(type) {
	case *model.SimpleAchievement:
		return a.ProjectID
	case *model.ContentAchievement:
		return a.ProjectID
	case *model.StreakAchievement:
		return a.ProjectID
	case *model.QuizAchievement:
		return a.ProjectID
	default:
		return ""
	}
}

func getAchievementEventID(achievement model.Achievement) *string {
	switch a := achievement.(type) {
	case *model.SimpleAchievement:
		return a.EventID
	case *model.ContentAchievement:
		return a.EventID
	case *model.StreakAchievement:
		return a.EventID
	case *model.QuizAchievement:
		return a.EventID
	default:
		return nil
	}
}

func getAchievementPushInfo(achievement model.Achievement) push.AchievementInfo {
	switch a := achievement.(type) {
	case *model.SimpleAchievement:
		return push.AchievementInfo{
			ID:               a.ID,
			Name:             a.Name,
			NotificationText: a.NotificationText,
			ImageCompleted:   a.ImageCompleted,
		}
	case *model.ContentAchievement:
		return push.AchievementInfo{
			ID:               a.ID,
			Name:             a.Name,
			NotificationText: a.NotificationText,
			ImageCompleted:   a.ImageCompleted,
		}
	case *model.StreakAchievement:
		return push.AchievementInfo{
			ID:               a.ID,
			Name:             a.Name,
			NotificationText: a.NotificationText,
			ImageCompleted:   a.ImageCompleted,
		}
	case *model.QuizAchievement:
		return push.AchievementInfo{
			ID:               a.ID,
			Name:             a.Name,
			NotificationText: a.NotificationText,
			ImageCompleted:   a.ImageCompleted,
		}
	default:
		return push.AchievementInfo{}
	}
}
