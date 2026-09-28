package bulk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/firebase"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/pubsub"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/push"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service handles bulk operations for challenges and achievements
type Service struct {
	DB                        *database.DB
	Cache                     *cache.CacheWithRegistry
	Loaders                   *loaders.Loaders
	PushService               *push.Service
	FirebaseService           *firebase.Service
	Publisher                 *pubsub.Publisher
	Logger                    *slog.Logger
	ContentAchievementService *services.ContentAchievementService
}

// NewService creates a new bulk operations service
func NewService(
	db *database.DB,
	cache *cache.CacheWithRegistry,
	loaders *loaders.Loaders,
	pushService *push.Service,
	firebaseService *firebase.Service,
	publisher *pubsub.Publisher,
	logger *slog.Logger,
	contentAchievementService *services.ContentAchievementService,
) *Service {
	return &Service{
		DB:                        db,
		Cache:                     cache,
		Loaders:                   loaders,
		PushService:               pushService,
		FirebaseService:           firebaseService,
		Publisher:                 publisher,
		Logger:                    logger,
		ContentAchievementService: contentAchievementService,
	}
}

// CreateBulkJobAndPublish creates a bulk job record and publishes to Pub/Sub.
// The params type determines the operation type automatically.
func (s *Service) CreateBulkJobAndPublish(
	ctx context.Context,
	createdBy string,
	projectID *string,
	totalCount int,
	params pubsub.BulkOperationParams,
) (*model.BulkJob, error) {
	// Generate job ID
	jobID := ulid.NewBulkJobID()
	operationType := params.OperationType()

	// Marshal params to JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	// Convert createdBy to pointer (nil for M2M users who pass empty string)
	var createdByPtr *string
	if createdBy != "" {
		createdByPtr = &createdBy
	}

	// Create job record
	createParams := sqlc.CreateBulkJobParams{
		ID:            jobID,
		Operationtype: string(operationType),
		Status:        string(pubsub.JobStatusPending),
		Createdby:     createdByPtr,
		Projectid:     projectID,
		Inputparams:   paramsJSON,
		Totalcount:    int32(totalCount),
	}

	row, err := s.DB.Queries.CreateBulkJob(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create bulk job: %w", err)
	}

	// Publish to Pub/Sub if enabled
	if s.Publisher != nil && s.Publisher.IsEnabled() {
		msg := pubsub.BulkOperationMessage{
			JobID:         jobID,
			OperationType: operationType,
			CreatedBy:     createdBy,
			Params:        paramsJSON,
		}
		if projectID != nil {
			msg.ProjectID = *projectID
		}

		messageID, err := s.Publisher.PublishBulkOperation(ctx, msg)
		if err != nil {
			// Mark job as failed if we can't publish
			s.DB.Queries.MarkBulkJobFailed(ctx, sqlc.MarkBulkJobFailedParams{
				ID:           jobID,
				Errormessage: fmt.Sprintf("failed to publish to pub/sub: %v", err),
			})
			return nil, fmt.Errorf("failed to publish to pub/sub: %w", err)
		}

		// Update job with message ID
		s.DB.Queries.UpdateBulkJobMessageID(ctx, sqlc.UpdateBulkJobMessageIDParams{
			ID:        jobID,
			Messageid: messageID,
		})
	} else {
		// If Pub/Sub is disabled, process synchronously
		s.Logger.Warn("Pub/Sub disabled, processing bulk operation synchronously",
			"job_id", jobID,
			"operation_type", operationType,
		)
		// Mark as failed since we can't process async
		s.DB.Queries.MarkBulkJobFailed(ctx, sqlc.MarkBulkJobFailedParams{
			ID:           jobID,
			Errormessage: "pub/sub is disabled, async processing not available",
		})
		return nil, fmt.Errorf("pub/sub is disabled, use synchronous mutation instead")
	}

	return convertBulkJobRowToModel(row), nil
}

