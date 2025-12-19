package push

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/loaders"
)

// SendTranslatedAchievementNotification sends a push notification for an achievement award
// with translated content based on the user's language preference.
// This is fire-and-forget - errors are logged but do not propagate.
// It should be called in a goroutine to avoid blocking the main flow.
func SendTranslatedAchievementNotification(
	pushService *Service,
	loadersInstance *loaders.Loaders,
	userID string,
	achievement AchievementInfo,
) {
	if pushService == nil || !pushService.IsConfigured() || loadersInstance == nil {
		return
	}

	bgCtx := context.Background()

	// Get user's language
	userLang := "no" // default
	userThunk := loadersInstance.UserByIDLoader.Load(bgCtx, userID)
	if user, err := userThunk(); err == nil && user != nil {
		userLang = user.Language
	}

	// Get achievement translation
	name := achievement.Name
	notificationText := achievement.NotificationText
	transKey := loaders.TranslationKey{
		EntityType: "achievement",
		EntityID:   achievement.ID,
		LangCode:   userLang,
	}
	transThunk := loadersInstance.TranslationLoader.Load(bgCtx, transKey)
	if trans, err := transThunk(); err == nil && trans != nil {
		if trans.Name != nil && *trans.Name != "" {
			name = *trans.Name
		}
		if trans.NotificationText != nil && *trans.NotificationText != "" {
			notificationText = *trans.NotificationText
		}
	}

	pushService.SendAchievementNotification(bgCtx, userID, AchievementInfo{
		ID:               achievement.ID,
		Name:             name,
		NotificationText: notificationText,
		ImageCompleted:   achievement.ImageCompleted,
	})
}
