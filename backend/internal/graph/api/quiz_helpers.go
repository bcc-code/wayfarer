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
	GetRandomizeQuestions() bool
	GetRevealCorrectAnswers() bool
	GetAllowRetakes() bool
	GetCompletionPoints() int32
	GetEndTime() pgtype.Timestamptz
}

// Adapter for sqlc.CreateQuizRow
type createQuizRowAdapter struct {
	*sqlc.CreateQuizRow
}

func (r createQuizRowAdapter) GetID() string                  { return r.ID }
func (r createQuizRowAdapter) GetProjectID() string           { return r.ProjectID }
func (r createQuizRowAdapter) GetChallengeID() string         { return r.ChallengeID }
func (r createQuizRowAdapter) GetName() string                { return r.Name }
func (r createQuizRowAdapter) GetDescription() string         { return r.Description }
func (r createQuizRowAdapter) GetImageUrl() *string           { return r.ImageUrl }
func (r createQuizRowAdapter) GetTimeoutSeconds() *int32      { return r.TimeoutSeconds }
func (r createQuizRowAdapter) GetRandomizeQuestions() bool    { return r.RandomizeQuestions }
func (r createQuizRowAdapter) GetRevealCorrectAnswers() bool  { return r.RevealCorrectAnswers }
func (r createQuizRowAdapter) GetAllowRetakes() bool          { return r.AllowRetakes }
func (r createQuizRowAdapter) GetCompletionPoints() int32     { return r.CompletionPoints }
func (r createQuizRowAdapter) GetEndTime() pgtype.Timestamptz { return r.EndTime }

// Adapter for sqlc.UpdateQuizRow
type updateQuizRowAdapter struct {
	*sqlc.UpdateQuizRow
}

func (r updateQuizRowAdapter) GetID() string                  { return r.ID }
func (r updateQuizRowAdapter) GetProjectID() string           { return r.ProjectID }
func (r updateQuizRowAdapter) GetChallengeID() string         { return r.ChallengeID }
func (r updateQuizRowAdapter) GetName() string                { return r.Name }
func (r updateQuizRowAdapter) GetDescription() string         { return r.Description }
func (r updateQuizRowAdapter) GetImageUrl() *string           { return r.ImageUrl }
func (r updateQuizRowAdapter) GetTimeoutSeconds() *int32      { return r.TimeoutSeconds }
func (r updateQuizRowAdapter) GetRandomizeQuestions() bool    { return r.RandomizeQuestions }
func (r updateQuizRowAdapter) GetRevealCorrectAnswers() bool  { return r.RevealCorrectAnswers }
func (r updateQuizRowAdapter) GetAllowRetakes() bool          { return r.AllowRetakes }
func (r updateQuizRowAdapter) GetCompletionPoints() int32     { return r.CompletionPoints }
func (r updateQuizRowAdapter) GetEndTime() pgtype.Timestamptz { return r.EndTime }

// Adapter for sqlc.GetQuizzesFilteredCursorRow
type filteredCursorQuizRowAdapter struct {
	*sqlc.GetQuizzesFilteredCursorRow
}

func (r filteredCursorQuizRowAdapter) GetID() string                  { return r.ID }
func (r filteredCursorQuizRowAdapter) GetProjectID() string           { return r.ProjectID }
func (r filteredCursorQuizRowAdapter) GetChallengeID() string         { return r.ChallengeID }
func (r filteredCursorQuizRowAdapter) GetName() string                { return r.Name }
func (r filteredCursorQuizRowAdapter) GetDescription() string         { return r.Description }
func (r filteredCursorQuizRowAdapter) GetImageUrl() *string           { return r.ImageUrl }
func (r filteredCursorQuizRowAdapter) GetTimeoutSeconds() *int32      { return r.TimeoutSeconds }
func (r filteredCursorQuizRowAdapter) GetRandomizeQuestions() bool    { return r.RandomizeQuestions }
func (r filteredCursorQuizRowAdapter) GetRevealCorrectAnswers() bool  { return r.RevealCorrectAnswers }
func (r filteredCursorQuizRowAdapter) GetAllowRetakes() bool          { return r.AllowRetakes }
func (r filteredCursorQuizRowAdapter) GetCompletionPoints() int32     { return r.CompletionPoints }
func (r filteredCursorQuizRowAdapter) GetEndTime() pgtype.Timestamptz { return r.EndTime }