// RetryBulkJob creates a new bulk job with the same parameters as an existing job and publishes it.
func (s *Service) RetryBulkJob(ctx context.Context, jobID string, createdBy string) (*model.BulkJob, error) {
	// Fetch original job
	originalJob, err := s.DB.Queries.GetBulkJobByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original bulk job: %w", err)
	}

	// Generate new job ID
	newJobID := ulid.NewBulkJobID()

	// Convert createdBy to pointer
	var createdByPtr *string
	if createdBy != "" {
		createdByPtr = &createdBy
	}

	// Create new job with same params
	row, err := s.DB.Queries.CreateBulkJob(ctx, sqlc.CreateBulkJobParams{
		ID:            newJobID,
		Operationtype: originalJob.OperationType,
		Status:        string(pubsub.JobStatusPending),
		Createdby:     createdByPtr,
		Projectid:     originalJob.ProjectID,
		Inputparams:   originalJob.InputParams,
		Totalcount:    originalJob.TotalCount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create retry bulk job: %w", err)
	}

	// Publish to Pub/Sub
	if s.Publisher != nil && s.Publisher.IsEnabled() {
		msg := pubsub.BulkOperationMessage{
			JobID:         newJobID,
			OperationType: pubsub.OperationType(originalJob.OperationType),
			CreatedBy:     createdBy,
			Params:        originalJob.InputParams,
		}
		if originalJob.ProjectID != nil {
			msg.ProjectID = *originalJob.ProjectID
		}

		messageID, err := s.Publisher.PublishBulkOperation(ctx, msg)
		if err != nil {
			s.DB.Queries.MarkBulkJobFailed(ctx, sqlc.MarkBulkJobFailedParams{
				ID:           newJobID,
				Errormessage: fmt.Sprintf("failed to publish to pub/sub: %v", err),
			})
			return nil, fmt.Errorf("failed to publish to pub/sub: %w", err)
		}

		s.DB.Queries.UpdateBulkJobMessageID(ctx, sqlc.UpdateBulkJobMessageIDParams{
			ID:        newJobID,
			Messageid: messageID,
		})
	} else {
		s.DB.Queries.MarkBulkJobFailed(ctx, sqlc.MarkBulkJobFailedParams{
			ID:           newJobID,
			Errormessage: "pub/sub is disabled, async processing not available",
		})
		return nil, fmt.Errorf("pub/sub is disabled, use synchronous mutation instead")
	}

	return convertBulkJobRowToModel(row), nil
}

// GetBulkJob retrieves a bulk job by ID
func (s *Service) GetBulkJob(ctx context.Context, jobID string) (*model.BulkJob, error) {
	row, err := s.DB.Queries.GetBulkJobByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk job: %w", err)
	}
	return convertBulkJobRowToModel(row), nil
}

