package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	pgx "github.com/jackc/pgx/v5"
)

// Answers are written while holding the submission's row lock. Finishing a
// session waits for the locks on all its submissions (completed ones included,
// see closeSessionAndSettleBets) before it loads the bets to settle, so an
// answer written under the lock is either committed in time to be settled or
// sees the session closed and is rejected. Without the lock, an in-flight
// answer could be stored after settlement and its bet would never be paid out.

// createQuizResponseLocked stores a new answer under the submission lock,
// re-checking that the submission is still open. If the question was already
// answered (e.g. a retried request), the stored response is returned instead.
func (r *Resolver) createQuizResponseLocked(ctx context.Context, userID string, params sqlc.CreateQuizResponseParams) (*sqlc.QuizResponse, error) {
	tx, err := r.DB.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.DB.Queries.WithTx(tx)

	submission, err := qtx.GetQuizSubmissionByIDForUpdate(ctx, params.Submissionid)
	if err != nil {
		return nil, fmt.Errorf("failed to load submission: %w", err)
	}
	if submission.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	if submission.CompletedAt.Valid {
		return nil, fmt.Errorf("submission already completed")
	}
	if submission.ExpiresAt.Valid && time.Now().After(submission.ExpiresAt.Time) {
		return nil, fmt.Errorf("submission expired")
	}

	existing, err := qtx.GetQuizResponseBySubmissionAndQuestion(ctx, sqlc.GetQuizResponseBySubmissionAndQuestionParams{
		Submissionid: params.Submissionid,
		Questionid:   params.Questionid,
	})
	if err == nil {
		return &sqlc.QuizResponse{
			ID:                existing.ID,
			SubmissionID:      existing.SubmissionID,
			QuestionID:        existing.QuestionID,
			SelectedAnswerIds: existing.SelectedAnswerIds,
			TextResponse:      existing.TextResponse,
			NumberResponse:    existing.NumberResponse,
			JsonResponse:      existing.JsonResponse,
			IsCorrect:         existing.IsCorrect,
			PointsEarned:      existing.PointsEarned,
			AnsweredAt:        existing.AnsweredAt,
			TimeSpentSeconds:  existing.TimeSpentSeconds,
			BetAmount:         existing.BetAmount,
		}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing response: %w", err)
	}

	response, err := qtx.CreateQuizResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to save response: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &sqlc.QuizResponse{
		ID:                response.ID,
		SubmissionID:      response.SubmissionID,
		QuestionID:        response.QuestionID,
		SelectedAnswerIds: response.SelectedAnswerIds,
		TextResponse:      response.TextResponse,
		NumberResponse:    response.NumberResponse,
		JsonResponse:      response.JsonResponse,
		IsCorrect:         response.IsCorrect,
		PointsEarned:      response.PointsEarned,
		AnsweredAt:        response.AnsweredAt,
		TimeSpentSeconds:  response.TimeSpentSeconds,
		BetAmount:         response.BetAmount,
	}, nil
}

// updateQuizResponseLocked updates an answer under the submission lock,
// re-checking that the submission's session (if any) is still OPEN.
func (r *Resolver) updateQuizResponseLocked(ctx context.Context, submissionID string, params sqlc.UpdateQuizResponseParams) (*sqlc.UpdateQuizResponseRow, error) {
	tx, err := r.DB.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.DB.Queries.WithTx(tx)

	submission, err := qtx.GetQuizSubmissionByIDForUpdate(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load submission: %w", err)
	}
	if submission.SessionID != nil && *submission.SessionID != "" {
		session, err := qtx.GetQuizSession(ctx, *submission.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to load session: %w", err)
		}
		if session.State != string(model.QuizSessionStateOpen) {
			return nil, fmt.Errorf("quiz session is %s, answers cannot be modified", session.State)
		}
	}

	updated, err := qtx.UpdateQuizResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update response: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return updated, nil
}
