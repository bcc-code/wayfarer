package push

import (
	"context"

	"github.com/bcc-media/wayfarer/i18n"
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

// SendTranslatedChallengeEnrollmentNotification sends a push notification for challenge enrollment
// with translated content based on the user's language preference.
// This is fire-and-forget - errors are logged but do not propagate.
// Only sends if notification_text is non-empty (either base or translated).
// It should be called in a goroutine to avoid blocking the main flow.
func SendTranslatedChallengeEnrollmentNotification(
	pushService *Service,
	loadersInstance *loaders.Loaders,
	userID string,
	challenge ChallengeInfo,
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

	// Get challenge translation
	name := challenge.Name
	notificationText := challenge.NotificationText
	transKey := loaders.TranslationKey{
		EntityType: "challenge",
		EntityID:   challenge.ID,
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

	pushService.SendChallengeEnrollmentNotification(bgCtx, userID, ChallengeInfo{
		ID:               challenge.ID,
		Name:             name,
		NotificationText: notificationText,
		Image:            challenge.Image,
	})
}

// SendTranslatedBetResultNotification sends a push notification for a bet result
// with translated content based on the user's language preference.
// This is fire-and-forget - errors are logged but do not propagate.
// It should be called in a goroutine to avoid blocking the main flow.
// No notification is sent if points is 0 (no change).
func SendTranslatedBetResultNotification(
	pushService *Service,
	loadersInstance *loaders.Loaders,
	userID string,
	challengeID string,
	quizID string,
	quizName string,
	points int,
) {
	if pushService == nil || !pushService.IsConfigured() || loadersInstance == nil {
		return
	}

	// Don't send notification if there's no change in points
	if points == 0 {
		return
	}

	bgCtx := context.Background()

	// Get user's language
	userLang := "nb" // default
	userThunk := loadersInstance.UserByIDLoader.Load(bgCtx, userID)
	if user, err := userThunk(); err == nil && user != nil {
		userLang = user.Language
	}

	// Get translated title and message
	title, message := i18n.FormatBetResultMessage(userLang, points)

	pushService.SendBetResultNotification(bgCtx, userID, BetResultInfo{
		ChallengeID: challengeID,
		QuizID:      quizID,
		QuizName:    quizName,
		Points:      points,
		Title:       title,
		Message:     message,
	})
}
