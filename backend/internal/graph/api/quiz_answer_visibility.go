package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
)

// canRevealQuestionAnswers reports whether the current user may see which of a
// question's answers are the correct ones (QuizPredefinedAnswer.isCorrect,
// QuizOrderingItem.correctOrder).
//
// Users who can manage the project always may — the admin question editor needs
// it. For everyone else the answers stay hidden until the most recent session
// they have access to has FINISHED. A quiz with no sessions has nothing to gate
// on, so the answers are visible there.
func (r *Resolver) canRevealQuestionAnswers(ctx context.Context, questionID string) (bool, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok || userID == "" {
		return false, nil
	}

	question, err := r.DB.Queries.GetQuizQuestionByID(ctx, questionID)
	if err != nil {
		return false, fmt.Errorf("failed to load question: %w", err)
	}

	quiz, err := r.Loaders.QuizByIDLoader.Load(ctx, question.QuizID)()
	if err != nil {
		return false, fmt.Errorf("failed to load quiz: %w", err)
	}
	if quiz != nil && r.RoleService.CanManageProject(ctx, userID, quiz.ProjectID) {
		return true, nil
	}

	sessions, err := r.DB.Queries.GetUserAccessibleSessions(ctx, sqlc.GetUserAccessibleSessionsParams{
		Userid: userID,
		Quizid: question.QuizID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to load user sessions: %w", err)
	}
	if len(sessions) == 0 {
		return true, nil
	}

	// Sessions are ordered by created_at DESC, so the first one is the most recent.
	return sessions[0].State == string(model.QuizSessionStateFinished), nil
}

// responseVerdict returns a response's stored grading verdict, hidden while the
// submission's session is still OPEN.
//
// Returning it earlier would tell a user whether their own answer was right
// while they can still change it — and, for a betting question, while the bet is
// still live. A submission outside a session is not gated.
func (r *Resolver) responseVerdict(ctx context.Context, submissionID string, verdict *bool) (*bool, error) {
	if verdict == nil {
		return nil, nil
	}

	submission, err := r.Loaders.QuizSubmissionByIDLoader.Load(ctx, submissionID)()
	if err != nil {
		return nil, fmt.Errorf("failed to load submission: %w", err)
	}
	if submission == nil || submission.SessionID == nil || *submission.SessionID == "" {
		return verdict, nil
	}

	session, err := r.Loaders.QuizSessionByIDLoader.Load(ctx, *submission.SessionID)()
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}
	if session != nil && session.State == string(model.QuizSessionStateOpen) {
		return nil, nil
	}

	return verdict, nil
}