func convertQuizRowToQuiz(row quizRow) *model.Quiz {
	var endTime *scalars.DateTime
	if row.GetEndTime().Valid {
		endTime = &scalars.DateTime{Time: row.GetEndTime().Time}
	}

	var timeoutSeconds *int
	if row.GetTimeoutSeconds() != nil {
		ts := int(*row.GetTimeoutSeconds())
		timeoutSeconds = &ts
	}

	return &model.Quiz{
		ID:                   row.GetID(),
		Name:                 row.GetName(),
		Description:          row.GetDescription(),
		Image:                row.GetImageUrl(),
		ProjectID:            row.GetProjectID(),
		ChallengeID:          row.GetChallengeID(),
		TimeoutSeconds:       timeoutSeconds,
		RandomizeQuestions:   row.GetRandomizeQuestions(),
		RevealCorrectAnswers: row.GetRevealCorrectAnswers(),
		AllowRetakes:         row.GetAllowRetakes(),
		CompletionPoints:     int(row.GetCompletionPoints()),
		EndTime:              endTime,
	}
}

func convertCreateQuizRowToQuiz(row *sqlc.CreateQuizRow) *model.Quiz {
	return convertQuizRowToQuiz(createQuizRowAdapter{row})
}

func convertUpdateQuizRowToQuiz(row *sqlc.UpdateQuizRow) *model.Quiz {
	return convertQuizRowToQuiz(updateQuizRowAdapter{row})
}

func convertFilteredCursorQuizRowToQuiz(row *sqlc.GetQuizzesFilteredCursorRow) *model.Quiz {
	return convertQuizRowToQuiz(filteredCursorQuizRowAdapter{row})
}

// quizQuestionRow defines the common interface for quiz question row types
type quizQuestionRow interface {
	GetID() string
	GetQuizID() string
	GetQuestionType() string
	GetQuestionText() string
	GetQuestionOrder() int32
	GetAllowMultipleSelection() *bool
	GetMinValue() pgtype.Numeric
	GetMaxValue() pgtype.Numeric
	GetStepValue() pgtype.Numeric
	GetTimeoutSeconds() *int32
	GetPoints() *int32
}

// Adapter for sqlc.CreateQuizQuestionRow
type createQuizQuestionRowAdapter struct {
	*sqlc.CreateQuizQuestionRow
}

func (r createQuizQuestionRowAdapter) GetID() string           { return r.ID }
func (r createQuizQuestionRowAdapter) GetQuizID() string       { return r.QuizID }
func (r createQuizQuestionRowAdapter) GetQuestionType() string { return r.QuestionType }
func (r createQuizQuestionRowAdapter) GetQuestionText() string { return r.QuestionText }
func (r createQuizQuestionRowAdapter) GetQuestionOrder() int32 { return r.QuestionOrder }
func (r createQuizQuestionRowAdapter) GetAllowMultipleSelection() *bool {
	return r.AllowMultipleSelection
}
func (r createQuizQuestionRowAdapter) GetMinValue() pgtype.Numeric  { return r.MinValue }
func (r createQuizQuestionRowAdapter) GetMaxValue() pgtype.Numeric  { return r.MaxValue }
func (r createQuizQuestionRowAdapter) GetStepValue() pgtype.Numeric { return r.StepValue }
func (r createQuizQuestionRowAdapter) GetTimeoutSeconds() *int32    { return r.TimeoutSeconds }
func (r createQuizQuestionRowAdapter) GetPoints() *int32            { return r.Points }

// Adapter for sqlc.UpdateQuizQuestionRow
type updateQuizQuestionRowAdapter struct {
	*sqlc.UpdateQuizQuestionRow
}

