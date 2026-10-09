package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"sync"

	"github.com/bcc-media/wayfarer/i18n"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/otel"
	"github.com/bcc-media/wayfarer/internal/services/betting"
	"github.com/bcc-media/wayfarer/internal/services/push"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// sessionBetQuestionTypes are the question types whose bets are settled by the
// core when a session finishes. ORDERING bets are settled by the
// ladder_to_heaven plugin via the quiz_session_finished webhook.
var sessionBetQuestionTypes = []string{"PREDEFINED"}

// closeSessionAndSettleBets auto-submits a finished session's open submissions,
// then pays out its bets. Settlement only runs once auto-submission succeeded,
// so no answer can be added after the bets are loaded. Failures are logged and
// traced, not returned: the session is FINISHED either way, and calling
// finishQuizSession again retries both steps (each is a no-op once done).
func (r *Resolver) closeSessionAndSettleBets(ctx context.Context, span trace.Span, sessionID string, quiz *model.Quiz) {
	// Not cancelled with the request: a payout must not stop halfway
	ctx = context.WithoutCancel(ctx)

	// Let answer writes that started while the session was OPEN commit first.
	// Auto-submit alone skips completed submissions, whose answers can still be
	// updated; answer writes starting after this see the session FINISHED.
	if err := r.DB.Queries.WaitForSessionSubmissionLocks(ctx, sessionID); err != nil {
		otel.RecordError(span, err)
		slog.Error("failed to wait for in-flight answers, bets left unsettled",
			"session_id", sessionID,
			"error", err,
		)
		return
	}

	if err := r.DB.Queries.AutoSubmitSessionSubmissions(ctx, sessionID); err != nil {
		otel.RecordError(span, err)
		slog.Error("failed to auto-submit submissions, bets left unsettled",
			"session_id", sessionID,
			"error", err,
		)
		return
	}

	result, err := r.settleSessionBets(ctx, sessionID, quiz)
	if err != nil {
		otel.RecordError(span, err)
		slog.Error("failed to settle session bets",
			"session_id", sessionID,
			"error", err,
		)
		return
	}
	span.SetAttributes(
		attribute.Int("bets.settled", result.Settled),
		attribute.Int("bets.failed", result.Failed),
	)
	if result.Failed > 0 {
		otel.RecordError(span, fmt.Errorf("%d session bets failed to settle", result.Failed))
		slog.Error("some session bets failed to settle, finish the session again to retry",
			"session_id", sessionID,
			"failed", result.Failed,
		)
	}
}

// invalidateSessionSubmissionCaches invalidates the caches of every user and
// submission in a session, e.g. after its submissions were auto-submitted.
func (r *Resolver) invalidateSessionSubmissionCaches(ctx context.Context, sessionID, projectID string) {
	submissions, err := r.DB.Queries.GetSessionSubmissionsWithUserData(ctx, sessionID)
	if err != nil {
		slog.Warn("failed to get session submissions for cache invalidation",
			"session_id", sessionID,
			"error", err,
		)
	} else {
		for _, sub := range submissions {
			r.Cache.InvalidateUser(sub.UserID)
			r.Cache.InvalidateUserQuizSubmissions(sub.UserID)
			r.Cache.InvalidateQuizSubmission(sub.SubmissionID)
		}
	}
	r.Cache.InvalidateProject(projectID)
	r.Cache.InvalidateQuizSession(sessionID)
}

// settleSessionBetsResult summarizes a settlement run.
type settleSessionBetsResult struct {
	Settled int
	Failed  int
	// Voided counts settled bets without an answer, whose stake was returned
	Voided int
	// UserNetPoints is the net bet result per user who had a judged bet settled.
	UserNetPoints map[string]int
	// VoidedUsers are users who had a void bet settled (stake returned)
	VoidedUsers map[string]bool
}

// settleSessionBets pays out all unsettled bets in a finished session.
// Each bet is settled in its own transaction; a failing bet is logged and
// skipped so the others are still paid out.
func (r *Resolver) settleSessionBets(ctx context.Context, sessionID string, quiz *model.Quiz) (settleSessionBetsResult, error) {
	result := settleSessionBetsResult{UserNetPoints: make(map[string]int), VoidedUsers: make(map[string]bool)}

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
		multiplierCorrect, errCorrect := numericToMultiplier(row.BettingMultiplierCorrect)
		multiplierWrong, errWrong := numericToMultiplier(row.BettingMultiplierWrong)
		if err := errors.Join(errCorrect, errWrong); err != nil {
			result.Failed++
			slog.Error("invalid bet multiplier, bet left unsettled",
				"error", err, "session_id", sessionID, "response_id", row.ID, "question_id", row.QuestionID)
			continue
		}

		lang := r.userLanguage(ctx, row.UserID)
		in := betting.SettleInput{
			Bet: betting.Bet{
				ResponseID:        row.ID,
				UserID:            row.UserID,
				Amount:            derefInt32(row.BetAmount),
				IsCorrect:         row.IsCorrect,
				MultiplierCorrect: multiplierCorrect,
				MultiplierWrong:   multiplierWrong,
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
		submissions[row.SubmissionID] = true
		if res.Void {
			// No result to announce: the stake was simply returned
			result.Voided++
			result.VoidedUsers[row.UserID] = true
			continue
		}
		result.UserNetPoints[row.UserID] += res.NetPoints
	}

	if result.Settled > 0 {
		r.afterSessionBetsSettled(ctx, quiz, eventID, submissions, result.UserNetPoints, result.VoidedUsers)
	}

	slog.Info("settled session bets",
		"session_id", sessionID,
		"settled", result.Settled,
		"voided", result.Voided,
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
// results (users with only void bets are not notified). Notifications are sent
// in the background with bounded concurrency; ctx should not be cancelled with
// the request (use context.WithoutCancel).
func (r *Resolver) afterSessionBetsSettled(ctx context.Context, quiz *model.Quiz, eventID *string, submissions map[string]bool, userNetPoints map[string]int, voidedUsers map[string]bool) {
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
	for userID := range voidedUsers {
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

// numericToMultiplier converts a NUMERIC multiplier to hundredths without
// going through float. Nil when not set; an error if it is not a whole number
// of hundredths or out of range.
func numericToMultiplier(n pgtype.Numeric) (*int64, error) {
	if !n.Valid {
		return nil, nil
	}
	if n.NaN || n.InfinityModifier != pgtype.Finite || n.Int == nil {
		return nil, betting.ErrMultiplierOutOfRange
	}

	// value = Int * 10^Exp, so hundredths = Int * 10^(Exp+2)
	shift := int64(n.Exp) + 2
	hundredths := new(big.Int).Set(n.Int)
	if shift >= 0 {
		hundredths.Mul(hundredths, new(big.Int).Exp(big.NewInt(10), big.NewInt(shift), nil))
	} else {
		var rem big.Int
		hundredths.QuoRem(hundredths, new(big.Int).Exp(big.NewInt(10), big.NewInt(-shift), nil), &rem)
		if rem.Sign() != 0 {
			return nil, betting.ErrMultiplierPrecision
		}
	}

	if !hundredths.IsInt64() || hundredths.Int64() < 0 || hundredths.Int64() > betting.MaxMultiplier {
		return nil, betting.ErrMultiplierOutOfRange
	}
	v := hundredths.Int64()
	return &v, nil
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}
