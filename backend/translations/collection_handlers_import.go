package translations

import (
	"context"
	"encoding/json"

	"github.com/bcc-media/wayfarer/common"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// Import handlers - receive translated data from Phrase and update database

func (s *Service) updateProjects(ctx context.Context, data []common.TranslationData) []error {
	errs := make([]error, 0)
	for _, d := range data {
		value := &ProjectTranslation{}
		err := json.Unmarshal(d.Value, value)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = s.queries.UpsertProjectTranslation(ctx, sqlc.UpsertProjectTranslationParams{
			ProjectID:    d.ID,
			LanguageCode: d.Language,
			Name:         value.Name,
			Description:  joinLines(value.Description),
			Rules:        joinLines(value.Rules),
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *Service) updateEvents(ctx context.Context, data []common.TranslationData) []error {
	errs := make([]error, 0)
	for _, d := range data {
		value := &NameDescriptionTranslation{}
		err := json.Unmarshal(d.Value, value)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = s.queries.UpsertEventTranslation(ctx, sqlc.UpsertEventTranslationParams{
			EventID:      d.ID,
			LanguageCode: d.Language,
			Name:         value.Name,
			Description:  value.Description.String,
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *Service) updateStreaks(ctx context.Context, data []common.TranslationData) []error {
	errs := make([]error, 0)
	for _, d := range data {
		value := &NameDescriptionTranslation{}
		err := json.Unmarshal(d.Value, value)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = s.queries.UpsertStreakTranslation(ctx, sqlc.UpsertStreakTranslationParams{
			StreakID:     d.ID,
			LanguageCode: d.Language,
			Name:         value.Name,
			Description:  value.Description.String,
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *Service) updateChallenges(ctx context.Context, data []common.TranslationData) []error {
	errs := make([]error, 0)
	for _, d := range data {
		value := &ChallengeTranslation{}
		err := json.Unmarshal(d.Value, value)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = s.queries.UpsertChallengeTranslation(ctx, sqlc.UpsertChallengeTranslationParams{
			ChallengeID:      d.ID,
			LanguageCode:     d.Language,
			Name:             value.Name,
			Description:      value.Description.String,
			ButtonText:       value.ButtonText,
			NotificationText: value.NotificationText.String,
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *Service) updateAchievements(ctx context.Context, data []common.TranslationData) []error {
	errs := make([]error, 0)
	for _, d := range data {
		value := &AchievementTranslation{}
		err := json.Unmarshal(d.Value, value)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = s.queries.UpsertAchievementTranslation(ctx, sqlc.UpsertAchievementTranslationParams{
			AchievementID:        d.ID,
			LanguageCode:         d.Language,
			Name:                 value.Name,
			DescriptionPending:   value.DescriptionPending,
			DescriptionCompleted: value.DescriptionCompleted,
			NotificationText:     value.NotificationText.String,
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *Service) updateQuizzes(ctx context.Context, data []common.TranslationData) []error {
	errs := make([]error, 0)
	for _, d := range data {
		value := &QuizTranslation{}
		err := json.Unmarshal(d.Value, value)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// Update quiz translation
		err = s.queries.UpsertQuizTranslation(ctx, sqlc.UpsertQuizTranslationParams{
			QuizID:       d.ID,
			LanguageCode: d.Language,
			Name:         value.Name,
			Description:  value.Description.String,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// Update question translations
		for _, q := range value.Questions {
			err = s.queries.UpsertQuizQuestionTranslation(ctx, sqlc.UpsertQuizQuestionTranslationParams{
				QuestionID:   q.ID,
				LanguageCode: d.Language,
				QuestionText: q.QuestionText,
			})
			if err != nil {
				errs = append(errs, err)
			}

			// Update answer translations
			for _, a := range q.Answers {
				err = s.queries.UpsertQuizAnswerTranslation(ctx, sqlc.UpsertQuizAnswerTranslationParams{
					AnswerID:     a.ID,
					LanguageCode: d.Language,
					AnswerText:   a.AnswerText,
				})
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	return errs
}

func (s *Service) updateConsents(ctx context.Context, data []common.TranslationData) []error {
	errs := make([]error, 0)
	for _, d := range data {
		value := &ConsentTranslation{}
		err := json.Unmarshal(d.Value, value)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = s.queries.UpsertConsentTranslationFromPhrase(ctx, sqlc.UpsertConsentTranslationFromPhraseParams{
			ConsentID:    d.ID,
			LanguageCode: d.Language,
			Title:        value.Title,
			ShortText:    value.ShortText,
			Body:         joinLines(value.Body),
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
