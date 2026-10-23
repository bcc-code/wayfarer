package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/otel"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// quizSessionStateUpdater is the query transitionQuizSession needs.
type quizSessionStateUpdater interface {
	UpdateQuizSessionState(ctx context.Context, arg sqlc.UpdateQuizSessionStateParams) (*sqlc.QuizSession, error)
}

// transitionQuizSession moves a session from one state to another. It returns
// ok=false, without an error, when the session was no longer in from: a
// concurrent call (another admin, or the scheduler) made the transition, and
// only that call may act on it, e.g. settle bets or dispatch webhooks.
func transitionQuizSession(ctx context.Context, q quizSessionStateUpdater, id, from, to string) (*sqlc.QuizSession, bool, error) {
	row, err := q.UpdateQuizSessionState(ctx, sqlc.UpdateQuizSessionStateParams{
		ID:        id,
		Fromstate: from,
		State:     to,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return row, true, nil
}

// refinishQuizSession handles a finish request for a session that is already
// FINISHED: it retries auto-submission and bet settlement in case they failed
// the first time (both are no-ops once done) and returns the session. It does
// not notify clients or dispatch quiz_session_finished; the call that made the
// transition did that.
//
// A call that lost the finish race to a concurrent one runs this while the
// winner may still be settling. Each bet is still paid once (ClaimUnsettledBets
// locks and re-checks it), but the two runs can split a session's batches
// between them, so a user whose bets land in both gets two result
// notifications, each with part of their net result. This needs more than
// betSettleBatchSize bets in the session and is accepted.
func (r *Resolver) refinishQuizSession(ctx context.Context, span trace.Span, session *sqlc.QuizSession, quiz *model.Quiz) *model.QuizSession {
	r.closeSessionAndSettleBets(ctx, span, session.ID, quiz)
	r.invalidateSessionSubmissionCaches(ctx, session.ID, quiz.ProjectID)
	return convertQuizSessionToModel(session)
}

// SettleFinishedSession runs the settle step of finishing a session for a
// session another caller has already moved to FINISHED: it waits for in-flight
// answers, auto-submits open submissions and pays out bets, like
// finishQuizSession does. The quiz scheduler calls it after its own
// transition. Settlement failures are logged and traced, not returned (see
// closeSessionAndSettleBets); an error means the session or quiz could not be
// loaded.
func (r *Resolver) SettleFinishedSession(ctx context.Context, sessionID string) error {
	ctx, span := otel.StartSpan(ctx, "quiz_session.settle_finished",
		attribute.String("session.id", sessionID),
	)
	defer span.End()

	session, err := r.DB.Queries.GetQuizSession(ctx, sessionID)
	if err != nil {
		otel.RecordError(span, err)
		return fmt.Errorf("failed to load session: %w", err)
	}
	quiz, err := r.Loaders.QuizByIDLoader.Load(ctx, session.QuizID)()
	if err != nil {
		otel.RecordError(span, err)
		return fmt.Errorf("failed to load quiz: %w", err)
	}
	if quiz == nil {
		return fmt.Errorf("quiz %s not found", session.QuizID)
	}

	r.closeSessionAndSettleBets(ctx, span, sessionID, quiz)
	return nil
}
