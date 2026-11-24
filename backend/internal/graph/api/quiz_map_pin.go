package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services/quizgrading"
)

// mapPinConfigFromInput converts the GraphQL input as given. It is checked by
// ValidateMapPinConfig, which needs to see the other mode's fields to reject them.
func mapPinConfigFromInput(in *model.MapPinConfigInput) quizgrading.MapPinConfig {
	c := quizgrading.MapPinConfig{
		Mode:       string(in.ScoringMode),
		Width:      in.MapWidth,
		Height:     in.MapHeight,
		FullRadius: in.FullPointsRadius,
		ZeroRadius: in.ZeroPointsRadius,
	}
	if in.Target != nil {
		c.Target = &quizgrading.MapPoint{X: in.Target.X, Y: in.Target.Y}
	}
	if in.Polygon != nil {
		c.Polygon = make([]quizgrading.MapPoint, len(in.Polygon))
		for i, p := range in.Polygon {
			c.Polygon[i] = quizgrading.MapPoint{X: p.X, Y: p.Y}
		}
	}
	return c
}

// validateMapPinInput validates the input and returns it normalized, with
// errors naming the input field ("mapPin.polygon[2]").
func validateMapPinInput(in *model.MapPinConfigInput) (quizgrading.MapPinConfig, error) {
	c, err := quizgrading.ValidateMapPinConfig(mapPinConfigFromInput(in))
	return c, mapPinFieldError("mapPin.", err)
}

// MAP_PIN always awards points, scored by MapPinScore, so it can't have a bet:
// a settled bet would overwrite the points (quiz_questions_map_pin_check).
func mapPinBettingError() error {
	return fieldError("bettingEnabled", "MAP_PIN questions can't have betting")
}

func mapPinPointsError() error {
	return fieldError("points", "must be greater than 0 for MAP_PIN questions")
}

