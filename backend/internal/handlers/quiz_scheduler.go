package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/firebase"
	"github.com/bcc-media/wayfarer/internal/services/webhooks"
	"github.com/gin-gonic/gin"
)

// QuizSchedulerHandler handles scheduled quiz session state transitions
type QuizSchedulerHandler struct {
	DB              *database.DB
	FirebaseService *firebase.Service
	WebhookService  *webhooks.Service
	Cache           *cache.CacheWithRegistry
}

// QuizSchedulerResponse contains the results of processing scheduled transitions
type QuizSchedulerResponse struct {
	Opened   int `json:"opened"`
	Locked   int `json:"locked"`
	Finished int `json:"finished"`
	Errors   int `json:"errors"`
}

// ProcessScheduledTransitions processes all pending quiz session state transitions
// This endpoint is designed to be called by Cloud Scheduler every minute
func (h *QuizSchedulerHandler) ProcessScheduledTransitions(c *gin.Context) {
	ctx := c.Request.Context()

	slog.Info("quiz_scheduler: starting scheduled transitions")

	response := QuizSchedulerResponse{}

	// Process pending opens (DRAFT -> OPEN)
	opened, openErrors := h.processOpenTransitions(ctx)
	response.Opened = opened
	response.Errors += openErrors

	// Process pending locks (OPEN -> LOCKED)
	locked, lockErrors := h.processLockTransitions(ctx)
	response.Locked = locked
	response.Errors += lockErrors

	// Process pending finishes (LOCKED -> FINISHED)
	finished, finishErrors := h.processFinishTransitions(ctx)
	response.Finished = finished
	response.Errors += finishErrors

	slog.Info("quiz_scheduler: scheduled transitions complete",
		"opened", response.Opened,
		"locked", response.Locked,
		"finished", response.Finished,
		"errors", response.Errors,
	)

	c.JSON(http.StatusOK, response)
}

// processOpenTransitions opens sessions that have reached their open_at time
func (h *QuizSchedulerHandler) processOpenTransitions(ctx context.Context) (int, int) {
	sessions, err := h.DB.Queries.GetSessionsPendingOpen(ctx)
	if err != nil {
		slog.Error("quiz_scheduler: failed to get pending open sessions", "error", err)
		return 0, 1
	}

	opened := 0
	errors := 0

	for _, session := range sessions {
		_, err := h.DB.Queries.UpdateQuizSessionState(ctx, sqlc.UpdateQuizSessionStateParams{
			ID:    session.ID,
			State: "OPEN",
		})
		if err != nil {
			slog.Error("quiz_scheduler: failed to open session",
				"session_id", session.ID,
				"error", err,
			)
			errors++
			continue
		}

		// Get quiz for project ID
		quiz, err := h.DB.Queries.GetQuizByID(ctx, session.QuizID)
		if err != nil {
			slog.Error("quiz_scheduler: failed to get quiz",
				"session_id", session.ID,
				"quiz_id", session.QuizID,
				"error", err,
			)
		} else if h.FirebaseService != nil {
			go h.FirebaseService.NotifyProjectQuizSessions(context.Background(), quiz.ProjectID)
		}

		slog.Info("quiz_scheduler: opened session",
			"session_id", session.ID,
			"quiz_id", session.QuizID,
		)
		opened++
	}

	return opened, errors
}

// processLockTransitions locks sessions that have reached their lock_at time
func (h *QuizSchedulerHandler) processLockTransitions(ctx context.Context) (int, int) {
	sessions, err := h.DB.Queries.GetSessionsPendingLock(ctx)
	if err != nil {
		slog.Error("quiz_scheduler: failed to get pending lock sessions", "error", err)
		return 0, 1
	}

	locked := 0
	errors := 0

	for _, session := range sessions {
		_, err := h.DB.Queries.UpdateQuizSessionState(ctx, sqlc.UpdateQuizSessionStateParams{
			ID:    session.ID,
			State: "LOCKED",
		})
		if err != nil {
			slog.Error("quiz_scheduler: failed to lock session",
				"session_id", session.ID,
				"error", err,
			)
			errors++
			continue
		}

		// Get quiz for project ID
		quiz, err := h.DB.Queries.GetQuizByID(ctx, session.QuizID)
		if err != nil {
			slog.Error("quiz_scheduler: failed to get quiz",
				"session_id", session.ID,
				"quiz_id", session.QuizID,
				"error", err,
			)
		} else if h.FirebaseService != nil {
			go h.FirebaseService.NotifyProjectQuizSessions(context.Background(), quiz.ProjectID)
		}

		slog.Info("quiz_scheduler: locked session",
			"session_id", session.ID,
			"quiz_id", session.QuizID,
		)
		locked++
	}

	return locked, errors
}

