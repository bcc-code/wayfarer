package api

import (
	"encoding/json"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/jackc/pgx/v5/pgtype"
)

// quizRow defines the common interface for quiz row types returned by different queries
type quizRow interface {
	GetID() string
	GetProjectID() string
	GetChallengeID() string
	GetName() string
	GetDescription() string
	GetImageUrl() *string
	GetTimeoutSeconds() *int32
	GetQuestionTimeoutSeconds() *int32
	GetRandomizeQuestions() bool
	GetRevealCorrectAnswers() bool
	GetAllowRetakes() bool
	GetCompletionPoints() int32
	GetPublishedAt() pgtype.Timestamptz
	GetEndTime() pgtype.Timestamptz
}

// Adapter for sqlc.CreateQuizRow
type createQuizRowAdapter struct {
	*sqlc.CreateQuizRow
}

func (r createQuizRowAdapter) GetID() string                        { return r.ID }
func (r createQuizRowAdapter) GetProjectID() string                 { return r.ProjectID }
func (r createQuizRowAdapter) GetChallengeID() string               { return r.ChallengeID }
func (r createQuizRowAdapter) GetName() string                      { return r.Name }
func (r createQuizRowAdapter) GetDescription() string               { return r.Description }
func (r createQuizRowAdapter) GetImageUrl() *string                 { return r.ImageUrl }
func (r createQuizRowAdapter) GetTimeoutSeconds() *int32            { return r.TimeoutSeconds }
func (r createQuizRowAdapter) GetQuestionTimeoutSeconds() *int32    { return r.QuestionTimeoutSeconds }
func (r createQuizRowAdapter) GetRandomizeQuestions() bool          { return r.RandomizeQuestions }
func (r createQuizRowAdapter) GetRevealCorrectAnswers() bool        { return r.RevealCorrectAnswers }
func (r createQuizRowAdapter) GetAllowRetakes() bool                { return r.AllowRetakes }
func (r createQuizRowAdapter) GetCompletionPoints() int32           { return r.CompletionPoints }
func (r createQuizRowAdapter) GetPublishedAt() pgtype.Timestamptz   { return r.PublishedAt }
func (r createQuizRowAdapter) GetEndTime() pgtype.Timestamptz       { return r.EndTime }

// Adapter for sqlc.UpdateQuizRow
type updateQuizRowAdapter struct {
	*sqlc.UpdateQuizRow
}

func (r updateQuizRowAdapter) GetID() string                        { return r.ID }
func (r updateQuizRowAdapter) GetProjectID() string                 { return r.ProjectID }
func (r updateQuizRowAdapter) GetChallengeID() string               { return r.ChallengeID }
func (r updateQuizRowAdapter) GetName() string                      { return r.Name }
func (r updateQuizRowAdapter) GetDescription() string               { return r.Description }
func (r updateQuizRowAdapter) GetImageUrl() *string                 { return r.ImageUrl }
func (r updateQuizRowAdapter) GetTimeoutSeconds() *int32            { return r.TimeoutSeconds }
func (r updateQuizRowAdapter) GetQuestionTimeoutSeconds() *int32    { return r.QuestionTimeoutSeconds }
func (r updateQuizRowAdapter) GetRandomizeQuestions() bool          { return r.RandomizeQuestions }
func (r updateQuizRowAdapter) GetRevealCorrectAnswers() bool        { return r.RevealCorrectAnswers }
func (r updateQuizRowAdapter) GetAllowRetakes() bool                { return r.AllowRetakes }
func (r updateQuizRowAdapter) GetCompletionPoints() int32           { return r.CompletionPoints }
func (r updateQuizRowAdapter) GetPublishedAt() pgtype.Timestamptz   { return r.PublishedAt }
func (r updateQuizRowAdapter) GetEndTime() pgtype.Timestamptz       { return r.EndTime }

// Adapter for sqlc.PublishQuizRow
type publishQuizRowAdapter struct {
	*sqlc.PublishQuizRow
}

