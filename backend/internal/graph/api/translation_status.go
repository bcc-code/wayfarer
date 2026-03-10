package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// Helper to check a *bool from sqlc (nullable boolean from SQL expression)
func boolTrue(b *bool) bool {
	return b != nil && *b
}

// challengeTranslationStatus fetches translation status for a challenge by ID.
func (r *Resolver) challengeTranslationStatus(ctx context.Context, challengeID string) ([]model.TranslationFieldStatus, error) {
	rows, err := r.DB.Queries.GetChallengeTranslationStatus(ctx, challengeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge translation status: %w", err)
	}

	result := make([]model.TranslationFieldStatus, 0, len(rows))
	for _, row := range rows {
		var fields []string
		if boolTrue(row.HasName) {
			fields = append(fields, "name")
		}
		if boolTrue(row.HasDescription) {
			fields = append(fields, "description")
		}
		if boolTrue(row.HasButtonText) {
			fields = append(fields, "buttonText")
		}
		if boolTrue(row.HasNotificationText) {
			fields = append(fields, "notificationText")
		}
		if len(fields) > 0 {
			result = append(result, model.TranslationFieldStatus{
				LanguageCode: row.LanguageCode,
				Fields:       fields,
			})
		}
	}
	return result, nil
}

// projectTranslationStatus fetches translation status for a project by ID.
func (r *Resolver) projectTranslationStatus(ctx context.Context, projectID string) ([]model.TranslationFieldStatus, error) {
	rows, err := r.DB.Queries.GetProjectTranslationStatus(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project translation status: %w", err)
	}

	result := make([]model.TranslationFieldStatus, 0, len(rows))
	for _, row := range rows {
		var fields []string
		if boolTrue(row.HasName) {
			fields = append(fields, "name")
		}
		if boolTrue(row.HasDescription) {
			fields = append(fields, "description")
		}
		if boolTrue(row.HasRules) {
			fields = append(fields, "rules")
		}
		if len(fields) > 0 {
			result = append(result, model.TranslationFieldStatus{
				LanguageCode: row.LanguageCode,
				Fields:       fields,
			})
		}
	}
	return result, nil
}

// eventTranslationStatus fetches translation status for an event by ID.
func (r *Resolver) eventTranslationStatus(ctx context.Context, eventID string) ([]model.TranslationFieldStatus, error) {
	rows, err := r.DB.Queries.GetEventTranslationStatus(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event translation status: %w", err)
	}

	result := make([]model.TranslationFieldStatus, 0, len(rows))
	for _, row := range rows {
		var fields []string
		if boolTrue(row.HasName) {
			fields = append(fields, "name")
		}
		if boolTrue(row.HasDescription) {
			fields = append(fields, "description")
		}
		if len(fields) > 0 {
			result = append(result, model.TranslationFieldStatus{
				LanguageCode: row.LanguageCode,
				Fields:       fields,
			})
		}
	}
	return result, nil
}

// streakTranslationStatus fetches translation status for a streak by ID.
func (r *Resolver) streakTranslationStatus(ctx context.Context, streakID string) ([]model.TranslationFieldStatus, error) {
	rows, err := r.DB.Queries.GetStreakTranslationStatus(ctx, streakID)
	if err != nil {
		return nil, fmt.Errorf("failed to get streak translation status: %w", err)
	}

	result := make([]model.TranslationFieldStatus, 0, len(rows))
	for _, row := range rows {
		var fields []string
		if boolTrue(row.HasName) {
			fields = append(fields, "name")
		}
		if boolTrue(row.HasDescription) {
			fields = append(fields, "description")
		}
		if len(fields) > 0 {
			result = append(result, model.TranslationFieldStatus{
				LanguageCode: row.LanguageCode,
				Fields:       fields,
			})
		}
	}
	return result, nil
}

