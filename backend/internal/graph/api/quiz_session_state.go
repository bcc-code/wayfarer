package api

import (
	"context"
	"errors"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/jackc/pgx/v5"
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
func (r *Resolver) refinishQuizSession(ctx context.Context, span trace.Span, session *sqlc.QuizSession, quiz *model.Quiz) *model.QuizSession {
	r.closeSessionAndSettleBets(ctx, span, session.ID, quiz)
	r.invalidateSessionSubmissionCaches(ctx, session.ID, quiz.ProjectID)
	return convertQuizSessionToModel(session)
}
