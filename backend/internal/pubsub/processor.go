package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	// BatchSize is the number of users to process in a single batch
	BatchSize = 100
)

// BulkOperationProcessor is the interface for processing bulk operations
type BulkOperationProcessor interface {
	ProcessBulkEnrollChallenge(ctx context.Context, jobID string, params BulkEnrollChallengeParams) error
	ProcessBulkUnenrollChallenge(ctx context.Context, jobID string, params BulkUnenrollChallengeParams) error
	ProcessBulkCompleteChallenge(ctx context.Context, jobID string, params BulkCompleteChallengeParams) error
	ProcessBulkPublishChallenge(ctx context.Context, jobID string, params BulkPublishChallengeParams) error
	ProcessBulkAwardAchievement(ctx context.Context, jobID string, params BulkAwardAchievementParams) error
	ProcessBulkGrantQuizSessionAccess(ctx context.Context, jobID string, params BulkGrantQuizSessionAccessParams) error
	ProcessBulkScoreAdjustment(ctx context.Context, jobID string, params BulkScoreAdjustmentParams) error
	ProcessFixMissingContentProgress(ctx context.Context, jobID string, params FixMissingContentProgressParams) error
}

// Processor handles processing of bulk operation messages
type Processor struct {
	db                  *database.DB
	operationProcessors BulkOperationProcessor
	logger              *slog.Logger
}

// NewProcessor creates a new bulk operation processor
func NewProcessor(db *database.DB, operationProcessors BulkOperationProcessor, logger *slog.Logger) *Processor {
	return &Processor{
		db:                  db,
		operationProcessors: operationProcessors,
		logger:              logger,
	}
}

// Process processes a bulk operation message
func (p *Processor) Process(ctx context.Context, msg BulkOperationMessage) error {
	// Mark job as processing
	_, err := p.db.Queries.MarkBulkJobProcessing(ctx, msg.JobID)
	if err != nil {
		return fmt.Errorf("failed to mark job as processing: %w", err)
	}

	// Process based on operation type
	var processErr error
	switch msg.OperationType {
	case OperationBulkEnrollChallenge:
		var params BulkEnrollChallengeParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			processErr = fmt.Errorf("failed to unmarshal params: %w", err)
		} else {
			processErr = p.operationProcessors.ProcessBulkEnrollChallenge(ctx, msg.JobID, params)
		}

	case OperationBulkUnenrollChallenge:
		var params BulkUnenrollChallengeParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			processErr = fmt.Errorf("failed to unmarshal params: %w", err)
		} else {
			processErr = p.operationProcessors.ProcessBulkUnenrollChallenge(ctx, msg.JobID, params)
		}

	case OperationBulkCompleteChallenge:
		var params BulkCompleteChallengeParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			processErr = fmt.Errorf("failed to unmarshal params: %w", err)
		} else {
			processErr = p.operationProcessors.ProcessBulkCompleteChallenge(ctx, msg.JobID, params)
		}

	case OperationBulkPublishChallenge:
		var params BulkPublishChallengeParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			processErr = fmt.Errorf("failed to unmarshal params: %w", err)
		} else {
			processErr = p.operationProcessors.ProcessBulkPublishChallenge(ctx, msg.JobID, params)
		}

	case OperationBulkAwardAchievement:
		var params BulkAwardAchievementParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			processErr = fmt.Errorf("failed to unmarshal params: %w", err)
		} else {
			processErr = p.operationProcessors.ProcessBulkAwardAchievement(ctx, msg.JobID, params)
		}

	case OperationBulkGrantQuizSessionAccess:
		var params BulkGrantQuizSessionAccessParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			processErr = fmt.Errorf("failed to unmarshal params: %w", err)
		} else {
			processErr = p.operationProcessors.ProcessBulkGrantQuizSessionAccess(ctx, msg.JobID, params)
		}

	case OperationBulkScoreAdjustment:
		var params BulkScoreAdjustmentParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			processErr = fmt.Errorf("failed to unmarshal params: %w", err)
		} else {
			processErr = p.operationProcessors.ProcessBulkScoreAdjustment(ctx, msg.JobID, params)
		}

	case OperationFixMissingContentProgress:
		var params FixMissingContentProgressParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			processErr = fmt.Errorf("failed to unmarshal params: %w", err)
		} else {
			processErr = p.operationProcessors.ProcessFixMissingContentProgress(ctx, msg.JobID, params)
		}

	default:
		processErr = fmt.Errorf("unknown operation type: %s", msg.OperationType)
	}

	// Mark job as completed or failed
	if processErr != nil {
		_, markErr := p.db.Queries.MarkBulkJobFailed(ctx, sqlc.MarkBulkJobFailedParams{
			ID:           msg.JobID,
			Errormessage: processErr.Error(),
			Errordetails: pgtype.FlatArray[byte](nil),
		})
		if markErr != nil {
			p.logger.Error("Failed to mark job as failed",
				"job_id", msg.JobID,
				"mark_error", markErr,
				"process_error", processErr,
			)
		}
		return processErr
	}

	return nil
}

// UpdateJobProgress updates the job progress in the database
func (p *Processor) UpdateJobProgress(ctx context.Context, jobID string, processed, success, failure int) error {
	_, err := p.db.Queries.UpdateBulkJobProgress(ctx, sqlc.UpdateBulkJobProgressParams{
		ID:             jobID,
		Processedcount: int32(processed),
		Successcount:   int32(success),
		Failurecount:   int32(failure),
	})
	return err
}

// MarkJobCompleted marks a job as completed with final counts
func (p *Processor) MarkJobCompleted(ctx context.Context, jobID string, processed, success, failure int) error {
	_, err := p.db.Queries.MarkBulkJobCompleted(ctx, sqlc.MarkBulkJobCompletedParams{
		ID:             jobID,
		Processedcount: int32(processed),
		Successcount:   int32(success),
		Failurecount:   int32(failure),
	})
	return err
}