// processFinishTransitions finishes sessions that have reached their finish_at time
func (h *QuizSchedulerHandler) processFinishTransitions(ctx context.Context) (int, int) {
	sessions, err := h.DB.Queries.GetSessionsPendingFinish(ctx)
	if err != nil {
		slog.Error("quiz_scheduler: failed to get pending finish sessions", "error", err)
		return 0, 1
	}

	finished := 0
	errors := 0

	for _, session := range sessions {
		// Update state to FINISHED first to prevent duplicate processing on retry
		_, err := h.DB.Queries.UpdateQuizSessionState(ctx, sqlc.UpdateQuizSessionStateParams{
			ID:    session.ID,
			State: "FINISHED",
		})
		if err != nil {
			slog.Error("quiz_scheduler: failed to finish session",
				"session_id", session.ID,
				"error", err,
			)
			errors++
			continue
		}

		// Auto-submit active submissions
		err = h.DB.Queries.AutoSubmitSessionSubmissions(ctx, session.ID)
		if err != nil {
			// Log but don't fail - state is already FINISHED
			slog.Error("quiz_scheduler: failed to auto-submit submissions",
				"session_id", session.ID,
				"error", err,
			)
		}

		// Invalidate cache for affected users
		if h.Cache != nil {
			submissions, subErr := h.DB.Queries.GetSessionSubmissionsWithUserData(ctx, session.ID)
			if subErr != nil {
				slog.Warn("quiz_scheduler: failed to get submissions for cache invalidation",
					"session_id", session.ID,
					"error", subErr,
				)
			} else {
				for _, sub := range submissions {
					h.Cache.InvalidateUser(sub.UserID)
					h.Cache.InvalidateUserQuizSubmissions(sub.UserID)
					h.Cache.InvalidateQuizSubmission(sub.SubmissionID)
				}
			}
		}

		// Get quiz for project ID and webhook dispatch
		quiz, err := h.DB.Queries.GetQuizByID(ctx, session.QuizID)
		if err != nil {
			slog.Error("quiz_scheduler: failed to get quiz",
				"session_id", session.ID,
				"quiz_id", session.QuizID,
				"error", err,
			)
			errors++
			continue
		}

		// Invalidate project cache for leaderboard updates
		if h.Cache != nil {
			h.Cache.InvalidateProject(quiz.ProjectID)
		}

		// Notify Firestore
		if h.FirebaseService != nil {
			go h.FirebaseService.NotifyProjectQuizSessions(context.Background(), quiz.ProjectID)
		}

		// Dispatch webhook
		if h.WebhookService != nil {
			go h.dispatchFinishedWebhook(session, quiz)
		}

		slog.Info("quiz_scheduler: finished session",
			"session_id", session.ID,
			"quiz_id", session.QuizID,
		)
		finished++
	}

	return finished, errors
}

// dispatchFinishedWebhook builds and dispatches the webhook for a finished session
func (h *QuizSchedulerHandler) dispatchFinishedWebhook(session *sqlc.QuizSession, quiz *sqlc.GetQuizByIDRow) {
	ctx := context.Background()

	submissions, err := h.DB.Queries.GetSessionSubmissionsWithUserData(ctx, session.ID)
	if err != nil {
		slog.Error("quiz_scheduler: failed to get session submissions for webhook",
			"session_id", session.ID,
			"error", err,
		)
		return
	}

	results := make([]webhooks.QuizUserResult, 0, len(submissions))
	for _, sub := range submissions {
		var scorePercent *float64
		if sub.Score != nil && sub.MaxScore != nil && *sub.MaxScore > 0 {
			percent := float64(*sub.Score) / float64(*sub.MaxScore) * 100
			scorePercent = &percent
		}

		results = append(results, webhooks.QuizUserResult{
			UserID:        sub.UserID,
			MembersID:     sub.MembersID,
			ChurchID:      sub.ChurchID,
			Score:         sub.Score,
			MaxScore:      sub.MaxScore,
			ScorePercent:  scorePercent,
			Completed:     sub.Completed,
			AutoSubmitted: sub.AutoSubmitted,
		})
	}

	sessionName := ""
	if session.Name != nil {
		sessionName = *session.Name
	}

	h.WebhookService.DispatchQuizSessionFinished(ctx, quiz.ProjectID, webhooks.QuizSessionFinishedData{
		SessionID:   session.ID,
		SessionName: sessionName,
		QuizID:      quiz.ID,
		QuizName:    quiz.Name,
		ChallengeID: quiz.ChallengeID,
		FinishedAt:  time.Now().UTC(),
		Results:     results,
	})
}