func (r updateQuizQuestionRowAdapter) GetID() string           { return r.ID }
func (r updateQuizQuestionRowAdapter) GetQuizID() string       { return r.QuizID }
func (r updateQuizQuestionRowAdapter) GetQuestionType() string { return r.QuestionType }
func (r updateQuizQuestionRowAdapter) GetQuestionText() string { return r.QuestionText }
func (r updateQuizQuestionRowAdapter) GetQuestionOrder() int32 { return r.QuestionOrder }
func (r updateQuizQuestionRowAdapter) GetAllowMultipleSelection() *bool {
	return r.AllowMultipleSelection
}
func (r updateQuizQuestionRowAdapter) GetMinValue() pgtype.Numeric  { return r.MinValue }
func (r updateQuizQuestionRowAdapter) GetMaxValue() pgtype.Numeric  { return r.MaxValue }
func (r updateQuizQuestionRowAdapter) GetStepValue() pgtype.Numeric { return r.StepValue }
func (r updateQuizQuestionRowAdapter) GetTimeoutSeconds() *int32    { return r.TimeoutSeconds }
func (r updateQuizQuestionRowAdapter) GetPoints() *int32            { return r.Points }

// Adapter for sqlc.GetQuizQuestionByIDRow
type getQuizQuestionByIDRowAdapter struct {
	*sqlc.GetQuizQuestionByIDRow
}

func (r getQuizQuestionByIDRowAdapter) GetID() string           { return r.ID }
func (r getQuizQuestionByIDRowAdapter) GetQuizID() string       { return r.QuizID }
func (r getQuizQuestionByIDRowAdapter) GetQuestionType() string { return r.QuestionType }
func (r getQuizQuestionByIDRowAdapter) GetQuestionText() string { return r.QuestionText }
func (r getQuizQuestionByIDRowAdapter) GetQuestionOrder() int32 { return r.QuestionOrder }
func (r getQuizQuestionByIDRowAdapter) GetAllowMultipleSelection() *bool {
	return r.AllowMultipleSelection
}
func (r getQuizQuestionByIDRowAdapter) GetMinValue() pgtype.Numeric  { return r.MinValue }
func (r getQuizQuestionByIDRowAdapter) GetMaxValue() pgtype.Numeric  { return r.MaxValue }
func (r getQuizQuestionByIDRowAdapter) GetStepValue() pgtype.Numeric { return r.StepValue }
func (r getQuizQuestionByIDRowAdapter) GetTimeoutSeconds() *int32    { return r.TimeoutSeconds }
func (r getQuizQuestionByIDRowAdapter) GetPoints() *int32            { return r.Points }

// convertQuizQuestionRowToInterface converts a database row to the appropriate QuizQuestion implementation
func convertQuizQuestionRowToInterface(row quizQuestionRow) model.QuizQuestion {
	switch row.GetQuestionType() {
	case "PREDEFINED":
		return convertToPredefinedQuestion(row)
	case "FREE_TEXT":
		return convertToFreeTextQuestion(row)
	case "NUMBER":
		return convertToNumberQuestion(row)
	case "JSON":
		return convertToJsonQuestion(row)
	case "ORDERING":
		return convertToOrderingQuestion(row)
	default:
		// Default to FreeTextQuestion for unknown types
		return convertToFreeTextQuestion(row)
	}
}

func convertQuestionTimeoutSeconds(ts *int32) *int {
	if ts == nil {
		return nil
	}
	v := int(*ts)
	return &v
}

func convertQuestionPoints(pts *int32) *int {
	if pts == nil {
		return nil
	}
	v := int(*pts)
	return &v
}

func convertToPredefinedQuestion(row quizQuestionRow) *model.PredefinedQuestion {
	allowMultiple := false
	if row.GetAllowMultipleSelection() != nil {
		allowMultiple = *row.GetAllowMultipleSelection()
	}
	return &model.PredefinedQuestion{
		ID:                     row.GetID(),
		QuizID:                 row.GetQuizID(),
		QuestionText:           row.GetQuestionText(),
		QuestionOrder:          int(row.GetQuestionOrder()),
		TimeoutSeconds:         convertQuestionTimeoutSeconds(row.GetTimeoutSeconds()),
		Points:                 convertQuestionPoints(row.GetPoints()),
		AllowMultipleSelection: allowMultiple,
	}
}

func convertToFreeTextQuestion(row quizQuestionRow) *model.FreeTextQuestion {
	return &model.FreeTextQuestion{
		ID:             row.GetID(),
		QuizID:         row.GetQuizID(),
		QuestionText:   row.GetQuestionText(),
		QuestionOrder:  int(row.GetQuestionOrder()),
		TimeoutSeconds: convertQuestionTimeoutSeconds(row.GetTimeoutSeconds()),
		Points:         convertQuestionPoints(row.GetPoints()),
	}
}

