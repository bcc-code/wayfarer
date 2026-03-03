package translations

import (
	"context"

	"github.com/bcc-media/wayfarer/common"
	"gopkg.in/guregu/null.v4"
)

// Export handlers - fetch data from database and prepare for translation

func (s *Service) getDataForProjects(ctx context.Context) ([]common.TranslationData, error) {
	rows, err := s.queries.GetProjectsForTranslation(ctx)
	if err != nil {
		return nil, err
	}

	toSend := []common.TranslationData{}
	for _, row := range rows {
		value := ProjectTranslation{
			Name:        row.Name,
			Description: splitLines(row.Description),
			Rules:       splitLinesPtr(row.Rules),
		}
		toSend = append(toSend, common.TranslationData{
			Language: "no", // Base language is Norwegian
			Value:    mustToJSON(value),
			ID:       row.ID, // ULID with PR prefix
		})
	}

	return toSend, nil
}

func (s *Service) getDataForEvents(ctx context.Context) ([]common.TranslationData, error) {
	rows, err := s.queries.GetEventsForTranslation(ctx)
	if err != nil {
		return nil, err
	}

	toSend := []common.TranslationData{}
	for _, row := range rows {
		value := NameDescriptionTranslation{
			Name:        row.Name,
			Description: null.StringFrom(row.Description),
		}
		toSend = append(toSend, common.TranslationData{
			Language: "no",
			Value:    mustToJSON(value),
			ID:       row.ID, // ULID with EV prefix
		})
	}

	return toSend, nil
}

func (s *Service) getDataForStreaks(ctx context.Context) ([]common.TranslationData, error) {
	rows, err := s.queries.GetStreaksForTranslation(ctx)
	if err != nil {
		return nil, err
	}

	toSend := []common.TranslationData{}
	for _, row := range rows {
		value := NameDescriptionTranslation{
			Name:        row.Name,
			Description: null.StringFrom(row.Description),
		}
		toSend = append(toSend, common.TranslationData{
			Language: "no",
			Value:    mustToJSON(value),
			ID:       row.ID, // ULID with SK prefix
		})
	}

	return toSend, nil
}

func (s *Service) getDataForChallenges(ctx context.Context) ([]common.TranslationData, error) {
	rows, err := s.queries.GetChallengesForTranslation(ctx)
	if err != nil {
		return nil, err
	}

	toSend := []common.TranslationData{}
	for _, row := range rows {
		value := ChallengeTranslation{
			Name:             row.Name,
			Description:      null.StringFrom(row.Description),
			ButtonText:       row.ButtonText,
			NotificationText: null.StringFrom(row.NotificationText),
		}
		toSend = append(toSend, common.TranslationData{
			Language: "no",
			Value:    mustToJSON(value),
			ID:       row.ID, // ULID with CL prefix
		})
	}

	return toSend, nil
}

func (s *Service) getDataForAchievements(ctx context.Context) ([]common.TranslationData, error) {
	rows, err := s.queries.GetAchievementsForTranslation(ctx)
	if err != nil {
		return nil, err
	}

	toSend := []common.TranslationData{}
	for _, row := range rows {
		value := AchievementTranslation{
			Name:                 row.Name,
			DescriptionPending:   row.DescriptionPending,
			DescriptionCompleted: row.DescriptionCompleted,
			NotificationText:     null.StringFrom(row.NotificationText),
		}
		toSend = append(toSend, common.TranslationData{
			Language: "no",
			Value:    mustToJSON(value),
			ID:       row.ID, // ULID with AC prefix
		})
	}

	return toSend, nil
}

func (s *Service) getDataForQuizzes(ctx context.Context) ([]common.TranslationData, error) {
	rows, err := s.queries.GetQuizzesForTranslation(ctx)
	if err != nil {
		return nil, err
	}

	toSend := []common.TranslationData{}
	for _, row := range rows {
		// Get questions for this quiz
		questions, err := s.queries.GetQuizQuestionsForTranslation(ctx, row.ID)
		if err != nil {
			continue
		}

		questionItems := make([]QuestionWithID, 0, len(questions))
		for _, q := range questions {
			// Get answers for this question
			answers, err := s.queries.GetQuizAnswersForTranslation(ctx, q.ID)
			if err != nil {
				continue
			}

			answerItems := make([]AnswerWithID, 0, len(answers))
			for _, a := range answers {
				answerItems = append(answerItems, AnswerWithID{
					AnswerText: a.AnswerText,
					ID:         a.ID,
				})
			}

			questionItems = append(questionItems, QuestionWithID{
				QuestionText: q.QuestionText,
				ID:           q.ID,
				Answers:      answerItems,
			})
		}

		value := QuizTranslation{
			Name:        row.Name,
			Description: null.StringFrom(row.Description),
			Questions:   questionItems,
		}

		toSend = append(toSend, common.TranslationData{
			Language: "no",
			Value:    mustToJSON(value),
			ID:       row.ID,
		})
	}

	return toSend, nil
}

func (s *Service) getDataForConsents(ctx context.Context) ([]common.TranslationData, error) {
	rows, err := s.queries.GetConsentsForTranslation(ctx)
	if err != nil {
		return nil, err
	}

	toSend := []common.TranslationData{}
	for _, row := range rows {
		value := ConsentTranslation{
			Title:     row.Title,
			ShortText: row.ShortText,
			Body:      splitLines(row.Body),
		}
		toSend = append(toSend, common.TranslationData{
			Language: "no",
			Value:    mustToJSON(value),
			ID:       row.ID,
		})
	}

	return toSend, nil
}
