package firebase

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/firestore"
)

// NotifyUserAchievements updates the timestamp for user achievements.
// Path: users/{userId}/notifications/achievements
func (s *Service) NotifyUserAchievements(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/notifications/achievements", userID))
}

// NotifyUserChallenges updates the timestamp for user challenge completions.
// Path: users/{userId}/notifications/challenges
func (s *Service) NotifyUserChallenges(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/notifications/challenges", userID))
}

// NotifyUserContent updates the timestamp for user content progress.
// Path: users/{userId}/notifications/content
func (s *Service) NotifyUserContent(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/notifications/content", userID))
}

// NotifyUserQuizzes updates the timestamp for user quiz submissions.
// Path: users/{userId}/notifications/quizzes
func (s *Service) NotifyUserQuizzes(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/notifications/quizzes", userID))
}

// NotifyUserProjects updates the timestamp for user project membership.
// Path: users/{userId}/notifications/projects
func (s *Service) NotifyUserProjects(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/notifications/projects", userID))
}

// NotifyProjectChallenges updates the timestamp for project challenges.
// Path: projects/{projectId}/notifications/challenges
func (s *Service) NotifyProjectChallenges(ctx context.Context, projectID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("projects/%s/notifications/challenges", projectID))
}

// NotifyProjectEvents updates the timestamp for project events.
// Path: projects/{projectId}/notifications/events
func (s *Service) NotifyProjectEvents(ctx context.Context, projectID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("projects/%s/notifications/events", projectID))
}

// NotifyProjectQuizSessions updates the timestamp for project quiz sessions.
// Path: projects/{projectId}/notifications/quiz_sessions
// Clients subscribe to this and filter locally for their accessible sessions.
func (s *Service) NotifyProjectQuizSessions(ctx context.Context, projectID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("projects/%s/notifications/quiz_sessions", projectID))
}

// NotifyAdminFeedback updates the timestamp for admin feedback notifications.
// Path: admin/feedback
func (s *Service) NotifyAdminFeedback(ctx context.Context) error {
	return s.updateTimestamp(ctx, "admin/feedback")
}

// updateTimestamp sets the updatedAt field to the current server timestamp.
// No-op if service or firestore client is nil.
func (s *Service) updateTimestamp(ctx context.Context, path string) error {
	if s == nil || s.firestoreClient == nil {
		slog.Debug("Firebase service not initialized, skipping notification", "path", path)
		return nil
	}

	slog.Debug("Updating Firestore timestamp", "path", path)

	_, err := s.firestoreClient.Doc(path).Set(ctx, map[string]any{
		"updatedAt": firestore.ServerTimestamp,
	})
	if err != nil {
		slog.Error("Failed to update Firestore timestamp", "path", path, "error", err)
		return fmt.Errorf("failed to update timestamp at %s: %w", path, err)
	}

	slog.Debug("Firestore timestamp updated", "path", path)
	return nil
}