func convertToNumberQuestion(row quizQuestionRow) *model.NumberQuestion {
	var minValue, maxValue, stepValue *float64
	if row.GetMinValue().Valid {
		val, _ := row.GetMinValue().Float64Value()
		fv := val.Float64
		minValue = &fv
	}
	if row.GetMaxValue().Valid {
		val, _ := row.GetMaxValue().Float64Value()
		fv := val.Float64
		maxValue = &fv
	}
	if row.GetStepValue().Valid {
		val, _ := row.GetStepValue().Float64Value()
		fv := val.Float64
		stepValue = &fv
	}
	return &model.NumberQuestion{
		ID:             row.GetID(),
		QuizID:         row.GetQuizID(),
		QuestionText:   row.GetQuestionText(),
		QuestionOrder:  int(row.GetQuestionOrder()),
		TimeoutSeconds: convertQuestionTimeoutSeconds(row.GetTimeoutSeconds()),
		Points:         convertQuestionPoints(row.GetPoints()),
		MinValue:       minValue,
		MaxValue:       maxValue,
		StepValue:      stepValue,
	}
}

func convertToJsonQuestion(row quizQuestionRow) *model.JSONQuestion {
	return &model.JSONQuestion{
		ID:             row.GetID(),
		QuizID:         row.GetQuizID(),
		QuestionText:   row.GetQuestionText(),
		QuestionOrder:  int(row.GetQuestionOrder()),
		TimeoutSeconds: convertQuestionTimeoutSeconds(row.GetTimeoutSeconds()),
		Points:         convertQuestionPoints(row.GetPoints()),
	}
}

func convertToOrderingQuestion(row quizQuestionRow) *model.OrderingQuestion {
	return &model.OrderingQuestion{
		ID:             row.GetID(),
		QuizID:         row.GetQuizID(),
		QuestionText:   row.GetQuestionText(),
		QuestionOrder:  int(row.GetQuestionOrder()),
		TimeoutSeconds: convertQuestionTimeoutSeconds(row.GetTimeoutSeconds()),
		Points:         convertQuestionPoints(row.GetPoints()),
	}
}

// Helper functions to convert specific row types
func convertCreateQuizQuestionRowToInterface(row *sqlc.CreateQuizQuestionRow) model.QuizQuestion {
	return convertQuizQuestionRowToInterface(createQuizQuestionRowAdapter{row})
}

func convertUpdateQuizQuestionRowToInterface(row *sqlc.UpdateQuizQuestionRow) model.QuizQuestion {
	return convertQuizQuestionRowToInterface(updateQuizQuestionRowAdapter{row})
}

func convertGetQuizQuestionByIDRowToInterface(row *sqlc.GetQuizQuestionByIDRow) model.QuizQuestion {
	return convertQuizQuestionRowToInterface(getQuizQuestionByIDRowAdapter{row})
}

// quizSubmissionRow defines the common fields for quiz submission row types
type quizSubmissionRow struct {
	ID            string
	QuizID        string
	UserID        string
	SessionID     *string
	StartedAt     pgtype.Timestamptz
	CompletedAt   pgtype.Timestamptz
	ExpiresAt     pgtype.Timestamptz
	QuestionOrder []byte
	Score         *int32
	MaxScore      *int32
	PointsAwarded *int32
	AutoSubmitted bool
}

func convertSubmissionRowToModel(row quizSubmissionRow) *model.QuizSubmission {
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
		SessionID:     row.SessionID,
		StartedAt:     scalars.DateTime{Time: row.StartedAt.Time},
		CompletedAt:   completedAt,
		ExpiresAt:     expiresAt,
		QuestionOrder: questionOrder,
		Score:         score,
		MaxScore:      maxScore,
		PointsAwarded: pointsAwarded,
		AutoSubmitted: row.AutoSubmitted,
	}
}

// Wrapper functions for each submission row type

func convertCreateQuizSubmissionRow(row *sqlc.CreateQuizSubmissionRow) *model.QuizSubmission {
	return convertSubmissionRowToModel(quizSubmissionRow{
		ID: row.ID, QuizID: row.QuizID, UserID: row.UserID, SessionID: row.SessionID,
		StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, ExpiresAt: row.ExpiresAt,
		QuestionOrder: row.QuestionOrder, Score: row.Score, MaxScore: row.MaxScore,
		PointsAwarded: row.PointsAwarded, AutoSubmitted: row.AutoSubmitted,
	})
}

