package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
)

// canRevealQuestionAnswers reports whether the current user may see which of a
// question's answers are correct. Project managers always may; everyone else
// only once the most recent session they can access has FINISHED.
func (r *Resolver) canRevealQuestionAnswers(ctx context.Context, questionID string) (bool, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok || userID == "" {
		return false, nil
	}

	question, err := r.DB.Queries.GetQuizQuestionByID(ctx, questionID)
	if err != nil {
		return false, fmt.Errorf("failed to load question: %w", err)
	}

	return r.canRevealQuizAnswers(ctx, userID, question.QuizID, true)
}

// canRevealMapPinGeometry reports whether the current user may see a MAP_PIN
// question's correct position or area. As canRevealQuestionAnswers, except a
// user with no accessible session may not: teams play one after another, and a
// team whose session doesn't exist yet must not read the answer.
func (r *Resolver) canRevealMapPinGeometry(ctx context.Context, quizID string) (bool, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok || userID == "" {
		return false, nil
	}
	return r.canRevealQuizAnswers(ctx, userID, quizID, false)
}

// canRevealQuizAnswers reports whether userID may see a quiz's correct answers:
// a project manager always, anyone else once their most recent accessible
// session is FINISHED. withoutSession is the answer when they have none.
func (r *Resolver) canRevealQuizAnswers(ctx context.Context, userID, quizID string, withoutSession bool) (bool, error) {
	quiz, err := r.Loaders.QuizByIDLoader.Load(ctx, quizID)()
	if err != nil {
		return false, fmt.Errorf("failed to load quiz: %w", err)
	}
	if quiz != nil && r.RoleService.CanManageProject(ctx, userID, quiz.ProjectID) {
		return true, nil
	}

	sessions, err := r.DB.Queries.GetUserAccessibleSessions(ctx, sqlc.GetUserAccessibleSessionsParams{
		Userid: userID,
		Quizid: quizID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to load user sessions: %w", err)
	}
	if len(sessions) == 0 {
		return withoutSession, nil
	}

	// Ordered by created_at DESC.
	return sessions[0].State == string(model.QuizSessionStateFinished), nil
}

// responseVerdict returns a response's stored grading verdict, hidden while the
// session is OPEN so a user cannot learn the answer while they can still change it.
func (r *Resolver) responseVerdict(ctx context.Context, submissionID string, verdict *bool) (*bool, error) {
	if verdict == nil {
		return nil, nil
	}
	visible, err := r.responseResultVisible(ctx, submissionID)
	if err != nil || !visible {
		return nil, err
	}
	return verdict, nil
}

// responsePoints returns a response's stored points under the same rule as
// responseVerdict. Used where the points give the answer away (MAP_PIN).
func (r *Resolver) responsePoints(ctx context.Context, submissionID string, points *int) (*int, error) {
	if points == nil {
		return nil, nil
	}
	visible, err := r.responseResultVisible(ctx, submissionID)
	if err != nil || !visible {
		return nil, err
	}
	return points, nil
}

// responseResultVisible reports whether a submission's results may be shown:
// not while its session is OPEN, for anyone. A submission without a session
// (M2M) has nothing to hide them behind. A submission or session that can't be
// found hides them: the foreign keys rule that out, but this is access control.
func (r *Resolver) responseResultVisible(ctx context.Context, submissionID string) (bool, error) {
	submission, err := r.Loaders.QuizSubmissionByIDLoader.Load(ctx, submissionID)()
	if err != nil {
		return false, fmt.Errorf("failed to load submission: %w", err)
	}
	if submission == nil {
		return false, nil
	}
	if submission.SessionID == nil || *submission.SessionID == "" {
		return true, nil
	}

	session, err := r.Loaders.QuizSessionByIDLoader.Load(ctx, *submission.SessionID)()
	if err != nil {
		return false, fmt.Errorf("failed to load session: %w", err)
	}
	if session == nil {
		return false, nil
	}
	return session.State != string(model.QuizSessionStateOpen), nil
}
