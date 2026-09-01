package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/firebase"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/bulk"
	"github.com/bcc-media/wayfarer/internal/services/email"
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
	EmailService       *email.Service
	UserSyncService    *services.UserSyncService
	BulkService        *bulk.Service
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

// LoadQuizByID loads a quiz by ID.
// Visibility is controlled by the challenge's published_at and session access.
func (r *Resolver) LoadQuizByID(ctx context.Context, quizID string) (*model.Quiz, error) {
	thunk := r.Loaders.QuizByIDLoader.Load(ctx, quizID)
	quiz, err := thunk()
	if err != nil {
		return nil, err
	}
	if quiz == nil {
		return nil, fmt.Errorf("quiz not found")
	}
	return quiz, nil
}

// LoadChallengeWithVisibility loads a challenge and enforces visibility rules.
// Admins can see all challenges via direct lookup. Non-admins can only see published challenges.
// For quiz challenges, session access grants visibility regardless of publish status.
func (r *Resolver) LoadChallengeWithVisibility(ctx context.Context, challengeID string) (model.Challenge, error) {
	thunk := r.Loaders.ChallengeByIDLoader.Load(ctx, challengeID)
	challenge, err := thunk()
	if err != nil {
		return nil, err
	}

	userID, _ := middleware.GetUserID(ctx)
	projectID := getChallengeProjectID(challenge)

	// Admins can see all challenges via direct lookup
	if userID != "" && r.RoleService.CanManageProject(ctx, userID, projectID) {
		return challenge, nil
	}

	// For quiz challenges, check session access first - session access grants visibility
	if _, ok := challenge.(*model.QuizChallenge); ok && userID != "" {
		quizThunk := r.Loaders.QuizByChallengeIDLoader.Load(ctx, challengeID)
		quiz, err := quizThunk()
		if err == nil && quiz != nil {
			accessibleQuizIDs, err := r.getUserAccessibleQuizIDs(ctx, userID, projectID, []string{quiz.ID})
			if _, hasAccess := accessibleQuizIDs[quiz.ID]; err == nil && hasAccess {
				return challenge, nil // Session access grants visibility
			}
		}
	}

	// Non-admins without session access can only see published challenges
	publishedAt := getChallengePublishedAt(challenge)
	if publishedAt == nil || publishedAt.Time.After(time.Now()) {
		return nil, fmt.Errorf("challenge not found")
	}

	// Check visibility: enrolled OR visible_at in past
	visibleAt := getChallengeVisibleAt(challenge)
	isVisible := visibleAt != nil && !visibleAt.Time.After(time.Now())

	if !isVisible && userID != "" {
		// Check if user is enrolled
		enrolled, err := r.DB.Queries.IsUserEnrolledInChallenge(ctx, sqlc.IsUserEnrolledInChallengeParams{
			Userid:      userID,
			Challengeid: getChallengeID(challenge),
		})
		if err != nil || !enrolled {
			return nil, fmt.Errorf("challenge not found")
		}
	} else if !isVisible {
		// No user ID and not visible = not found
		return nil, fmt.Errorf("challenge not found")
	}

	return challenge, nil
}
