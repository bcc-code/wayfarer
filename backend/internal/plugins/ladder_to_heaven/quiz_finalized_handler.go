package ladder_to_heaven

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/firebase"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/bcc-media/wayfarer/internal/webhook"
	"github.com/gin-gonic/gin"
)

// quizFinalizedHandler handles webhook requests for quiz session finished events.
// Despite the name (kept for compatibility), this processes the quiz_session_finished
// webhook which fires when an admin finishes a session, processing all users' betting.
type quizFinalizedHandler struct {
	db        *database.DB
	cache     *cache.CacheWithRegistry
	loaders   *loaders.Loaders
	secretKey string
	firebase  *firebase.Service
}

// sessionFinishedRequest matches the outbound WebhookPayload format for quiz_session_finished events
type sessionFinishedRequest struct {
	EventType string              `json:"event_type" binding:"required"`
	Timestamp time.Time           `json:"timestamp" binding:"required"`
	ProjectID string              `json:"project_id" binding:"required"`
	Data      sessionFinishedData `json:"data" binding:"required"`
}

// sessionFinishedData contains the quiz session finished data
type sessionFinishedData struct {
	SessionID   string              `json:"session_id" binding:"required"`
	SessionName string              `json:"session_name"`
	QuizID      string              `json:"quiz_id" binding:"required"`
	QuizName    string              `json:"quiz_name"`
	ChallengeID string              `json:"challenge_id"`
	FinishedAt  time.Time           `json:"finished_at"`
	Results     []sessionUserResult `json:"results"`
}

// sessionUserResult contains a single user's quiz result from the webhook
type sessionUserResult struct {
	UserID        string   `json:"user_id"`
	MembersID     string   `json:"members_id"`
	ChurchID      string   `json:"church_id"`
	Score         *int32   `json:"score"`
	MaxScore      *int32   `json:"max_score"`
	ScorePercent  *float64 `json:"score_percentage"`
	Completed     bool     `json:"completed"`
	AutoSubmitted bool     `json:"auto_submitted"`
}

const (
	// Source type for bet entries in score_journal
	sourceTypeBet = "BET"
	// Question type for ordering questions
	questionTypeOrdering = "ORDERING"
	// Expected number of items in ordering question for penalty calculation
	expectedOrderingItems = 4
)

// Betting multipliers based on correct positions
const (
	multiplierAllCorrect = 2.0
	multiplierTwoCorrect = 1.5
	multiplierOneCorrect = 1.25
	multiplierAllWrong   = 0.0 // All wrong: no winnings (stake is lost)
)

