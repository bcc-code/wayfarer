package api

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
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
	if err != nil {
		var mpErr *quizgrading.MapPinError
		if errors.As(err, &mpErr) {
			return c, &BetValidationError{Field: "mapPin." + mpErr.Field, Message: mpErr.Message}
		}
		return c, err
	}
	return c, nil
}

// MAP_PIN always awards points, scored by MapPinScore, so it can't have a bet:
// a settled bet would overwrite the points (quiz_questions_map_pin_check).
func mapPinBettingError() error {
	return &BetValidationError{Field: "bettingEnabled", Message: "MAP_PIN questions can't have betting"}
}

func mapPinPointsError() error {
	return &BetValidationError{Field: "points", Message: "must be greater than 0 for MAP_PIN questions"}
}

// checkMapPinCreate applies the MAP_PIN rules to a new question. It returns the
// validated config for a MAP_PIN question and nil for any other type.
func checkMapPinCreate(input model.CreateQuizQuestionInput) (*quizgrading.MapPinConfig, error) {
	if input.QuestionType != model.QuizQuestionTypeMapPin {
		if input.MapPin != nil {
			return nil, &BetValidationError{Field: "mapPin", Message: "is only used by MAP_PIN questions"}
		}
		return nil, nil
	}

	if input.MapPin == nil {
		return nil, &BetValidationError{Field: "mapPin", Message: "is required for MAP_PIN questions"}
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
			return mapPinUpdate{}, &BetValidationError{Field: "mapPin", Message: "is only used by MAP_PIN questions"}
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
	return &BetValidationError{
		Field:   field,
		Message: "can't be changed after users have answered this question: their points would not be recalculated. Delete it and add a new one",
	}
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
