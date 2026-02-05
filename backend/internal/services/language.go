package services

import (
	"context"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// LanguageQuerier defines the database operations needed for language management
type LanguageQuerier interface {
	UpdateUserLanguage(ctx context.Context, arg sqlc.UpdateUserLanguageParams) error
}

// UserLoader defines the interface for loading users
type UserLoader interface {
	Load(ctx context.Context, key string) dataloader.Thunk[*model.User]
}

// LanguageService handles user language preference synchronization
type LanguageService struct {
	queries    LanguageQuerier
	cache      *cache.CacheWithRegistry
	userLoader UserLoader
	logger     *slog.Logger
}

// NewLanguageService creates a new language service
func NewLanguageService(queries LanguageQuerier, c *cache.CacheWithRegistry, userLoader UserLoader, logger *slog.Logger) *LanguageService {
	return &LanguageService{
		queries:    queries,
		cache:      c,
		userLoader: userLoader,
		logger:     logger,
	}
}

// SyncUserLanguage compares the requested language with stored language and updates if different.
// This is designed to be called asynchronously (fire-and-forget) to avoid blocking the request.
func (s *LanguageService) SyncUserLanguage(ctx context.Context, userID string, requestedLanguage string) {
	if userID == "" || requestedLanguage == "" {
		return
	}

	// Load user to get stored language
	thunk := s.userLoader.Load(ctx, userID)
	user, err := thunk()
	if err != nil {
		s.logger.Warn("failed to load user for language sync",
			"user_id", userID,
			"error", err,
		)
		return
	}

	if user == nil {
		s.logger.Warn("user not found for language sync",
			"user_id", userID,
		)
		return
	}

	// Compare and update if different
	if user.Language == requestedLanguage {
		return
	}

	if err := s.queries.UpdateUserLanguage(ctx, sqlc.UpdateUserLanguageParams{
		UserID:   userID,
		Language: requestedLanguage,
	}); err != nil {
		s.logger.Warn("failed to update user language",
			"user_id", userID,
			"old_language", user.Language,
			"new_language", requestedLanguage,
			"error", err,
		)
		return
	}

	s.logger.Debug("user language updated",
		"user_id", userID,
		"old_language", user.Language,
		"new_language", requestedLanguage,
	)

	// Invalidate user cache to refresh the language
	s.cache.InvalidateUser(userID)
}
