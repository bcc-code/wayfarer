package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services/quizgrading"
)

// gradeQuizResponse grades a response, loading the question's answers
// (Ristretto-cached static data) only for question types that are graded.
// Returns nil when the response is not graded.
func gradeQuizResponse(ctx context.Context, l *loaders.Loaders, questionType, questionID string, response quizgrading.Response) (*bool, error) {
	if !quizgrading.IsGradable(questionType) {
		return nil, nil
	}

	answers, err := l.QuizAnswersByQuestionLoader.Load(ctx, questionID)()
	if err != nil {
		return nil, fmt.Errorf("failed to load answers: %w", err)
	}

	return quizgrading.Grade(questionType, response, toGradingAnswers(answers)), nil
}

// toGradingAnswers converts loaded answers, keeping their answer_order.
func toGradingAnswers(answers []*model.QuizPredefinedAnswer) []quizgrading.Answer {
	result := make([]quizgrading.Answer, len(answers))
	for i, a := range answers {
		result[i] = quizgrading.Answer{ID: a.ID, IsCorrect: a.IsCorrectValue}
	}
	return result
}
