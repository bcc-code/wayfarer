package firebase

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

// NotifyUserAchievements updates the timestamp for user achievements.
// Path: users/{userId}/achievements
func (s *Service) NotifyUserAchievements(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/achievements", userID))
}

// NotifyUserChallenges updates the timestamp for user challenge completions.
// Path: users/{userId}/challenges
func (s *Service) NotifyUserChallenges(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/challenges", userID))
}

// NotifyUserContent updates the timestamp for user content progress.
// Path: users/{userId}/content
func (s *Service) NotifyUserContent(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/content", userID))
}

// NotifyUserQuizzes updates the timestamp for user quiz submissions.
// Path: users/{userId}/quizzes
func (s *Service) NotifyUserQuizzes(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/quizzes", userID))
}

// NotifyUserProjects updates the timestamp for user project membership.
// Path: users/{userId}/projects
func (s *Service) NotifyUserProjects(ctx context.Context, userID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("users/%s/projects", userID))
}

// NotifyProjectChallenges updates the timestamp for project challenges.
// Path: projects/{projectId}/challenges
func (s *Service) NotifyProjectChallenges(ctx context.Context, projectID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("projects/%s/challenges", projectID))
}

// NotifyProjectEvents updates the timestamp for project events.
// Path: projects/{projectId}/events
func (s *Service) NotifyProjectEvents(ctx context.Context, projectID string) error {
	return s.updateTimestamp(ctx, fmt.Sprintf("projects/%s/events", projectID))
}

// NotifyAdminFeedback updates the timestamp for admin feedback notifications.
// Path: admin/feedback (global, not project-scoped)
func (s *Service) NotifyAdminFeedback(ctx context.Context) error {
	return s.updateTimestamp(ctx, "admin/feedback")
}

// updateTimestamp sets the updatedAt field to the current server timestamp.
// No-op if service or firestore client is nil.
func (s *Service) updateTimestamp(ctx context.Context, path string) error {
	if s == nil || s.firestoreClient == nil {
		return nil
	}

	_, err := s.firestoreClient.Doc(path).Set(ctx, map[string]interface{}{
		"updatedAt": firestore.ServerTimestamp,
	})
	if err != nil {
		return fmt.Errorf("failed to update timestamp at %s: %w", path, err)
	}

	return nil
}