func (r publishQuizRowAdapter) GetID() string                        { return r.ID }
func (r publishQuizRowAdapter) GetProjectID() string                 { return r.ProjectID }
func (r publishQuizRowAdapter) GetChallengeID() string               { return r.ChallengeID }
func (r publishQuizRowAdapter) GetName() string                      { return r.Name }
func (r publishQuizRowAdapter) GetDescription() string               { return r.Description }
func (r publishQuizRowAdapter) GetImageUrl() *string                 { return r.ImageUrl }
func (r publishQuizRowAdapter) GetTimeoutSeconds() *int32            { return r.TimeoutSeconds }
func (r publishQuizRowAdapter) GetQuestionTimeoutSeconds() *int32    { return r.QuestionTimeoutSeconds }
func (r publishQuizRowAdapter) GetRandomizeQuestions() bool          { return r.RandomizeQuestions }
func (r publishQuizRowAdapter) GetRevealCorrectAnswers() bool        { return r.RevealCorrectAnswers }
func (r publishQuizRowAdapter) GetAllowRetakes() bool                { return r.AllowRetakes }
func (r publishQuizRowAdapter) GetCompletionPoints() int32           { return r.CompletionPoints }
func (r publishQuizRowAdapter) GetPublishedAt() pgtype.Timestamptz   { return r.PublishedAt }
func (r publishQuizRowAdapter) GetEndTime() pgtype.Timestamptz       { return r.EndTime }

// Adapter for sqlc.GetQuizzesFilteredCursorRow
type filteredCursorQuizRowAdapter struct {
	*sqlc.GetQuizzesFilteredCursorRow
}

func (r filteredCursorQuizRowAdapter) GetID() string                        { return r.ID }
func (r filteredCursorQuizRowAdapter) GetProjectID() string                 { return r.ProjectID }
func (r filteredCursorQuizRowAdapter) GetChallengeID() string               { return r.ChallengeID }
func (r filteredCursorQuizRowAdapter) GetName() string                      { return r.Name }
func (r filteredCursorQuizRowAdapter) GetDescription() string               { return r.Description }
func (r filteredCursorQuizRowAdapter) GetImageUrl() *string                 { return r.ImageUrl }
func (r filteredCursorQuizRowAdapter) GetTimeoutSeconds() *int32            { return r.TimeoutSeconds }
func (r filteredCursorQuizRowAdapter) GetQuestionTimeoutSeconds() *int32    { return r.QuestionTimeoutSeconds }
func (r filteredCursorQuizRowAdapter) GetRandomizeQuestions() bool          { return r.RandomizeQuestions }
func (r filteredCursorQuizRowAdapter) GetRevealCorrectAnswers() bool        { return r.RevealCorrectAnswers }
func (r filteredCursorQuizRowAdapter) GetAllowRetakes() bool                { return r.AllowRetakes }
func (r filteredCursorQuizRowAdapter) GetCompletionPoints() int32           { return r.CompletionPoints }
func (r filteredCursorQuizRowAdapter) GetPublishedAt() pgtype.Timestamptz   { return r.PublishedAt }
func (r filteredCursorQuizRowAdapter) GetEndTime() pgtype.Timestamptz       { return r.EndTime }

func convertQuizRowToQuiz(row quizRow) *model.Quiz {
	var publishedAt *scalars.DateTime
	if row.GetPublishedAt().Valid {
		publishedAt = &scalars.DateTime{Time: row.GetPublishedAt().Time}
	}

	var endTime *scalars.DateTime
	if row.GetEndTime().Valid {
		endTime = &scalars.DateTime{Time: row.GetEndTime().Time}
	}

	var timeoutSeconds *int
	if row.GetTimeoutSeconds() != nil {
		ts := int(*row.GetTimeoutSeconds())
		timeoutSeconds = &ts
	}

	var questionTimeoutSeconds *int
	if row.GetQuestionTimeoutSeconds() != nil {
		qts := int(*row.GetQuestionTimeoutSeconds())
		questionTimeoutSeconds = &qts
	}

	return &model.Quiz{
		ID:                     row.GetID(),
		Name:                   row.GetName(),
		Description:            row.GetDescription(),
		Image:                  row.GetImageUrl(),
		ProjectID:              row.GetProjectID(),
		ChallengeID:            row.GetChallengeID(),
		TimeoutSeconds:         timeoutSeconds,
		QuestionTimeoutSeconds: questionTimeoutSeconds,
		RandomizeQuestions:     row.GetRandomizeQuestions(),
		RevealCorrectAnswers:   row.GetRevealCorrectAnswers(),
		AllowRetakes:           row.GetAllowRetakes(),
		CompletionPoints:       int(row.GetCompletionPoints()),
		PublishedAt:            publishedAt,
		EndTime:                endTime,
	}
}

func convertCreateQuizRowToQuiz(row *sqlc.CreateQuizRow) *model.Quiz {
	return convertQuizRowToQuiz(createQuizRowAdapter{row})
}

func convertUpdateQuizRowToQuiz(row *sqlc.UpdateQuizRow) *model.Quiz {
	return convertQuizRowToQuiz(updateQuizRowAdapter{row})
}