func convertUpdateQuizSubmissionRow(row *sqlc.UpdateQuizSubmissionRow) *model.QuizSubmission {
	return convertSubmissionRowToModel(quizSubmissionRow{
		ID: row.ID, QuizID: row.QuizID, UserID: row.UserID, SessionID: row.SessionID,
		StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, ExpiresAt: row.ExpiresAt,
		QuestionOrder: row.QuestionOrder, Score: row.Score, MaxScore: row.MaxScore,
		PointsAwarded: row.PointsAwarded, AutoSubmitted: row.AutoSubmitted,
	})
}

func convertGetQuizSubmissionByIDRow(row *sqlc.GetQuizSubmissionByIDRow) *model.QuizSubmission {
	return convertSubmissionRowToModel(quizSubmissionRow{
		ID: row.ID, QuizID: row.QuizID, UserID: row.UserID, SessionID: row.SessionID,
		StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, ExpiresAt: row.ExpiresAt,
		QuestionOrder: row.QuestionOrder, Score: row.Score, MaxScore: row.MaxScore,
		PointsAwarded: row.PointsAwarded, AutoSubmitted: row.AutoSubmitted,
	})
}

func convertGetQuizSubmissionsFilteredCursorRow(row *sqlc.GetQuizSubmissionsFilteredCursorRow) *model.QuizSubmission {
	return convertSubmissionRowToModel(quizSubmissionRow{
		ID: row.ID, QuizID: row.QuizID, UserID: row.UserID, SessionID: row.SessionID,
		StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, ExpiresAt: row.ExpiresAt,
		QuestionOrder: row.QuestionOrder, Score: row.Score, MaxScore: row.MaxScore,
		PointsAwarded: row.PointsAwarded, AutoSubmitted: row.AutoSubmitted,
	})
}

// convertResponseRowToInterface converts a database row to the appropriate QuizResponse implementation
// It determines the type based on which response field is populated
func convertResponseRowToInterface(row *sqlc.QuizResponse, questionType string) model.QuizResponse {
	var answeredAt *scalars.DateTime
	if row.AnsweredAt.Valid {
		answeredAt = &scalars.DateTime{Time: row.AnsweredAt.Time}
	}

	var timeSpentSeconds *int
	if row.TimeSpentSeconds != nil {
		tss := int(*row.TimeSpentSeconds)
		timeSpentSeconds = &tss
	}

	var pointsEarned *int
	if row.PointsEarned != nil {
		pe := int(*row.PointsEarned)
		pointsEarned = &pe
	}

	switch questionType {
	case "PREDEFINED":
		return convertToPredefinedResponse(row, answeredAt, timeSpentSeconds, pointsEarned)
	case "FREE_TEXT":
		return convertToFreeTextResponse(row, answeredAt, timeSpentSeconds, pointsEarned)
	case "NUMBER":
		return convertToNumberResponse(row, answeredAt, timeSpentSeconds, pointsEarned)
	case "JSON":
		return convertToJsonResponse(row, answeredAt, timeSpentSeconds, pointsEarned)
	case "ORDERING":
		return convertToOrderingResponse(row, answeredAt, timeSpentSeconds, pointsEarned)
	default:
		// Default to FreeTextResponse for unknown types
		return convertToFreeTextResponse(row, answeredAt, timeSpentSeconds, pointsEarned)
	}
}

func convertToPredefinedResponse(row *sqlc.QuizResponse, answeredAt *scalars.DateTime, timeSpentSeconds *int, pointsEarned *int) *model.PredefinedResponse {
	var selectedAnswerIDs []string
	if row.SelectedAnswerIds != nil {
		_ = json.Unmarshal(row.SelectedAnswerIds, &selectedAnswerIDs)
	}
	return &model.PredefinedResponse{
		ID:                row.ID,
		SubmissionID:      row.SubmissionID,
		QuestionID:        row.QuestionID,
		SelectedAnswerIds: selectedAnswerIDs,
		IsCorrect:         row.IsCorrect,
		AnsweredAt:        answeredAt,
		TimeSpentSeconds:  timeSpentSeconds,
		PointsEarned:      pointsEarned,
	}
}