// achievementTranslationStatus fetches translation status for an achievement by ID.
func (r *Resolver) achievementTranslationStatus(ctx context.Context, achievementID string) ([]model.TranslationFieldStatus, error) {
	rows, err := r.DB.Queries.GetAchievementTranslationStatus(ctx, achievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement translation status: %w", err)
	}

	result := make([]model.TranslationFieldStatus, 0, len(rows))
	for _, row := range rows {
		var fields []string
		if boolTrue(row.HasName) {
			fields = append(fields, "name")
		}
		if boolTrue(row.HasDescriptionPending) {
			fields = append(fields, "descriptionPending")
		}
		if boolTrue(row.HasDescriptionCompleted) {
			fields = append(fields, "descriptionCompleted")
		}
		if boolTrue(row.HasNotificationText) {
			fields = append(fields, "notificationText")
		}
		if len(fields) > 0 {
			result = append(result, model.TranslationFieldStatus{
				LanguageCode: row.LanguageCode,
				Fields:       fields,
			})
		}
	}
	return result, nil
}

// consentTranslationStatus fetches translation status for a consent by ID.
func (r *Resolver) consentTranslationStatus(ctx context.Context, consentID string) ([]model.TranslationFieldStatus, error) {
	rows, err := r.DB.Queries.GetConsentTranslationStatus(ctx, consentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get consent translation status: %w", err)
	}

	result := make([]model.TranslationFieldStatus, 0, len(rows))
	for _, row := range rows {
		var fields []string
		if boolTrue(row.HasTitle) {
			fields = append(fields, "title")
		}
		if boolTrue(row.HasShortText) {
			fields = append(fields, "shortText")
		}
		if boolTrue(row.HasBody) {
			fields = append(fields, "body")
		}
		if len(fields) > 0 {
			result = append(result, model.TranslationFieldStatus{
				LanguageCode: row.LanguageCode,
				Fields:       fields,
			})
		}
	}
	return result, nil
}

// quizQuestionTranslationStatus fetches translation status for a quiz question by ID.
func (r *Resolver) quizQuestionTranslationStatus(ctx context.Context, questionID string) ([]model.TranslationFieldStatus, error) {
	rows, err := r.DB.Queries.GetQuizQuestionTranslationStatus(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz question translation status: %w", err)
	}

	result := make([]model.TranslationFieldStatus, 0, len(rows))
	for _, row := range rows {
		var fields []string
		if boolTrue(row.HasQuestionText) {
			fields = append(fields, "questionText")
		}
		if len(fields) > 0 {
			result = append(result, model.TranslationFieldStatus{
				LanguageCode: row.LanguageCode,
				Fields:       fields,
			})
		}
	}
	return result, nil
}

// quizAnswerTranslationStatus fetches translation status for a quiz predefined answer by ID.
func (r *Resolver) quizAnswerTranslationStatus(ctx context.Context, answerID string) ([]model.TranslationFieldStatus, error) {
	rows, err := r.DB.Queries.GetQuizAnswerTranslationStatus(ctx, answerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz answer translation status: %w", err)
	}

	result := make([]model.TranslationFieldStatus, 0, len(rows))
	for _, row := range rows {
		var fields []string
		if boolTrue(row.HasAnswerText) {
			fields = append(fields, "answerText")
		}
		if len(fields) > 0 {
			result = append(result, model.TranslationFieldStatus{
				LanguageCode: row.LanguageCode,
				Fields:       fields,
			})
		}
	}
	return result, nil
}

// quizTranslationStatus fetches translation status for a quiz by ID.
func (r *Resolver) quizTranslationStatus(ctx context.Context, quizID string) ([]model.TranslationFieldStatus, error) {
	rows, err := r.DB.Queries.GetQuizTranslationStatus(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz translation status: %w", err)
	}

	result := make([]model.TranslationFieldStatus, 0, len(rows))
	for _, row := range rows {
		var fields []string
		if boolTrue(row.HasName) {
			fields = append(fields, "name")
		}
		if boolTrue(row.HasDescription) {
			fields = append(fields, "description")
		}
		if len(fields) > 0 {
			result = append(result, model.TranslationFieldStatus{
				LanguageCode: row.LanguageCode,
				Fields:       fields,
			})
		}
	}
	return result, nil
}