func convertPublishQuizRowToQuiz(row *sqlc.PublishQuizRow) *model.Quiz {
	return convertQuizRowToQuiz(publishQuizRowAdapter{row})
}

func convertFilteredCursorQuizRowToQuiz(row *sqlc.GetQuizzesFilteredCursorRow) *model.Quiz {
	return convertQuizRowToQuiz(filteredCursorQuizRowAdapter{row})
}

func convertCreateQuizQuestionRowToQuizQuestion(row *sqlc.QuizQuestion) *model.QuizQuestion {
	var allowMultipleSelection *bool
	if row.AllowMultipleSelection != nil {
		allowMultipleSelection = row.AllowMultipleSelection
	}

	var minValue, maxValue, stepValue *float64
	if row.MinValue.Valid {
		val, _ := row.MinValue.Float64Value()
		fv := val.Float64
		minValue = &fv
	}
	if row.MaxValue.Valid {
		val, _ := row.MaxValue.Float64Value()
		fv := val.Float64
		maxValue = &fv
	}
	if row.StepValue.Valid {
		val, _ := row.StepValue.Float64Value()
		fv := val.Float64
		stepValue = &fv
	}

	return &model.QuizQuestion{
		ID:                     row.ID,
		QuizID:                 row.QuizID,
		QuestionType:           model.QuizQuestionType(row.QuestionType),
		QuestionText:           row.QuestionText,
		QuestionOrder:          int(row.QuestionOrder),
		AllowMultipleSelection: allowMultipleSelection,
		MinValue:               minValue,
		MaxValue:               maxValue,
		StepValue:              stepValue,
	}
}

func convertSubmissionRowToModel(row *sqlc.QuizSubmission) *model.QuizSubmission {
	var completedAt *scalars.DateTime
	if row.CompletedAt.Valid {
		completedAt = &scalars.DateTime{Time: row.CompletedAt.Time}
	}

	var expiresAt *scalars.DateTime
	if row.ExpiresAt.Valid {
		expiresAt = &scalars.DateTime{Time: row.ExpiresAt.Time}
	}

	var score, maxScore, pointsAwarded *int
	if row.Score != nil {
		s := int(*row.Score)
		score = &s
	}
	if row.MaxScore != nil {
		ms := int(*row.MaxScore)
		maxScore = &ms
	}
	if row.PointsAwarded != nil {
		pa := int(*row.PointsAwarded)
		pointsAwarded = &pa
	}

	// Parse question order JSON
	var questionOrder []string
	if err := json.Unmarshal(row.QuestionOrder, &questionOrder); err != nil {
		// Return empty array on error
		questionOrder = []string{}
	}

	return &model.QuizSubmission{
		ID:            row.ID,
		QuizID:        row.QuizID,
		UserID:        row.UserID,
		StartedAt:     scalars.DateTime{Time: row.StartedAt.Time},
		CompletedAt:   completedAt,
		ExpiresAt:     expiresAt,
		QuestionOrder: questionOrder,
		Score:         score,
		MaxScore:      maxScore,
		PointsAwarded: pointsAwarded,
	}
}

func convertResponseRowToModel(row *sqlc.QuizResponse) *model.QuizResponse {
	var selectedAnswerIDs []string
	if row.SelectedAnswerIds != nil {
		_ = json.Unmarshal(row.SelectedAnswerIds, &selectedAnswerIDs)
	}

	var numberResponse *float64
	if row.NumberResponse.Valid {
		val, _ := row.NumberResponse.Float64Value()
		nr := val.Float64
		numberResponse = &nr
	}

	var jsonResponse *string
	if row.JsonResponse != nil {
		jsonStr := string(row.JsonResponse)
		jsonResponse = &jsonStr
	}

	var answeredAt *scalars.DateTime
	if row.AnsweredAt.Valid {
		answeredAt = &scalars.DateTime{Time: row.AnsweredAt.Time}
	}

	var timeSpentSeconds *int
	if row.TimeSpentSeconds != nil {
		tss := int(*row.TimeSpentSeconds)
		timeSpentSeconds = &tss
	}

	return &model.QuizResponse{
		ID:                row.ID,
		SubmissionID:      row.SubmissionID,
		QuestionID:        row.QuestionID,
		SelectedAnswerIds: selectedAnswerIDs,
		TextResponse:      row.TextResponse,
		NumberResponse:    numberResponse,
		JSONResponse:      jsonResponse,
		IsCorrect:         row.IsCorrect,
		AnsweredAt:        answeredAt,
		TimeSpentSeconds:  timeSpentSeconds,
	}
}
