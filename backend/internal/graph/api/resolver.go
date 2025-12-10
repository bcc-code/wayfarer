package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
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