// GetMyBulkJobs retrieves bulk jobs created by a user
func (s *Service) GetMyBulkJobs(ctx context.Context, userID string, limit int) ([]*model.BulkJob, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.DB.Queries.GetBulkJobsByCreator(ctx, sqlc.GetBulkJobsByCreatorParams{
		Createdby:  userID,
		Limitcount: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk jobs: %w", err)
	}
	result := make([]*model.BulkJob, len(rows))
	for i, row := range rows {
		result[i] = convertBulkJobRowToModel(row)
	}
	return result, nil
}

// ListBulkJobsParams holds parameters for listing bulk jobs with pagination
type ListBulkJobsParams struct {
	Filter *model.BulkJobFilter
	First  *int
	After  *string
	Last   *int
	Before *string
}

// ListBulkJobsResult holds the result of listing bulk jobs
type ListBulkJobsResult struct {
	Jobs       []*model.BulkJob
	TotalCount int
	HasMore    bool
}

// ListBulkJobs retrieves bulk jobs with filtering and pagination
func (s *Service) ListBulkJobs(ctx context.Context, params ListBulkJobsParams) (*ListBulkJobsResult, error) {
	// Build filter params
	var status, operationType, projectID, createdBy *string
	if params.Filter != nil {
		if params.Filter.Status != nil {
			s := string(*params.Filter.Status)
			status = &s
		}
		operationType = params.Filter.OperationType
		projectID = params.Filter.ProjectID
		createdBy = params.Filter.CreatedBy
	}

	// Get total count
	totalCount, err := s.DB.Queries.CountBulkJobs(ctx, sqlc.CountBulkJobsParams{
		Status:        status,
		Operationtype: operationType,
		Projectid:     projectID,
		Createdby:     createdBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count bulk jobs: %w", err)
	}

	// Determine pagination direction and limit
	var rows []*sqlc.BulkJob
	var hasMore bool

	// Default limit
	limit := 20
	if params.First != nil {
		limit = *params.First
	} else if params.Last != nil {
		limit = *params.Last
	}

	// Fetch one extra to determine hasMore
	fetchLimit := limit + 1

	if params.Last != nil && params.Before != nil {
		// Backward pagination
		rows, err = s.DB.Queries.ListBulkJobsBackward(ctx, sqlc.ListBulkJobsBackwardParams{
			Status:        status,
			Operationtype: operationType,
			Projectid:     projectID,
			Createdby:     createdBy,
			Beforeid:      params.Before,
			Limitcount:    int32(fetchLimit),
		})
	} else {
		// Forward pagination (default)
		rows, err = s.DB.Queries.ListBulkJobsForward(ctx, sqlc.ListBulkJobsForwardParams{
			Status:        status,
			Operationtype: operationType,
			Projectid:     projectID,
			Createdby:     createdBy,
			Afterid:       params.After,
			Limitcount:    int32(fetchLimit),
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list bulk jobs: %w", err)
	}

	// Determine if there are more results
	if len(rows) > limit {
		hasMore = true
		rows = rows[:limit]
	}

	// Convert to model
	result := make([]*model.BulkJob, len(rows))
	for i, row := range rows {
		result[i] = convertBulkJobRowToModel(row)
	}

	return &ListBulkJobsResult{
		Jobs:       result,
		TotalCount: int(totalCount),
		HasMore:    hasMore,
	}, nil
}

// ResolveEnrollmentTarget converts an EnrollmentTargetInput to a list of user IDs
func (s *Service) ResolveEnrollmentTarget(ctx context.Context, target model.EnrollmentTargetInput) ([]string, error) {
	// Validate exactly one target type is specified
	setCount := 0
	if len(target.UserIds) > 0 {
		setCount++
	}
	if target.ChurchInProject != nil {
		setCount++
	}
	if len(target.TeamIds) > 0 {
		setCount++
	}
	if len(target.SuperTeamIds) > 0 {
		setCount++
	}
	if target.AllProjectMembers != nil {
		setCount++
	}

	if setCount == 0 {
		return nil, fmt.Errorf("must specify at least one target type")
	}
	if setCount > 1 {
		return nil, fmt.Errorf("must specify exactly one target type")
	}

	// Resolve based on target type
	if len(target.UserIds) > 0 {
		return target.UserIds, nil
	}

	if target.ChurchInProject != nil {
		userIDs, err := s.DB.Queries.GetUserIDsInChurchAndProject(ctx, sqlc.GetUserIDsInChurchAndProjectParams{
			Churchid:  target.ChurchInProject.ChurchID,
			Projectid: target.ChurchInProject.ProjectID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get users in church and project: %w", err)
		}
		return userIDs, nil
	}

	if len(target.TeamIds) > 0 {
		userIDs, err := s.DB.Queries.GetUserIDsInTeams(ctx, target.TeamIds)
		if err != nil {
			return nil, fmt.Errorf("failed to get team members: %w", err)
		}
		return userIDs, nil
	}

	if len(target.SuperTeamIds) > 0 {
		userIDs, err := s.DB.Queries.GetUserIDsInSuperTeams(ctx, target.SuperTeamIds)
		if err != nil {
			return nil, fmt.Errorf("failed to get super team members: %w", err)
		}
		return userIDs, nil
	}

	if target.AllProjectMembers != nil {
		userIDs, err := s.DB.Queries.GetUserIDsInProject(ctx, *target.AllProjectMembers)
		if err != nil {
			return nil, fmt.Errorf("failed to get project members: %w", err)
		}
		return userIDs, nil
	}

	return nil, fmt.Errorf("no valid target specified")
}

// convertBulkJobRowToModel converts a database row to a GraphQL model
func convertBulkJobRowToModel(row *sqlc.BulkJob) *model.BulkJob {
	job := &model.BulkJob{
		ID:             row.ID,
		OperationType:  row.OperationType,
		Status:         model.BulkJobStatus(row.Status),
		TotalCount:     int(row.TotalCount),
		ProcessedCount: int(row.ProcessedCount),
		SuccessCount:   int(row.SuccessCount),
		FailureCount:   int(row.FailureCount),
		ErrorMessage:   row.ErrorMessage,
	}

	if row.CreatedAt.Valid {
		job.CreatedAt = scalars.DateTime{Time: row.CreatedAt.Time}
	}

	if row.StartedAt.Valid {
		t := scalars.DateTime{Time: row.StartedAt.Time}
		job.StartedAt = &t
	}

	if row.CompletedAt.Valid {
		t := scalars.DateTime{Time: row.CompletedAt.Time}
		job.CompletedAt = &t
	}

	return job
}

// EnrollUsersInChallenge enrolls users in a challenge in batches
func (s *Service) EnrollUsersInChallenge(ctx context.Context, challengeID string, userIDs []string, sendNotifications bool) (int, int, error) {
	if len(userIDs) == 0 {
		return 0, 0, nil
	}

	// Load challenge for cache invalidation
	thunk := s.Loaders.ChallengeByIDLoader.Load(ctx, challengeID)
	challenge, err := thunk()
	if err != nil {
		return 0, 0, fmt.Errorf("challenge not found: %w", err)
	}

	projectID := getChallengeProjectID(challenge)
	eventID := getChallengeEventID(challenge)

	successCount := 0
	failureCount := 0

	// Process in batches
	for i := 0; i < len(userIDs); i += pubsub.BatchSize {
		end := i + pubsub.BatchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		batch := userIDs[i:end]

		// Bulk enroll
		err := s.DB.Queries.BulkEnrollUsersInChallenge(ctx, sqlc.BulkEnrollUsersInChallengeParams{
			Userids:     batch,
			Challengeid: challengeID,
		})
		if err != nil {
			s.Logger.Error("Failed to enroll batch",
				"challenge_id", challengeID,
				"batch_start", i,
				"batch_size", len(batch),
				"error", err,
			)
			failureCount += len(batch)
			continue
		}

		successCount += len(batch)

		// Cache invalidation
		for _, userID := range batch {
			s.Cache.InvalidateUser(userID)
		}

		// Notify Firestore listeners
		if s.FirebaseService != nil {
			for _, userID := range batch {
				go s.FirebaseService.NotifyUserChallenges(context.Background(), userID)
			}
		}

		// Send push notifications
		if sendNotifications && s.PushService != nil {
			challengeInfo := getChallengePushInfo(challenge)
			for _, userID := range batch {
				go push.SendTranslatedChallengeEnrollmentNotification(
					s.PushService,
					s.Loaders,
					userID,
					challengeInfo,
				)
			}
		}
	}

	// Invalidate challenge cache once
	s.Cache.InvalidateChallenge(challengeID, projectID, eventID)

	return successCount, failureCount, nil
}

// UnenrollUsersFromChallenge unenrolls users from a challenge in batches
func (s *Service) UnenrollUsersFromChallenge(ctx context.Context, challengeID string, userIDs []string) (int, int, error) {
	if len(userIDs) == 0 {
		return 0, 0, nil
	}

	// Load challenge for cache invalidation
	thunk := s.Loaders.ChallengeByIDLoader.Load(ctx, challengeID)
	challenge, err := thunk()
	if err != nil {
		return 0, 0, fmt.Errorf("challenge not found: %w", err)
	}

	projectID := getChallengeProjectID(challenge)
	eventID := getChallengeEventID(challenge)

	successCount := 0
	failureCount := 0

	// Process in batches
	for i := 0; i < len(userIDs); i += pubsub.BatchSize {
		end := i + pubsub.BatchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		batch := userIDs[i:end]

		// Bulk unenroll
		err := s.DB.Queries.BulkUnenrollUsersFromChallenge(ctx, sqlc.BulkUnenrollUsersFromChallengeParams{
			Challengeid: challengeID,
			Userids:     batch,
		})
		if err != nil {
			s.Logger.Error("Failed to unenroll batch",
				"challenge_id", challengeID,
				"batch_start", i,
				"batch_size", len(batch),
				"error", err,
			)
			failureCount += len(batch)
			continue
		}

		successCount += len(batch)

		// Cache invalidation
		for _, userID := range batch {
			s.Cache.InvalidateUser(userID)
		}

		// Notify Firestore listeners
		if s.FirebaseService != nil {
			for _, userID := range batch {
				go s.FirebaseService.NotifyUserChallenges(context.Background(), userID)
			}
		}
	}

	// Invalidate challenge cache once
	s.Cache.InvalidateChallenge(challengeID, projectID, eventID)

	return successCount, failureCount, nil
}

// CompleteUsersChallenge marks challenges as completed for users in batches
func (s *Service) CompleteUsersChallenge(ctx context.Context, challengeID string, userIDs []string, completedAt *time.Time) (int, int, error) {
	if len(userIDs) == 0 {
		return 0, 0, nil
	}

	// Load challenge for cache invalidation
	thunk := s.Loaders.ChallengeByIDLoader.Load(ctx, challengeID)
	challenge, err := thunk()
	if err != nil {
		return 0, 0, fmt.Errorf("challenge not found: %w", err)
	}

	projectID := getChallengeProjectID(challenge)
	eventID := getChallengeEventID(challenge)

	successCount := 0
	failureCount := 0

	// Convert completedAt for database
	var completedAtTimestamp pgtype.Timestamptz
	if completedAt != nil {
		completedAtTimestamp = pgtype.Timestamptz{Time: *completedAt, Valid: true}
	}

	// Process in batches
	for i := 0; i < len(userIDs); i += pubsub.BatchSize {
		end := i + pubsub.BatchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		batch := userIDs[i:end]

		// Bulk complete
		err := s.DB.Queries.BulkCompleteChallenges(ctx, sqlc.BulkCompleteChallengesParams{
			Userids:     batch,
			Challengeid: challengeID,
			Completedat: completedAtTimestamp,
		})
		if err != nil {
			s.Logger.Error("Failed to complete batch",
				"challenge_id", challengeID,
				"batch_start", i,
				"batch_size", len(batch),
				"error", err,
			)
			failureCount += len(batch)
			continue
		}

		successCount += len(batch)

		// Cache invalidation
		for _, userID := range batch {
			s.Cache.InvalidateUser(userID)
		}

		// Notify Firestore listeners
		if s.FirebaseService != nil {
			for _, userID := range batch {
				go s.FirebaseService.NotifyUserChallenges(context.Background(), userID)
			}
		}
	}

	// Invalidate challenge cache once
	s.Cache.InvalidateChallenge(challengeID, projectID, eventID)

	return successCount, failureCount, nil
}

// PublishChallenges publishes challenges in batches
func (s *Service) PublishChallenges(ctx context.Context, challengeIDs []string, publishedAt time.Time) (int, int, error) {
	if len(challengeIDs) == 0 {
		return 0, 0, nil
	}

	successCount := 0
	failureCount := 0
	notifiedProjects := make(map[string]bool)

	// Load challenges for cache invalidation
	for _, id := range challengeIDs {
		thunk := s.Loaders.ChallengeByIDLoader.Load(ctx, id)
		challenge, err := thunk()
		if err != nil {
			s.Logger.Error("Failed to load challenge for publishing",
				"challenge_id", id,
				"error", err,
			)
			failureCount++
			continue
		}

		// Publish individual challenge
		_, err = s.DB.Queries.PublishChallenge(ctx, sqlc.PublishChallengeParams{
			ID:          id,
			Publishedat: pgtype.Timestamptz{Time: publishedAt, Valid: true},
		})
		if err != nil {
			s.Logger.Error("Failed to publish challenge",
				"challenge_id", id,
				"error", err,
			)
			failureCount++
			continue
		}

		successCount++

		// Cache invalidation
		projectID := getChallengeProjectID(challenge)
		eventID := getChallengeEventID(challenge)
		s.Cache.InvalidateChallenge(id, projectID, eventID)
		notifiedProjects[projectID] = true
	}

	// Notify Firestore listeners for each unique project
	if s.FirebaseService != nil {
		for projectID := range notifiedProjects {
			go s.FirebaseService.NotifyProjectChallenges(context.Background(), projectID)
		}
	}

	return successCount, failureCount, nil
}

// AwardAchievements awards an achievement to users in batches
func (s *Service) AwardAchievements(ctx context.Context, achievementID string, userIDs []string) (int, int, error) {
	if len(userIDs) == 0 {
		return 0, 0, nil
	}

	// Validate achievement exists
	thunk := s.Loaders.AchievementByIDLoader.Load(ctx, achievementID)
	_, err := thunk()
	if err != nil {
		return 0, 0, fmt.Errorf("achievement not found: %w", err)
	}

	successCount := 0
	failureCount := 0

	// Process in batches
	for i := 0; i < len(userIDs); i += pubsub.BatchSize {
		end := i + pubsub.BatchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		batch := userIDs[i:end]

		// Award to each user in batch
		for _, userID := range batch {
			err := s.DB.Queries.AwardUserAchievement(ctx, sqlc.AwardUserAchievementParams{
				UserID:        userID,
				AchievementID: achievementID,
				AchievedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			})
			if err != nil {
				s.Logger.Error("Failed to award achievement",
					"achievement_id", achievementID,
					"user_id", userID,
					"error", err,
				)
				failureCount++
				continue
			}

			successCount++
			s.Cache.InvalidateUser(userID)

			// Notify Firestore listeners
			if s.FirebaseService != nil {
				go s.FirebaseService.NotifyUserAchievements(context.Background(), userID)
			}
		}
	}

	// Invalidate achievement cache once
	s.Cache.InvalidateAchievement(achievementID)

	return successCount, failureCount, nil
}

// Helper functions for extracting challenge info

func getChallengeProjectID(challenge model.Challenge) string {
	switch c := challenge.(type) {
	case *model.SimpleChallenge:
		return c.ProjectID
	case *model.QuizChallenge:
		return c.ProjectID
	case *model.ExternalChallenge:
		return c.ProjectID
	case *model.PluginChallenge:
		return c.ProjectID
	}
	return ""
}

func getChallengeEventID(challenge model.Challenge) *string {
	switch c := challenge.(type) {
	case *model.SimpleChallenge:
		return c.EventID
	case *model.QuizChallenge:
		return c.EventID
	case *model.ExternalChallenge:
		return c.EventID
	case *model.PluginChallenge:
		return c.EventID
	}
	return nil
}

func getChallengePushInfo(challenge model.Challenge) push.ChallengeInfo {
	switch c := challenge.(type) {
	case *model.SimpleChallenge:
		return push.ChallengeInfo{
			ID:               c.ID,
			Name:             c.Name,
			NotificationText: c.NotificationText,
		}
	case *model.QuizChallenge:
		return push.ChallengeInfo{
			ID:               c.ID,
			Name:             c.Name,
			NotificationText: c.NotificationText,
		}
	case *model.ExternalChallenge:
		return push.ChallengeInfo{
			ID:               c.ID,
			Name:             c.Name,
			NotificationText: c.NotificationText,
		}
	case *model.PluginChallenge:
		return push.ChallengeInfo{
			ID:               c.ID,
			Name:             c.Name,
			NotificationText: c.NotificationText,
		}
	}
	return push.ChallengeInfo{}
}

// GrantQuizSessionAccess grants quiz session access to users.
// This is the shared implementation used by both sync and async mutations.
func (s *Service) GrantQuizSessionAccess(ctx context.Context, params pubsub.BulkGrantQuizSessionAccessParams) (int, int, error) {
	// 1. Collect user IDs from various sources with source type tracking
	userIdsToGrant := make(map[string]string) // map[userID]sourceType

	// Direct user IDs
	for _, uid := range params.UserIDs {
		userIdsToGrant[uid] = "DIRECT"
	}

	// Team members
	if len(params.TeamIDs) > 0 {
		for _, teamID := range params.TeamIDs {
			thunk := s.Loaders.UserIDsByTeamLoader.Load(ctx, teamID)
			teamUserIDs, err := thunk()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get team members: %w", err)
			}
			for _, uid := range teamUserIDs {
				if _, exists := userIdsToGrant[uid]; !exists {
					userIdsToGrant[uid] = "TEAM"
				}
			}
		}
	}

	// Super team members
	if len(params.SuperTeamIDs) > 0 {
		for _, superTeamID := range params.SuperTeamIDs {
			thunk := s.Loaders.UserIDsBySuperTeamLoader.Load(ctx, superTeamID)
			superTeamUserIDs, err := thunk()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get super team members: %w", err)
			}
			for _, uid := range superTeamUserIDs {
				if _, exists := userIdsToGrant[uid]; !exists {
					userIdsToGrant[uid] = "SUPER_TEAM"
				}
			}
		}
	}

	// Church members
	if len(params.ChurchIDs) > 0 {
		for _, churchID := range params.ChurchIDs {
			key := loaders.ChurchProjectKey{ChurchID: churchID, ProjectID: params.ProjectID}
			thunk := s.Loaders.UserIDsByChurchInProjectLoader.Load(ctx, key)
			churchUserIDs, err := thunk()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get church members: %w", err)
			}
			for _, uid := range churchUserIDs {
				if _, exists := userIdsToGrant[uid]; !exists {
					userIdsToGrant[uid] = "CHURCH"
				}
			}
		}
	}

	// All project users
	if params.AllProjectUsers {
		thunk := s.Loaders.UserIDsInProjectLoader.Load(ctx, params.ProjectID)
		projectUserIDs, err := thunk()
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get project users: %w", err)
		}
		for _, uid := range projectUserIDs {
			if _, exists := userIdsToGrant[uid]; !exists {
				userIdsToGrant[uid] = "ALL"
			}
		}
	}

	if len(userIdsToGrant) == 0 {
		return 0, 0, nil
	}

	// 2. Batch insert (500 per batch)
	successCount := 0
	failureCount := 0
	const batchSize = 500

	ids := make([]string, 0, batchSize)
	sessionIds := make([]string, 0, batchSize)
	userIds := make([]string, 0, batchSize)
	grantedBys := make([]string, 0, batchSize)
	sourceTypes := make([]string, 0, batchSize)
	sourceIds := make([]string, 0, batchSize)

	for uid, sourceType := range userIdsToGrant {
		ids = append(ids, ulid.NewQuizSessionAccessID())
		sessionIds = append(sessionIds, params.SessionID)
		userIds = append(userIds, uid)
		grantedBys = append(grantedBys, params.GrantedBy)
		sourceTypes = append(sourceTypes, sourceType)
		sourceIds = append(sourceIds, "")

		if len(ids) >= batchSize {
			granted, err := s.DB.Queries.BulkCreateQuizSessionAccess(ctx, sqlc.BulkCreateQuizSessionAccessParams{
				Ids:         ids,
				Sessionids:  sessionIds,
				Userids:     userIds,
				Grantedbys:  grantedBys,
				Sourcetypes: sourceTypes,
				Sourceids:   sourceIds,
			})
			if err != nil {
				s.Logger.Error("Failed to grant access batch",
					"session_id", params.SessionID,
					"batch_size", len(ids),
					"error", err,
				)
				failureCount += len(ids)
			} else {
				successCount += int(granted)
			}
			// Reset slices
			ids = ids[:0]
			sessionIds = sessionIds[:0]
			userIds = userIds[:0]
			grantedBys = grantedBys[:0]
			sourceTypes = sourceTypes[:0]
			sourceIds = sourceIds[:0]
		}
	}

	// Insert remaining
	if len(ids) > 0 {
		granted, err := s.DB.Queries.BulkCreateQuizSessionAccess(ctx, sqlc.BulkCreateQuizSessionAccessParams{
			Ids:         ids,
			Sessionids:  sessionIds,
			Userids:     userIds,
			Grantedbys:  grantedBys,
			Sourcetypes: sourceTypes,
			Sourceids:   sourceIds,
		})
		if err != nil {
			s.Logger.Error("Failed to grant access batch",
				"session_id", params.SessionID,
				"batch_size", len(ids),
				"error", err,
			)
			failureCount += len(ids)
		} else {
			successCount += int(granted)
		}
	}

	if successCount > 0 {
		s.Cache.InvalidateQuizSessionAccess()
	}

	// 3. Firebase notification
	if s.FirebaseService != nil && successCount > 0 {
		go s.FirebaseService.NotifyProjectQuizSessions(context.Background(), params.ProjectID)
	}

	return successCount, failureCount, nil
}

// CreateBulkScoreAdjustments creates score journal entries for multiple users in batches
func (s *Service) CreateBulkScoreAdjustments(
	ctx context.Context,
	params pubsub.BulkScoreAdjustmentParams,
) (int, int, error) {
	if len(params.Adjustments) == 0 {
		return 0, 0, nil
	}

	s.Logger.Info("CreateBulkScoreAdjustments: starting",
		"project_id", params.ProjectID,
		"event_id", params.EventID,
		"adjustment_count", len(params.Adjustments),
		"awarded_by", params.AwardedBy,
	)

	// Determine awarded_by pointer (nil for M2M)
	var awardedByPtr *string
	if params.AwardedBy != "" {
		awardedByPtr = &params.AwardedBy
	}

	// Convert eventID to pointer
	var eventIDPtr *string
	if params.EventID != "" {
		eventIDPtr = &params.EventID
	}

	successCount := 0
	failureCount := 0

	// Process in batches of 100
	for i := 0; i < len(params.Adjustments); i += pubsub.BatchSize {
		end := i + pubsub.BatchSize
		if end > len(params.Adjustments) {
			end = len(params.Adjustments)
		}
		batch := params.Adjustments[i:end]

		s.Logger.Debug("CreateBulkScoreAdjustments: processing batch",
			"batch_start", i,
			"batch_size", len(batch),
		)

		// Prepare arrays for batch insert
		ids := make([]string, len(batch))
		userIDs := make([]string, len(batch))
		points := make([]int32, len(batch))
		reasons := make([]string, len(batch))

		for j, adj := range batch {
			ids[j] = ulid.NewScoreJournalID()
			userIDs[j] = adj.UserID
			points[j] = adj.Points
			reasons[j] = adj.Reason
		}

		// Execute batch insert
		_, err := s.DB.Queries.CreateBulkScoreAdjustmentBatch(ctx, sqlc.CreateBulkScoreAdjustmentBatchParams{
			Ids:       ids,
			ProjectID: params.ProjectID,
			UserIds:   userIDs,
			EventID:   eventIDPtr,
			Points:    points,
			Reasons:   reasons,
			AwardedBy: awardedByPtr,
		})
		if err != nil {
			s.Logger.Error("Failed to create bulk score adjustments",
				"batch_start", i,
				"batch_size", len(batch),
				"project_id", params.ProjectID,
				"project_id_len", len(params.ProjectID),
				"event_id", params.EventID,
				"event_id_len", len(params.EventID),
				"first_user_id", userIDs[0],
				"first_user_id_len", len(userIDs[0]),
				"first_id", ids[0],
				"first_id_len", len(ids[0]),
				"error", err,
			)
			failureCount += len(batch)
			continue
		}

		successCount += len(batch)

		// Cache invalidation per batch
		for _, adj := range batch {
			s.Cache.InvalidateUser(adj.UserID)
		}
	}

	// Invalidate project and event caches once at end
	s.Cache.InvalidateProject(params.ProjectID)
	if params.EventID != "" {
		s.Cache.InvalidateEvent(params.EventID)
	}

	s.Logger.Info("CreateBulkScoreAdjustments: completed",
		"project_id", params.ProjectID,
		"success_count", successCount,
		"failure_count", failureCount,
	)

	return successCount, failureCount, nil
}

// FixMissingContentProgress processes missing content events using the ContentAchievementService.
// This ensures all hooks (cache invalidation, webhooks, push notifications, Firebase, score journals) are triggered.
func (s *Service) FixMissingContentProgress(ctx context.Context, params pubsub.FixMissingContentProgressParams) (int, int, error) {
	if len(params.UserIDs) == 0 {
		return 0, 0, nil
	}

	// Get missing events for the specific users in this batch
	events, err := s.DB.Queries.GetMissingContentEventsForUsers(ctx, params.UserIDs)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get missing content events: %w", err)
	}

	if len(events) == 0 {
		return 0, 0, nil
	}

	// Track unique users for Firebase notification
	userSet := make(map[string]bool)
	for _, event := range events {
		userSet[event.UserID] = true
	}

	s.Logger.Info("FixMissingContentProgress: processing missing content events",
		"event_count", len(events),
		"user_count", len(userSet),
		"batch_user_count", len(params.UserIDs))

	successCount := 0
	failureCount := 0

	// Process each event using the ContentAchievementService
	for _, event := range events {
		s.ContentAchievementService.ProcessContentEvent(ctx, event.UserID, event.TaskID)
		successCount++
	}

	// Notify Firebase for content updates for each affected user
	if s.FirebaseService != nil {
		for userID := range userSet {
			go s.FirebaseService.NotifyUserContent(context.Background(), userID)
		}
	}

	s.Logger.Info("FixMissingContentProgress: completed",
		"success_count", successCount,
		"failure_count", failureCount,
		"users_notified", len(userSet))

	return successCount, failureCount, nil
}

