package pubsub

import "encoding/json"

// OperationType represents the type of bulk operation
type OperationType string

const (
	// Challenge operations
	OperationBulkEnrollChallenge   OperationType = "BULK_ENROLL_CHALLENGE"
	OperationBulkUnenrollChallenge OperationType = "BULK_UNENROLL_CHALLENGE"
	OperationBulkCompleteChallenge OperationType = "BULK_COMPLETE_CHALLENGE"
	OperationBulkPublishChallenge  OperationType = "BULK_PUBLISH_CHALLENGE"

	// Achievement operations
	OperationBulkAwardAchievement OperationType = "BULK_AWARD_ACHIEVEMENT"

	// Quiz session operations
	OperationBulkGrantQuizSessionAccess OperationType = "BULK_GRANT_QUIZ_SESSION_ACCESS"

	// Score adjustment operations
	OperationBulkScoreAdjustment OperationType = "BULK_SCORE_ADJUSTMENT"

	// Maintenance operations
	OperationFixMissingContentProgress OperationType = "FIX_MISSING_CONTENT_PROGRESS"
)

// BulkOperationMessage is the message published to Pub/Sub for async processing
type BulkOperationMessage struct {
	JobID         string          `json:"job_id"`
	OperationType OperationType   `json:"operation_type"`
	CreatedBy     string          `json:"created_by"`
	ProjectID     string          `json:"project_id,omitempty"`
	Params        json.RawMessage `json:"params"`
}

// BulkOperationParams is implemented by all typed bulk operation parameter structs.
// Each implementation returns its corresponding operation type.
type BulkOperationParams interface {
	OperationType() OperationType
}

// BulkEnrollChallengeParams contains parameters for bulk challenge enrollment
type BulkEnrollChallengeParams struct {
	ChallengeID string   `json:"challenge_id"`
	UserIDs     []string `json:"user_ids"`
}

func (p BulkEnrollChallengeParams) OperationType() OperationType {
	return OperationBulkEnrollChallenge
}

// BulkUnenrollChallengeParams contains parameters for bulk challenge unenrollment
type BulkUnenrollChallengeParams struct {
	ChallengeID string   `json:"challenge_id"`
	UserIDs     []string `json:"user_ids"`
}

func (p BulkUnenrollChallengeParams) OperationType() OperationType {
	return OperationBulkUnenrollChallenge
}

// BulkCompleteChallengeParams contains parameters for bulk challenge completion
type BulkCompleteChallengeParams struct {
	ChallengeID string   `json:"challenge_id"`
	UserIDs     []string `json:"user_ids"`
	CompletedAt string   `json:"completed_at,omitempty"` // RFC3339 timestamp
}

func (p BulkCompleteChallengeParams) OperationType() OperationType {
	return OperationBulkCompleteChallenge
}

// BulkPublishChallengeParams contains parameters for bulk challenge publishing
type BulkPublishChallengeParams struct {
	ChallengeIDs []string `json:"challenge_ids"`
	PublishedAt  string   `json:"published_at"` // RFC3339 timestamp
}

func (p BulkPublishChallengeParams) OperationType() OperationType {
	return OperationBulkPublishChallenge
}

// BulkAwardAchievementParams contains parameters for bulk achievement awarding
type BulkAwardAchievementParams struct {
	AchievementID string   `json:"achievement_id"`
	UserIDs       []string `json:"user_ids"`
	TeamID        string   `json:"team_id,omitempty"`
}

func (p BulkAwardAchievementParams) OperationType() OperationType {
	return OperationBulkAwardAchievement
}

// BulkGrantQuizSessionAccessParams contains parameters for bulk quiz session access granting
type BulkGrantQuizSessionAccessParams struct {
	SessionID       string   `json:"session_id"`
	UserIDs         []string `json:"user_ids,omitempty"`
	TeamIDs         []string `json:"team_ids,omitempty"`
	SuperTeamIDs    []string `json:"super_team_ids,omitempty"`
	ChurchIDs       []string `json:"church_ids,omitempty"`
	AllProjectUsers bool     `json:"all_project_users,omitempty"`
	ProjectID       string   `json:"project_id"`
	GrantedBy       string   `json:"granted_by"`
}

func (p BulkGrantQuizSessionAccessParams) OperationType() OperationType {
	return OperationBulkGrantQuizSessionAccess
}

// BulkScoreAdjustmentItem represents a single score adjustment for a user
type BulkScoreAdjustmentItem struct {
	UserID string `json:"user_id"`
	Points int32  `json:"points"`
	Reason string `json:"reason,omitempty"`
}

// BulkScoreAdjustmentParams contains parameters for bulk score adjustment
type BulkScoreAdjustmentParams struct {
	ProjectID   string                    `json:"project_id"`
	EventID     string                    `json:"event_id,omitempty"`
	Adjustments []BulkScoreAdjustmentItem `json:"adjustments"`
	AwardedBy   string                    `json:"awarded_by,omitempty"`
}

func (p BulkScoreAdjustmentParams) OperationType() OperationType {
	return OperationBulkScoreAdjustment
}

// FixMissingContentProgressParams contains parameters for fixing missing content progress
// (empty struct - processes all pending events from the database)
type FixMissingContentProgressParams struct{}

func (p FixMissingContentProgressParams) OperationType() OperationType {
	return OperationFixMissingContentProgress
}

// JobStatus represents the status of a bulk job
type JobStatus string

const (
	JobStatusPending    JobStatus = "PENDING"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
)

// PubSubPushMessage represents the structure of a GCP Pub/Sub push delivery message
type PubSubPushMessage struct {
	Message struct {
		Attributes  map[string]string `json:"attributes"`
		Data        string            `json:"data"` // Base64 encoded
		MessageID   string            `json:"messageId"`
		PublishTime string            `json:"publishTime"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}