// checkMapPinCreate applies the MAP_PIN rules to a new question. It returns the
// validated config for a MAP_PIN question and nil for any other type.
func checkMapPinCreate(input model.CreateQuizQuestionInput) (*quizgrading.MapPinConfig, error) {
	if input.QuestionType != model.QuizQuestionTypeMapPin {
		if input.MapPin != nil {
			return nil, fieldError("mapPin", "is only used by MAP_PIN questions")
		}
		return nil, nil
	}

	if input.MapPin == nil {
		return nil, fieldError("mapPin", "is required for MAP_PIN questions")
	}
	if input.Points == nil || *input.Points <= 0 {
		return nil, mapPinPointsError()
	}
	if input.BettingEnabled != nil && *input.BettingEnabled {
		return nil, mapPinBettingError()
	}
	c, err := validateMapPinInput(input.MapPin)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// mapPinUpdate is what an update does to a question's map config.
type mapPinUpdate struct {
	// Config replaces the stored config; nil leaves it alone.
	Config *quizgrading.MapPinConfig
	// LockedField is set when the update changes how pins score ("mapPin" or
	// "points"). Nothing regrades stored responses, so that is only allowed
	// while nobody has answered the question.
	LockedField string
}

// checkMapPinUpdate applies the MAP_PIN rules to an update of a stored question.
// An unchanged config or points value doesn't lock: an admin form resends the
// whole question to fix a typo.
func checkMapPinUpdate(stored *sqlc.GetQuizQuestionByIDRow, input model.UpdateQuizQuestionInput) (mapPinUpdate, error) {
	if stored.QuestionType != "MAP_PIN" {
		if input.MapPin != nil {
			return mapPinUpdate{}, fieldError("mapPin", "is only used by MAP_PIN questions")
		}
		return mapPinUpdate{}, nil
	}

	if input.BettingEnabled != nil && *input.BettingEnabled {
		return mapPinUpdate{}, mapPinBettingError()
	}

	var update mapPinUpdate
	if input.Points != nil {
		if *input.Points <= 0 {
			return mapPinUpdate{}, mapPinPointsError()
		}
		if stored.Points == nil || int(*stored.Points) != *input.Points {
			update.LockedField = "points"
		}
	}
	if input.MapPin != nil {
		c, err := validateMapPinInput(input.MapPin)
		if err != nil {
			return mapPinUpdate{}, err
		}
		update.Config = &c
		// An unreadable stored config compares unequal, so it locks too.
		current, err := getQuizQuestionByIDRowAdapter{stored}.GetMapPinColumns().Config()
		if err != nil || !current.Equal(c) {
			update.LockedField = "mapPin"
		}
	}
	return update, nil
}

// mapPinAnsweredError rejects a scoring change on an answered question.
func mapPinAnsweredError(field string) error {
	return fieldError(field, "can't be changed after users have answered this question: their points would not be recalculated. Delete it and add a new one")
}

// convertToMapPinQuestion builds a MapPinQuestion. A config that can't be read
// is logged and kept as far as it goes: the question still loads, and scoring
// rejects the incomplete config instead of awarding wrong points.
func convertToMapPinQuestion(row quizQuestionRow) *model.MapPinQuestion {
	config, err := row.GetMapPinColumns().Config()
	if err != nil {
		slog.Error("failed to read MAP_PIN config", "question_id", row.GetID(), "error", err)
	}
	return &model.MapPinQuestion{
		ID:                       row.GetID(),
		QuizID:                   row.GetQuizID(),
		QuestionText:             row.GetQuestionText(),
		ImageURL:                 row.GetImageURL(),
		QuestionOrder:            int(row.GetQuestionOrder()),
		TimeoutSeconds:           convertQuestionTimeoutSeconds(row.GetTimeoutSeconds()),
		Points:                   convertQuestionPoints(row.GetPoints()),
		BettingEnabled:           row.GetBettingEnabled(),
		BettingMinPercentage:     convertBettingPercentage(row.GetBettingMinPercentage()),
		BettingMaxPercentage:     convertBettingPercentage(row.GetBettingMaxPercentage()),
		BettingMinAbsolute:       convertBettingAbsolute(row.GetBettingMinAbsolute()),
		BettingMaxAbsolute:       convertBettingAbsolute(row.GetBettingMaxAbsolute()),
		BettingMultiplierCorrect: convertBettingPercentage(row.GetBettingMultiplierCorrect()),
		BettingMultiplierWrong:   convertBettingPercentage(row.GetBettingMultiplierWrong()),
		MapPin:                   config,
	}
}

func convertMapPoint(p quizgrading.MapPoint) model.MapPoint {
	return model.MapPoint{X: p.X, Y: p.Y}
}

// correctMapPosition returns a POINT_RADIUS question's target, or nil when the
// question is POLYGON or the user may not see it yet.
func (r *Resolver) correctMapPosition(ctx context.Context, q *model.MapPinQuestion) (*model.MapPoint, error) {
	if q.MapPin.Target == nil {
		return nil, nil
	}
	ok, err := r.canRevealMapPinGeometry(ctx, q.QuizID)
	if err != nil || !ok {
		return nil, err
	}
	p := convertMapPoint(*q.MapPin.Target)
	return &p, nil
}

// correctMapArea returns a POLYGON question's polygon, or nil when the question
// is POINT_RADIUS or the user may not see it yet.
func (r *Resolver) correctMapArea(ctx context.Context, q *model.MapPinQuestion) ([]model.MapPoint, error) {
	if q.MapPin.Polygon == nil {
		return nil, nil
	}
	ok, err := r.canRevealMapPinGeometry(ctx, q.QuizID)
	if err != nil || !ok {
		return nil, err
	}
	area := make([]model.MapPoint, len(q.MapPin.Polygon))
	for i, p := range q.MapPin.Polygon {
		area[i] = convertMapPoint(p)
	}
	return area, nil
}

// mapPinAnswer validates a MAP_PIN answer against the question's map and
// returns the pin and the JSON it is stored as in json_response.
func mapPinAnswer(position *model.MapPointInput, c quizgrading.MapPinConfig) (quizgrading.MapPoint, []byte, error) {
	if position == nil {
		return quizgrading.MapPoint{}, nil, fieldError("mapPosition", "is required for MAP_PIN questions")
	}
	p := quizgrading.MapPoint{X: position.X, Y: position.Y}
	if err := quizgrading.ValidateMapPosition(p, c.Width, c.Height); err != nil {
		return p, nil, mapPinFieldError("", err)
	}
	stored, err := json.Marshal(p)
	if err != nil {
		return p, nil, fmt.Errorf("failed to encode map position: %w", err)
	}
	return p, stored, nil
}

// checkMapPinAnswerFields rejects mixing MAP_PIN with other answer fields on
// submit: a MAP_PIN question takes only mapPosition, and only MAP_PIN takes it,
// so the stored JSON is always a validated pin. The error names the offending
// input field; with several, the first in input order.
func checkMapPinAnswerFields(questionType string, input model.SubmitQuizAnswerInput) error {
	if questionType != "MAP_PIN" {
		if input.MapPosition != nil {
			return fieldError("mapPosition", fmt.Sprintf("cannot set a MAP_PIN answer on a %s question", questionType))
		}
		return nil
	}
	for _, f := range []struct {
		field, questionType string
		present             bool
	}{
		{"selectedAnswerIds", "PREDEFINED", input.SelectedAnswerIds != nil},
		{"textResponse", "FREE_TEXT", input.TextResponse != nil},
		{"numberResponse", "NUMBER", input.NumberResponse != nil},
		{"jsonResponse", "JSON", input.JSONResponse != nil},
		{"submittedOrder", "ORDERING", input.SubmittedOrder != nil},
	} {
		if f.present {
			return fieldError(f.field, fmt.Sprintf("cannot set a %s answer on a MAP_PIN question", f.questionType))
		}
	}
	return nil
}

// convertToMapPinResponse builds a MapPinResponse. json_response holds a pin
// validated on the way in. One that can't be read is logged and left nil, so
// reading submittedPosition fails instead of showing a pin nobody placed.
func convertToMapPinResponse(row *sqlc.QuizResponse, answeredAt *scalars.DateTime, timeSpentSeconds *int, pointsEarned *int, betAmount *int) *model.MapPinResponse {
	return &model.MapPinResponse{
		ID:                row.ID,
		SubmissionID:      row.SubmissionID,
		QuestionID:        row.QuestionID,
		SubmittedPosition: loaders.MapPinPosition(row.ID, row.JsonResponse),
		AnsweredAt:        answeredAt,
		TimeSpentSeconds:  timeSpentSeconds,
		PointsEarnedValue: pointsEarned,
		BetAmount:         betAmount,
		ScoreJournalID:    row.ScoreJournalID,
	}
}

// mapPinResponseDistance returns a POINT_RADIUS pin's distance to the target,
// computed by MapPinScore from the cached question, or nil for POLYGON or
// while the session is OPEN.
func (r *Resolver) mapPinResponseDistance(ctx context.Context, response *model.MapPinResponse) (*float64, error) {
	// Null distance is a valid answer (POLYGON); a missing pin is a broken model.
	if response.SubmittedPosition == nil {
		return nil, fmt.Errorf("MAP_PIN response %s has no submitted position", response.ID)
	}
	visible, err := r.responseResultVisible(ctx, response.SubmissionID)
	if err != nil || !visible {
		return nil, err
	}

	submission, err := r.Loaders.QuizSubmissionByIDLoader.Load(ctx, response.SubmissionID)()
	if err != nil {
		return nil, fmt.Errorf("failed to load submission: %w", err)
	}
	if submission == nil {
		return nil, nil
	}
	questions, err := r.Loaders.QuizQuestionsByQuizLoader.Load(ctx, submission.QuizID)()
	if err != nil {
		return nil, fmt.Errorf("failed to load questions: %w", err)
	}
	for _, q := range questions {
		mp, ok := q.(*model.MapPinQuestion)
		if !ok || mp.ID != response.QuestionID {
			continue
		}
		if mp.MapPin.Mode != quizgrading.MapPinModePointRadius {
			return nil, nil
		}
		pin := quizgrading.MapPoint{X: response.SubmittedPosition.X, Y: response.SubmittedPosition.Y}
		_, distance, err := quizgrading.MapPinScore(mp.MapPin, 0, pin)
		if err != nil {
			return nil, fmt.Errorf("failed to measure map pin distance: %w", err)
		}
		return distance, nil
	}
	return nil, nil
}

// mapPinBetResultError rejects recording a bet result on a MAP_PIN response.
// MAP_PIN can't have a bet, and a bet result would overwrite the points the
// pin was scored.
func mapPinBetResultError(responseID string) error {
	return fieldError("responseId", fmt.Sprintf("%s answers a MAP_PIN question, which has no bet", responseID))
}

// fieldError is invalid input on one input field.
func fieldError(field, message string) error {
	return &FieldValidationError{Field: field, Message: message}
}

// mapPinFieldError turns a quizgrading.MapPinError into a field error, its
// field prefixed with the input it sits in ("mapPin."). Other errors and nil
// pass through.
func mapPinFieldError(prefix string, err error) error {
	var mpErr *quizgrading.MapPinError
	if errors.As(err, &mpErr) {
		return fieldError(prefix+mpErr.Field, mpErr.Message)
	}
	return err
}
