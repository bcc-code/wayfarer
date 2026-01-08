package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/firebase"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/push"
	"github.com/bcc-media/wayfarer/internal/services/webhooks"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB                 *database.DB
	Loaders            *loaders.Loaders
	Cache              *cache.CacheWithRegistry
	RoleService        *services.RoleService
	LeaderboardService *services.LeaderboardService
	Settings           *services.SettingsService
	PushService        *push.Service
	WebhookService     *webhooks.Service
	FirebaseService    *firebase.Service
	InstanceID         string
}

// getUserChallengeCompletedAt returns the completion timestamp for the current user and challenge
func (r *Resolver) getUserChallengeCompletedAt(ctx context.Context, challengeID string) (*scalars.DateTime, error) {
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok || currentUserID == "" {
		return nil, nil
	}

	key := loaders.UserChallengeKey{UserID: currentUserID, ChallengeID: challengeID}
	thunk := r.Loaders.UserChallengeCompletionTimestampLoader.Load(ctx, key)
	ts, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load completion timestamp: %w", err)
	}
	if ts == nil {
		return nil, nil
	}
	return &scalars.DateTime{Time: *ts}, nil
}

// getUserChallengeEnrolledAt returns the enrollment timestamp for the current user and challenge
func (r *Resolver) getUserChallengeEnrolledAt(ctx context.Context, challengeID string) (*scalars.DateTime, error) {
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok || currentUserID == "" {
		return nil, nil
	}

	key := loaders.UserChallengeKey{UserID: currentUserID, ChallengeID: challengeID}
	thunk := r.Loaders.UserChallengeEnrollmentTimestampLoader.Load(ctx, key)
	ts, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load enrollment timestamp: %w", err)
	}
	if ts == nil {
		return nil, nil
	}
	return &scalars.DateTime{Time: *ts}, nil
}

// LoadQuizWithVisibility loads a quiz and enforces visibility rules for non-admins.
// Admins can see all quizzes, non-admins can only see published quizzes.
func (r *Resolver) LoadQuizWithVisibility(ctx context.Context, quizID string) (*model.Quiz, error) {
	thunk := r.Loaders.QuizByIDLoader.Load(ctx, quizID)
	quiz, err := thunk()
	if err != nil {
		return nil, err
	}
	if quiz == nil {
		return nil, fmt.Errorf("quiz not found")
	}

	// Admins can see all quizzes
	userID, _ := middleware.GetUserID(ctx)
	if userID != "" && r.RoleService.CanManageProject(ctx, userID, quiz.ProjectID) {
		return quiz, nil
	}

	// Non-admins can only see published quizzes
	if quiz.PublishedAt == nil || quiz.PublishedAt.Time.After(time.Now()) {
		return nil, fmt.Errorf("quiz not found")
	}

	return quiz, nil
}

// LoadChallengeWithVisibility loads a challenge and enforces visibility rules for non-admins.
// Admins can see all challenges, non-admins can only see published challenges.
func (r *Resolver) LoadChallengeWithVisibility(ctx context.Context, challengeID string) (model.Challenge, error) {
	thunk := r.Loaders.ChallengeByIDLoader.Load(ctx, challengeID)
	challenge, err := thunk()
	if err != nil {
		return nil, err
	}

	// Admins can see all challenges
	userID, _ := middleware.GetUserID(ctx)
	projectID := getChallengeProjectID(challenge)
	if userID != "" && r.RoleService.CanManageProject(ctx, userID, projectID) {
		return challenge, nil
	}

	// Non-admins can only see published challenges
	publishedAt := getChallengePublishedAt(challenge)
	if publishedAt == nil || publishedAt.Time.After(time.Now()) {
		return nil, fmt.Errorf("challenge not found")
	}

	return challenge, nil
}
