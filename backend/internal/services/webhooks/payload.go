package webhooks

import (
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// EventType represents the type of webhook event
type EventType string

const (
	EventTypeExternalContent     EventType = "external_content_event"
	EventTypePointsAwarded       EventType = "points_awarded"
	EventTypeQuizSessionFinished EventType = "quiz_session_finished"
	EventTypeTeamNameChanged     EventType = "team_name_changed"
	EventTypeQuizFinalized       EventType = "quiz_finalized"
)

// WebhookPayload is the base structure sent to webhook endpoints
type WebhookPayload struct {
	EventType string      `json:"event_type"`
	Timestamp time.Time   `json:"timestamp"`
	ProjectID string      `json:"project_id"`
	User      *UserData   `json:"user,omitempty"`
	Data      interface{} `json:"data"`
}

// UserData contains user information included in webhook payloads
type UserData struct {
	ID        string `json:"id"`
	MembersID string `json:"members_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}

// UserRow is a type constraint for sqlc user row types that have the fields needed for UserData.
type UserRow interface {
	*sqlc.GetUserByIDRow | *sqlc.GetUserByPersonUUIDRow
}

// NewUserData creates a UserData from a sqlc user row.
func NewUserData[T UserRow](user T) *UserData {
	name := ""
	// Use type assertion to access fields since Go generics don't support field access
	switch u := any(user).(type) {
	case *sqlc.GetUserByIDRow:
		if u.DisplayName != nil {
			name = *u.DisplayName
		}
		return &UserData{
			ID:        u.ID,
			MembersID: u.MembersID,
			Email:     u.Email,
			Name:      name,
		}
	case *sqlc.GetUserByPersonUUIDRow:
		if u.DisplayName != nil {
			name = *u.DisplayName
		}
		return &UserData{
			ID:        u.ID,
			MembersID: u.MembersID,
			Email:     u.Email,
			Name:      name,
		}
	}
	return nil
}

// ExternalContentEventData for external_content_event webhooks
type ExternalContentEventData struct {
	TaskID             string  `json:"task_id"`
	PlanID             string  `json:"plan_id,omitempty"`
	ContentProgress    float64 `json:"content_progress,omitempty"`
	ConsumedAt         string  `json:"consumed_at"`
	AchievementID      string  `json:"achievement_id,omitempty"`
	AchievementAwarded bool    `json:"achievement_awarded"`
}

// PointsAwardedData for points_awarded webhooks
type PointsAwardedData struct {
	Points      int32  `json:"points"`
	SourceType  string `json:"source_type"` // ACHIEVEMENT, QUIZ, CHALLENGE, MANUAL
	SourceID    string `json:"source_id,omitempty"`
	Reason      string `json:"reason,omitempty"`
	EventID     string `json:"event_id,omitempty"`
	ChallengeID string `json:"challenge_id,omitempty"`
}

// TestEventData for test webhook calls
type TestEventData struct {
	Message string `json:"message"`
}

// QuizSessionFinishedData for quiz_session_finished webhooks
type QuizSessionFinishedData struct {
	SessionID   string           `json:"session_id"`
	SessionName string           `json:"session_name,omitempty"`
	QuizID      string           `json:"quiz_id"`
	QuizName    string           `json:"quiz_name"`
	ChallengeID string           `json:"challenge_id"`
	FinishedAt  time.Time        `json:"finished_at"`
	Results     []QuizUserResult `json:"results"`
}

// QuizUserResult contains a single user's quiz result
type QuizUserResult struct {
	UserID        string   `json:"user_id"`
	MembersID     string   `json:"members_id"`
	ChurchID      string   `json:"church_id"`
	Score         *int32   `json:"score"`
	MaxScore      *int32   `json:"max_score"`
	ScorePercent  *float64 `json:"score_percentage"`
	Completed     bool     `json:"completed"`
	AutoSubmitted bool     `json:"auto_submitted"`
}

// TeamNameChangedData contains data for team name change webhook
type TeamNameChangedData struct {
	TeamID  string `json:"team_id"`
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

// QuizFinalizedData contains data for quiz finalized webhook (user completes their submission)
type QuizFinalizedData struct {
	SubmissionID        string    `json:"submission_id"`
	QuizID              string    `json:"quiz_id"`
	QuizName            string    `json:"quiz_name"`
	ChallengeID         string    `json:"challenge_id"`
	ChallengeName       string    `json:"challenge_name,omitempty"`
	EventID             *string   `json:"event_id,omitempty"`
	SessionID           *string   `json:"session_id,omitempty"`
	Score               int32     `json:"score"`
	MaxScore            int32     `json:"max_score"`
	ScorePercentage     float64   `json:"score_percentage"`
	PointsAwarded       int32     `json:"points_awarded"`
	AchievementsAwarded []string  `json:"achievements_awarded"`
	CompletedAt         time.Time `json:"completed_at"`
}