// FixMissingStreakProgress processes missing streak events using the ContentAchievementService.
// Uses the original consumed_at timestamp for correct deadline enforcement.
func (s *Service) FixMissingStreakProgress(ctx context.Context, params pubsub.FixMissingStreakProgressParams) (int, int, error) {
	if len(params.UserIDs) == 0 {
		return 0, 0, nil
	}

	events, err := s.DB.Queries.GetMissingStreakEventsForUsers(ctx, params.UserIDs)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get missing streak events: %w", err)
	}

	if len(events) == 0 {
		return 0, 0, nil
	}

	userSet := make(map[string]bool)
	for _, event := range events {
		userSet[event.UserID] = true
	}

	s.Logger.Info("FixMissingStreakProgress: processing missing streak events",
		"event_count", len(events),
		"user_count", len(userSet),
		"batch_user_count", len(params.UserIDs))

	successCount := 0
	failureCount := 0

	for _, event := range events {
		if !event.ConsumedAt.Valid {
			failureCount++
			continue
		}
		s.ContentAchievementService.ProcessStreakEvent(ctx, event.UserID, event.TaskID, event.ConsumedAt.Time)
		successCount++
	}

	if s.FirebaseService != nil {
		for userID := range userSet {
			go s.FirebaseService.NotifyUserContent(context.Background(), userID)
		}
	}

	s.Logger.Info("FixMissingStreakProgress: completed",
		"success_count", successCount,
		"failure_count", failureCount,
		"users_notified", len(userSet))

	return successCount, failureCount, nil
}
