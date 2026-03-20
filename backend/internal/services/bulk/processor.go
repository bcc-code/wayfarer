package bulk

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/pubsub"
)

// Processor implements pubsub.BulkOperationProcessor
type Processor struct {
	service *Service
}

// NewProcessor creates a new bulk operation processor
func NewProcessor(service *Service) *Processor {
	return &Processor{service: service}
}

// ProcessBulkEnrollChallenge processes bulk challenge enrollment
func (p *Processor) ProcessBulkEnrollChallenge(ctx context.Context, jobID string, params pubsub.BulkEnrollChallengeParams) error {
	successCount, failureCount, err := p.service.EnrollUsersInChallenge(ctx, params.ChallengeID, params.UserIDs, true)
	if err != nil {
		return err
	}

	// Mark job completed
	_, err = p.service.DB.Queries.MarkBulkJobCompleted(ctx, sqlc.MarkBulkJobCompletedParams{
		ID:             jobID,
		Processedcount: int32(successCount + failureCount),
		Successcount:   int32(successCount),
		Failurecount:   int32(failureCount),
	})
	return err
}

// ProcessBulkUnenrollChallenge processes bulk challenge unenrollment
func (p *Processor) ProcessBulkUnenrollChallenge(ctx context.Context, jobID string, params pubsub.BulkUnenrollChallengeParams) error {
	successCount, failureCount, err := p.service.UnenrollUsersFromChallenge(ctx, params.ChallengeID, params.UserIDs)
	if err != nil {
		return err
	}

	// Mark job completed
	_, err = p.service.DB.Queries.MarkBulkJobCompleted(ctx, sqlc.MarkBulkJobCompletedParams{
		ID:             jobID,
		Processedcount: int32(successCount + failureCount),
		Successcount:   int32(successCount),
		Failurecount:   int32(failureCount),
	})
	return err
}

// ProcessBulkCompleteChallenge processes bulk challenge completion
func (p *Processor) ProcessBulkCompleteChallenge(ctx context.Context, jobID string, params pubsub.BulkCompleteChallengeParams) error {
	var completedAt *time.Time
	if params.CompletedAt != "" {
		t, err := time.Parse(time.RFC3339, params.CompletedAt)
		if err != nil {
			return fmt.Errorf("invalid completed_at timestamp: %w", err)
		}
		completedAt = &t
	}

	successCount, failureCount, err := p.service.CompleteUsersChallenge(ctx, params.ChallengeID, params.UserIDs, completedAt)
	if err != nil {
		return err
	}

	// Mark job completed
	_, err = p.service.DB.Queries.MarkBulkJobCompleted(ctx, sqlc.MarkBulkJobCompletedParams{
		ID:             jobID,
		Processedcount: int32(successCount + failureCount),
		Successcount:   int32(successCount),
		Failurecount:   int32(failureCount),
	})
	return err
}

// ProcessBulkPublishChallenge processes bulk challenge publishing
func (p *Processor) ProcessBulkPublishChallenge(ctx context.Context, jobID string, params pubsub.BulkPublishChallengeParams) error {
	publishedAt, err := time.Parse(time.RFC3339, params.PublishedAt)
	if err != nil {
		return fmt.Errorf("invalid published_at timestamp: %w", err)
	}

	successCount, failureCount, err := p.service.PublishChallenges(ctx, params.ChallengeIDs, publishedAt)
	if err != nil {
		return err
	}

	// Mark job completed
	_, err = p.service.DB.Queries.MarkBulkJobCompleted(ctx, sqlc.MarkBulkJobCompletedParams{
		ID:             jobID,
		Processedcount: int32(successCount + failureCount),
		Successcount:   int32(successCount),
		Failurecount:   int32(failureCount),
	})
	return err
}

// ProcessBulkAwardAchievement processes bulk achievement awarding
func (p *Processor) ProcessBulkAwardAchievement(ctx context.Context, jobID string, params pubsub.BulkAwardAchievementParams) error {
	// If teamId is specified, resolve to user IDs
	userIDs := params.UserIDs
	if params.TeamID != "" {
		teamUserIDs, err := p.service.DB.Queries.GetUserIDsInTeams(ctx, []string{params.TeamID})
		if err != nil {
			return fmt.Errorf("failed to get team members: %w", err)
		}
		userIDs = teamUserIDs
	}

	successCount, failureCount, err := p.service.AwardAchievements(ctx, params.AchievementID, userIDs)
	if err != nil {
		return err
	}

	// Mark job completed
	_, err = p.service.DB.Queries.MarkBulkJobCompleted(ctx, sqlc.MarkBulkJobCompletedParams{
		ID:             jobID,
		Processedcount: int32(successCount + failureCount),
		Successcount:   int32(successCount),
		Failurecount:   int32(failureCount),
	})
	return err
}

// ProcessBulkGrantQuizSessionAccess processes bulk quiz session access granting
func (p *Processor) ProcessBulkGrantQuizSessionAccess(ctx context.Context, jobID string, params pubsub.BulkGrantQuizSessionAccessParams) error {
	successCount, failureCount, err := p.service.GrantQuizSessionAccess(ctx, params)
	if err != nil {
		return err
	}

	// Mark job completed
	_, err = p.service.DB.Queries.MarkBulkJobCompleted(ctx, sqlc.MarkBulkJobCompletedParams{
		ID:             jobID,
		Processedcount: int32(successCount + failureCount),
		Successcount:   int32(successCount),
		Failurecount:   int32(failureCount),
	})
	return err
}

// ProcessBulkScoreAdjustment processes bulk score adjustment
func (p *Processor) ProcessBulkScoreAdjustment(ctx context.Context, jobID string, params pubsub.BulkScoreAdjustmentParams) error {
	successCount, failureCount, err := p.service.CreateBulkScoreAdjustments(ctx, params)
	if err != nil {
		return err
	}

	// Mark job completed
	_, err = p.service.DB.Queries.MarkBulkJobCompleted(ctx, sqlc.MarkBulkJobCompletedParams{
		ID:             jobID,
		Processedcount: int32(successCount + failureCount),
		Successcount:   int32(successCount),
		Failurecount:   int32(failureCount),
	})
	return err
}

// ProcessFixMissingContentProgress processes missing content progress events
func (p *Processor) ProcessFixMissingContentProgress(ctx context.Context, jobID string, _ pubsub.FixMissingContentProgressParams) error {
	successCount, failureCount, err := p.service.FixMissingContentProgress(ctx)
	if err != nil {
		return err
	}

	// Mark job completed
	_, err = p.service.DB.Queries.MarkBulkJobCompleted(ctx, sqlc.MarkBulkJobCompletedParams{
		ID:             jobID,
		Processedcount: int32(successCount + failureCount),
		Successcount:   int32(successCount),
		Failurecount:   int32(failureCount),
	})
	return err
}
