package api

import (
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/services/push"
)

// isAchievementAwardable checks if an achievement can be awarded based on awardable_from timestamp.
// Returns nil if awardable, error if not.
func isAchievementAwardable(awardableFrom *scalars.DateTime) error {
	if awardableFrom != nil && awardableFrom.Time.After(time.Now()) {
		return fmt.Errorf("achievement is not yet available for awarding")
	}
	return nil
}

// getAchievementAwardableFrom extracts the AwardableFrom field from any Achievement type
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

// getAchievementProjectID extracts the ProjectID field from any Achievement type
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

// getAchievementEventID extracts the EventID field from any Achievement type
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

// getAchievementPushInfo extracts push notification info from any Achievement type
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
