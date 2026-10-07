package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bcc-media/wayfarer/i18n"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services/betting"
	"github.com/bcc-media/wayfarer/internal/services/push"
	"github.com/jackc/pgx/v5/pgtype"
)

// sessionBetQuestionTypes are the question types whose bets are settled by the
// core when a session finishes. ORDERING bets are settled by the
// ladder_to_heaven plugin via the quiz_session_finished webhook.
var sessionBetQuestionTypes = []string{"PREDEFINED"}

// settleSessionBetsResult summarizes a settlement run.
type settleSessionBetsResult struct {
	Settled int
	Failed  int
	// UserNetPoints is the net bet result per user who had a bet settled.
	UserNetPoints map[string]int
}

// settleSessionBets pays out all unsettled bets in a finished session.
// Each bet is settled in its own transaction; a failing bet is logged and
// skipped so the others are still paid out.
func (r *Resolver) settleSessionBets(ctx context.Context, sessionID string, quiz *model.Quiz) (settleSessionBetsResult, error) {
	result := settleSessionBetsResult{UserNetPoints: make(map[string]int)}

	bets, err := r.DB.Queries.GetUnsettledSessionBets(ctx, sqlc.GetUnsettledSessionBetsParams{
		Sessionid:     sessionID,
		Questiontypes: sessionBetQuestionTypes,
	})
	if err != nil {
		return result, fmt.Errorf("failed to load session bets: %w", err)
	}
	if len(bets) == 0 {
		return result, nil
	}

	var eventID *string
	var challengeName string
	if quiz.ChallengeID != "" {
		eventID, challengeName = r.challengeEventAndName(ctx, quiz.ChallengeID)
	}
	if challengeName == "" {
		challengeName = quiz.Name
	}

	submissions := make(map[string]bool)
	for _, row := range bets {
		lang := r.userLanguage(ctx, row.UserID)
		in := betting.SettleInput{
			Bet: betting.Bet{
				ResponseID:        row.ID,
				UserID:            row.UserID,
				Amount:            derefInt32(row.BetAmount),
				IsCorrect:         row.IsCorrect,
				MultiplierCorrect: numericToFloat(row.BettingMultiplierCorrect),
				MultiplierWrong:   numericToFloat(row.BettingMultiplierWrong),
			},
			ProjectID:     quiz.ProjectID,
			EventID:       eventID,
			Language:      lang,
			ChallengeName: r.translatedChallengeName(ctx, quiz.ChallengeID, lang, challengeName),
		}

		res, err := r.settleBetInTx(ctx, in)
		if errors.Is(err, betting.ErrAlreadySettled) {
			continue // settled by a concurrent run; this run's writes were rolled back
		}
		if err != nil {
			result.Failed++
			slog.Error("failed to settle bet",
				"error", err, "session_id", sessionID, "response_id", row.ID, "user_id", row.UserID)
			continue
		}
		if !res.Settled {
			continue
		}

		result.Settled++
		result.UserNetPoints[row.UserID] += res.NetPoints
		submissions[row.SubmissionID] = true
	}

	if result.Settled > 0 {
		r.afterSessionBetsSettled(ctx, quiz, eventID, submissions, result.UserNetPoints)
	}

	slog.Info("settled session bets",
		"session_id", sessionID,
		"settled", result.Settled,
		"failed", result.Failed,
		"users", len(result.UserNetPoints))

	return result, nil
}

// settleBetInTx settles one bet so that its journal entries and response update
// are written together or not at all.
func (r *Resolver) settleBetInTx(ctx context.Context, in betting.SettleInput) (betting.Result, error) {
	tx, err := r.DB.Pool.Begin(ctx)
	if err != nil {
		return betting.Result{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	res, err := betting.Settle(ctx, r.DB.Queries.WithTx(tx), in)
	if err != nil {
		return betting.Result{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return betting.Result{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return res, nil
}

// betNotifyConcurrency bounds how many users are notified at once after a
// session's bets are settled, so large sessions don't burst Firestore or push.
const betNotifyConcurrency = 16

// afterSessionBetsSettled invalidates caches and notifies users about their
// results. Notifications are sent in the background with bounded concurrency;
// ctx should not be cancelled with the request (use context.WithoutCancel).
func (r *Resolver) afterSessionBetsSettled(ctx context.Context, quiz *model.Quiz, eventID *string, submissions map[string]bool, userNetPoints map[string]int) {
	r.Cache.InvalidateProject(quiz.ProjectID)
	if eventID != nil {
		r.Cache.InvalidateEvent(*eventID)
	}
	for submissionID := range submissions {
		r.Cache.InvalidateQuizSubmission(submissionID)
	}
	for userID := range userNetPoints {
		r.Cache.InvalidateUser(userID)
		r.Cache.InvalidateUserQuizSubmissions(userID)
	}

	sendPush := r.PushService != nil && quiz.ChallengeID != ""
	if r.FirebaseService == nil && !sendPush {
		return
	}

	go func() {
		var wg sync.WaitGroup
		sem := make(chan struct{}, betNotifyConcurrency)
		for userID, netPoints := range userNetPoints {
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()

				if r.FirebaseService != nil {
					if err := r.FirebaseService.NotifyUserContent(ctx, userID); err != nil {
						slog.WarnContext(ctx, "failed to notify user about bet result", "error", err, "user_id", userID)
					}
				}
				if sendPush {
					push.SendTranslatedBetResultNotificationCtx(ctx,
						r.PushService, r.Loaders, userID, quiz.ChallengeID, quiz.ID, quiz.Name, netPoints)
				}
			}()
		}
		wg.Wait()
	}()
}

// challengeEventAndName returns the event ID and name of a challenge.
func (r *Resolver) challengeEventAndName(ctx context.Context, challengeID string) (*string, string) {
	challenge, err := r.Loaders.ChallengeByIDLoader.Load(ctx, challengeID)()
	if err != nil {
		slog.Warn("failed to load challenge for bet settlement", "error", err, "challenge_id", challengeID)
		return nil, ""
	}

	switch c := challenge.(type) {
	case *model.SimpleChallenge:
		return c.EventID, c.Name
	case *model.QuizChallenge:
		return c.EventID, c.Name
	case *model.ExternalChallenge:
		return c.EventID, c.Name
	case *model.PluginChallenge:
		return c.EventID, c.Name
	}
	return nil, ""
}

// userLanguage returns the user's language, or the default language.
func (r *Resolver) userLanguage(ctx context.Context, userID string) string {
	user, err := r.Loaders.UserByIDLoader.Load(ctx, userID)()
	if err == nil && user != nil && user.Language != "" {
		return user.Language
	}
	return i18n.DefaultLanguage
}

// translatedChallengeName returns the challenge name in lang, or fallback.
func (r *Resolver) translatedChallengeName(ctx context.Context, challengeID, lang, fallback string) string {
	if challengeID == "" {
		return fallback
	}
	trans, err := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "challenge",
		EntityID:   challengeID,
		LangCode:   lang,
	})()
	if err == nil && trans != nil && trans.Name != nil && *trans.Name != "" {
		return *trans.Name
	}
	return fallback
}

func numericToFloat(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	val, err := n.Float64Value()
	if err != nil || !val.Valid {
		return nil
	}
	return &val.Float64
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}
