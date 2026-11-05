package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
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

// scoredQuestion is what scoring a response needs to know about its question.
type scoredQuestion struct {
	Type   string
	ID     string
	Points *int
	MapPin quizgrading.MapPinConfig // MAP_PIN only
}

func scoredQuestionFromModel(q model.QuizQuestion) scoredQuestion {
	s := scoredQuestion{Type: quizQuestionTypeString(q), ID: q.GetID(), Points: q.GetPoints()}
	if mp, ok := q.(*model.MapPinQuestion); ok {
		s.MapPin = mp.MapPin
	}
	return s
}

func scoredQuestionFromRow(q *sqlc.GetQuizQuestionByIDRow) (scoredQuestion, error) {
	s := scoredQuestion{Type: q.QuestionType, ID: q.ID}
	if q.Points != nil {
		points := int(*q.Points)
		s.Points = &points
	}
	if q.QuestionType == "MAP_PIN" {
		mapPin, err := getQuizQuestionByIDRowAdapter{q}.GetMapPinColumns().Config()
		if err != nil {
			return s, fmt.Errorf("failed to read map config: %w", err)
		}
		s.MapPin = mapPin
	}
	return s, nil
}

// scoreQuizResponse returns a response's verdict and the points it earned (nil
// for none). MAP_PIN is scored by MapPinScore and never graded; it must be
// handled before gradeQuizResponse, whose IsGradable check would leave it at 0.
// Every other type earns the question's points when graded correct.
func scoreQuizResponse(ctx context.Context, l *loaders.Loaders, q scoredQuestion, response quizgrading.Response) (*bool, *int32, error) {
	if q.Type == "MAP_PIN" {
		if response.MapPosition == nil {
			return nil, nil, &BetValidationError{Field: "mapPosition", Message: "is required for MAP_PIN questions"}
		}
		maxPoints := 0
		if q.Points != nil {
			maxPoints = *q.Points
		}
		points, _, err := quizgrading.MapPinScore(q.MapPin, maxPoints, *response.MapPosition)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to score map pin: %w", err)
		}
		p := int32(points)
		return nil, &p, nil
	}

	isCorrect, err := gradeQuizResponse(ctx, l, q.Type, q.ID, response)
	if err != nil {
		return nil, nil, err
	}
	if isCorrect != nil && *isCorrect && q.Points != nil {
		p := int32(*q.Points)
		return isCorrect, &p, nil
	}
	return isCorrect, nil, nil
}