func convertToFreeTextResponse(row *sqlc.QuizResponse, answeredAt *scalars.DateTime, timeSpentSeconds *int, pointsEarned *int) *model.FreeTextResponse {
	textResponse := ""
	if row.TextResponse != nil {
		textResponse = *row.TextResponse
	}
	return &model.FreeTextResponse{
		ID:               row.ID,
		SubmissionID:     row.SubmissionID,
		QuestionID:       row.QuestionID,
		TextResponse:     textResponse,
		AnsweredAt:       answeredAt,
		TimeSpentSeconds: timeSpentSeconds,
		PointsEarned:     pointsEarned,
	}
}

func convertToNumberResponse(row *sqlc.QuizResponse, answeredAt *scalars.DateTime, timeSpentSeconds *int, pointsEarned *int) *model.NumberResponse {
	var numberResponse float64
	if row.NumberResponse.Valid {
		val, _ := row.NumberResponse.Float64Value()
		numberResponse = val.Float64
	}
	return &model.NumberResponse{
		ID:               row.ID,
		SubmissionID:     row.SubmissionID,
		QuestionID:       row.QuestionID,
		NumberResponse:   numberResponse,
		AnsweredAt:       answeredAt,
		TimeSpentSeconds: timeSpentSeconds,
		PointsEarned:     pointsEarned,
	}
}

func convertToJsonResponse(row *sqlc.QuizResponse, answeredAt *scalars.DateTime, timeSpentSeconds *int, pointsEarned *int) *model.JSONResponse {
	jsonResponse := ""
	if row.JsonResponse != nil {
		jsonResponse = string(row.JsonResponse)
	}
	return &model.JSONResponse{
		ID:               row.ID,
		SubmissionID:     row.SubmissionID,
		QuestionID:       row.QuestionID,
		JSONResponse:     jsonResponse,
		AnsweredAt:       answeredAt,
		TimeSpentSeconds: timeSpentSeconds,
		PointsEarned:     pointsEarned,
	}
}

func convertToOrderingResponse(row *sqlc.QuizResponse, answeredAt *scalars.DateTime, timeSpentSeconds *int, pointsEarned *int) *model.OrderingResponse {
	var submittedOrder []string
	if row.JsonResponse != nil {
		_ = json.Unmarshal(row.JsonResponse, &submittedOrder)
	}
	return &model.OrderingResponse{
		ID:               row.ID,
		SubmissionID:     row.SubmissionID,
		QuestionID:       row.QuestionID,
		SubmittedOrder:   submittedOrder,
		IsCorrectValue:   row.IsCorrect, // Store actual value for resolver
		AnsweredAt:       answeredAt,
		TimeSpentSeconds: timeSpentSeconds,
		PointsEarned:     pointsEarned,
	}
}

func convertGetSubmissionByUserAndSessionRow(row *sqlc.GetSubmissionByUserAndSessionRow) *model.QuizSubmission {
	return convertSubmissionRowToModel(quizSubmissionRow{
		ID: row.ID, QuizID: row.QuizID, UserID: row.UserID, SessionID: row.SessionID,
		StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, ExpiresAt: row.ExpiresAt,
		QuestionOrder: row.QuestionOrder, Score: row.Score, MaxScore: row.MaxScore,
		PointsAwarded: row.PointsAwarded, AutoSubmitted: row.AutoSubmitted,
	})
}

// convertQuizSessionToModel converts a sqlc.QuizSession to model.QuizSession
func convertQuizSessionToModel(row *sqlc.QuizSession) *model.QuizSession {
	var openAt, lockAt, finishAt *scalars.DateTime
	if row.OpenAt.Valid {
		openAt = &scalars.DateTime{Time: row.OpenAt.Time}
	}
	if row.LockAt.Valid {
		lockAt = &scalars.DateTime{Time: row.LockAt.Time}
	}
	if row.FinishAt.Valid {
		finishAt = &scalars.DateTime{Time: row.FinishAt.Time}
	}

	return &model.QuizSession{
		ID:          row.ID,
		QuizID:      row.QuizID,
		Name:        row.Name,
		State:       model.QuizSessionState(row.State),
		OpenAt:      openAt,
		LockAt:      lockAt,
		FinishAt:    finishAt,
		CreatedByID: row.CreatedBy,
		CreatedAt:   scalars.DateTime{Time: row.CreatedAt.Time},
	}
}