// handle processes incoming quiz session finished webhook requests.
// This is triggered when an admin finishes a quiz session, at which point
// we calculate and apply betting results for all users in the session.
func (h *quizFinalizedHandler) handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Read raw body for signature verification
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Warn("ladder_to_heaven: session_finished: failed to read request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Verify webhook signature if secret key is configured
	if h.secretKey != "" {
		signature := c.GetHeader("X-Webhook-Signature")
		if !webhook.VerifySignature(body, signature, h.secretKey) {
			slog.Warn("ladder_to_heaven: session_finished: invalid webhook signature")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Restore body for JSON binding
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	// Parse request body
	var req sessionFinishedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("ladder_to_heaven: session_finished: invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	slog.Info("ladder_to_heaven: session_finished: processing request",
		"session_id", req.Data.SessionID,
		"quiz_id", req.Data.QuizID,
		"user_count", len(req.Data.Results),
	)

	// Get all submissions for this session
	submissions, err := h.db.Queries.GetSubmissionsBySessionID(ctx, req.Data.SessionID)
	if err != nil {
		slog.Error("ladder_to_heaven: session_finished: failed to get submissions",
			"error", err, "session_id", req.Data.SessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get submissions"})
		return
	}

	if len(submissions) == 0 {
		slog.Info("ladder_to_heaven: session_finished: no submissions found",
			"session_id", req.Data.SessionID)
		c.JSON(http.StatusOK, gin.H{"message": "no submissions to process"})
		return
	}

	// Build submission ID to user ID map
	submissionIDs := make([]string, len(submissions))
	submissionUserMap := make(map[string]string) // submission_id -> user_id
	for i, sub := range submissions {
		submissionIDs[i] = sub.ID
		submissionUserMap[sub.ID] = sub.UserID
	}

	// Get all responses for all submissions
	responses, err := h.db.Queries.GetQuizResponsesBySubmissionIDs(ctx, submissionIDs)
	if err != nil {
		slog.Error("ladder_to_heaven: session_finished: failed to get responses",
			"error", err, "session_id", req.Data.SessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get responses"})
		return
	}

	if len(responses) == 0 {
		slog.Info("ladder_to_heaven: session_finished: no responses found",
			"session_id", req.Data.SessionID)
		c.JSON(http.StatusOK, gin.H{"message": "no responses to process"})
		return
	}

	// Get challenge to find event_id if quiz is linked to a challenge (using cache)
	var eventID *string
	if req.Data.ChallengeID != "" {
		eventID = h.getEventIDFromChallenge(ctx, req.Data.ChallengeID)
	}

	// Collect unique question IDs that need predefined answers
	questionIDsNeeded := make([]string, 0)
	for _, resp := range responses {
		if resp.QuestionType == questionTypeOrdering && resp.BettingEnabled {
			questionIDsNeeded = append(questionIDsNeeded, resp.QuestionID)
		}
	}

	// Load predefined answers using cache (batch load cache misses)
	questionAnswers := h.loadPredefinedAnswers(ctx, questionIDsNeeded)

	// Process each response
	processedCount := 0
	totalPointsAwarded := 0
	usersAffected := make(map[string]bool)

	for _, response := range responses {
		userID := submissionUserMap[response.SubmissionID]
		points, processed, err := h.processResponse(ctx, response, userID, req.ProjectID, eventID, questionAnswers)
		if err != nil {
			slog.Error("ladder_to_heaven: session_finished: failed to process response",
				"error", err, "response_id", response.ID)
			// Continue processing other responses
			continue
		}
		if processed {
			processedCount++
			totalPointsAwarded += points
			usersAffected[userID] = true
		}
	}

	// Invalidate caches if any responses were processed
	if processedCount > 0 {
		h.cache.InvalidateProject(req.ProjectID)
		if eventID != nil {
			h.cache.InvalidateEvent(*eventID)
		}

		// Invalidate cache for all affected users and notify via Firestore
		for userID := range usersAffected {
			h.cache.InvalidateUser(userID)
			if h.firebase != nil {
				go h.firebase.NotifyUserContent(context.Background(), userID)
			}
		}
	}

	slog.Info("ladder_to_heaven: session_finished: processed betting responses",
		"session_id", req.Data.SessionID,
		"processed_count", processedCount,
		"total_points_awarded", totalPointsAwarded,
		"users_affected", len(usersAffected))

	if processedCount > 0 {
		c.JSON(http.StatusCreated, gin.H{
			"message":              "processed betting responses",
			"responses_processed":  processedCount,
			"total_points_awarded": totalPointsAwarded,
			"users_affected":       len(usersAffected),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "no ordering questions with betting to process"})
	}
}

// getEventIDFromChallenge retrieves the event_id for a challenge using the dataloader.
func (h *quizFinalizedHandler) getEventIDFromChallenge(ctx context.Context, challengeID string) *string {
	thunk := h.loaders.ChallengeByIDLoader.Load(ctx, challengeID)
	challenge, err := thunk()
	if err != nil {
		slog.Warn("ladder_to_heaven: session_finished: failed to get challenge",
			"error", err, "challenge_id", challengeID)
		return nil
	}

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

// loadPredefinedAnswers loads predefined answers for the given question IDs using the dataloader.
// The dataloader handles caching and batching automatically.
func (h *quizFinalizedHandler) loadPredefinedAnswers(ctx context.Context, questionIDs []string) map[string][]*model.QuizPredefinedAnswer {
	result := make(map[string][]*model.QuizPredefinedAnswer)
	if len(questionIDs) == 0 {
		return result
	}

	// Deduplicate question IDs
	uniqueIDs := make(map[string]bool)
	for _, id := range questionIDs {
		uniqueIDs[id] = true
	}

	// Load all answers via dataloader - collect thunks first, then resolve
	// This allows the dataloader to batch all requests together
	type thunkResult struct {
		questionID string
		thunk      func() ([]*model.QuizPredefinedAnswer, error)
	}
	thunks := make([]thunkResult, 0, len(uniqueIDs))
	for questionID := range uniqueIDs {
		thunks = append(thunks, thunkResult{
			questionID: questionID,
			thunk:      h.loaders.QuizAnswersByQuestionLoader.Load(ctx, questionID),
		})
	}

	// Resolve all thunks (dataloader batches the underlying database calls)
	for _, t := range thunks {
		answers, err := t.thunk()
		if err != nil {
			slog.Warn("ladder_to_heaven: session_finished: failed to load predefined answers",
				"error", err, "question_id", t.questionID)
			continue
		}
		result[t.questionID] = answers
	}

	return result
}

// processResponse processes a single quiz response for betting points.
// Creates two score journal entries: stake (negative) and winnings (positive, if any).
// Returns the net points (winnings - stake), whether it was processed, and any error.
func (h *quizFinalizedHandler) processResponse(
	ctx context.Context,
	response *sqlc.GetQuizResponsesBySubmissionIDsRow,
	userID string,
	projectID string,
	eventID *string,
	questionAnswers map[string][]*model.QuizPredefinedAnswer,
) (int, bool, error) {
	// Skip if no bet was placed
	if response.BetAmount == nil || *response.BetAmount == 0 {
		return 0, false, nil
	}

	// Skip if not an ordering question or betting not enabled
	if response.QuestionType != questionTypeOrdering || !response.BettingEnabled {
		return 0, false, nil
	}

	// Check idempotency - skip if already processed
	if response.ScoreJournalID != nil && *response.ScoreJournalID != "" {
		slog.Debug("ladder_to_heaven: session_finished: response already processed",
			"response_id", response.ID, "journal_id", *response.ScoreJournalID)
		return 0, false, nil
	}

	// Parse the submitted order from json_response
	submittedOrder, err := parseOrderingResponse(response.JsonResponse)
	if err != nil {
		slog.Warn("ladder_to_heaven: session_finished: failed to parse ordering response",
			"error", err, "response_id", response.ID)
		return 0, false, nil
	}

	// Get the correct order from predefined answers
	correctAnswers, ok := questionAnswers[response.QuestionID]
	if !ok || len(correctAnswers) == 0 {
		slog.Warn("ladder_to_heaven: session_finished: no predefined answers found",
			"response_id", response.ID, "question_id", response.QuestionID)
		return 0, false, nil
	}

	// Count correct positions
	correctCount := countCorrectPositions(submittedOrder, correctAnswers)
	totalItems := len(correctAnswers)

	// Calculate points with two-entry system: stake (negative) + winnings (positive)
	betAmount := int(*response.BetAmount)
	multiplier := calculateBetMultiplier(correctCount, totalItems)
	winnings := int(float64(betAmount) * multiplier)
	netPoints := winnings - betAmount

	// Create stake entry (always - this is what the user "put in the pool")
	stakeJournalID := ulid.NewScoreJournalID()
	stakeReason := "Bet stake"
	_, err = h.db.Queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
		ID:         stakeJournalID,
		ProjectID:  projectID,
		UserID:     userID,
		EventID:    eventID,
		Points:     int32(-betAmount),
		SourceType: sourceTypeBet,
		SourceID:   &response.ID,
		Reason:     &stakeReason,
	})
	if err != nil {
		return 0, false, err
	}

	// Create winnings entry (only if positive - this is what the user "wins back")
	var resultJournalID string
	if winnings > 0 {
		winningsJournalID := ulid.NewScoreJournalID()
		winningsReason := "Bet winnings"
		_, err = h.db.Queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
			ID:         winningsJournalID,
			ProjectID:  projectID,
			UserID:     userID,
			EventID:    eventID,
			Points:     int32(winnings),
			SourceType: sourceTypeBet,
			SourceID:   &response.ID,
			Reason:     &winningsReason,
		})
		if err != nil {
			return 0, false, err
		}
		resultJournalID = winningsJournalID
	} else {
		resultJournalID = stakeJournalID
	}

	// Update response with net points earned and journal ID
	_, err = h.db.Queries.UpdateBetResultWithJournal(ctx, sqlc.UpdateBetResultWithJournalParams{
		ID:             response.ID,
		Pointsearned:   int32(netPoints),
		Scorejournalid: resultJournalID,
	})
	if err != nil {
		return 0, false, err
	}

	slog.Debug("ladder_to_heaven: session_finished: processed bet",
		"response_id", response.ID,
		"user_id", userID,
		"bet_amount", betAmount,
		"correct_positions", correctCount,
		"total_items", totalItems,
		"multiplier", multiplier,
		"winnings", winnings,
		"net_points", netPoints)

	return netPoints, true, nil
}

// parseOrderingResponse extracts the ordered item IDs from the JSON response.
// The json_response is expected to be an array of answer IDs in the user's submitted order.
func parseOrderingResponse(jsonResponse []byte) ([]string, error) {
	if len(jsonResponse) == 0 {
		return nil, nil
	}

	var order []string
	if err := json.Unmarshal(jsonResponse, &order); err != nil {
		return nil, err
	}
	return order, nil
}

// countCorrectPositions counts how many items are in the correct position.
func countCorrectPositions(submitted []string, correct []*model.QuizPredefinedAnswer) int {
	count := 0
	for i := 0; i < len(submitted) && i < len(correct); i++ {
		if submitted[i] == correct[i].ID {
			count++
		}
	}
	return count
}

// calculateBetMultiplier returns the winnings multiplier based on correct positions.
// The multiplier is applied to the bet amount to determine winnings (not net).
// Net points = (bet * multiplier) - bet = bet * (multiplier - 1)
// - All positions correct: 2.0x winnings (net: +bet)
// - 2 positions correct: 1.5x winnings (net: +0.5*bet)
// - 1 position correct: 1.25x winnings (net: +0.25*bet)
// - 0 positions correct (all 4 wrong): 0x winnings (net: -bet, stake lost)
// - Other cases (e.g., 3 correct out of 4): 0x winnings (net: -bet, stake lost)
func calculateBetMultiplier(correctCount, totalItems int) float64 {
	if totalItems == 0 {
		return 0
	}
	if correctCount == totalItems {
		return multiplierAllCorrect
	}
	// Penalty only applies when all items are wrong and there are exactly 4 items
	if correctCount == 0 && totalItems == expectedOrderingItems {
		return multiplierAllWrong
	}
	switch correctCount {
	case 2:
		return multiplierTwoCorrect
	case 1:
		return multiplierOneCorrect
	default:
		return 0
	}
}
