package services

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/graph-gophers/dataloader/v7"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestSyncUserLanguage_SameLanguage_NoUpdate(t *testing.T) {
	mockQueries := mocks.NewMockLanguageQuerier(t)
	mockUserLoader := mocks.NewMockUserLoader(t)
	cacheInstance := newTestCache()
	logger := newTestLogger()

	service := NewLanguageService(mockQueries, cacheInstance, mockUserLoader, logger)

	ctx := context.Background()
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	requestedLang := "en"

	// Mock: User has the same language already
	user := &model.User{
		ID:       userID,
		Language: "en", // Same as requested
	}

	// Create a thunk that returns the user
	thunk := dataloader.Thunk[*model.User](func() (*model.User, error) {
		return user, nil
	})
	mockUserLoader.On("Load", mock.Anything, userID).Return(thunk)

	// UpdateUserLanguage should NOT be called since languages match
	// No mock expectation set for UpdateUserLanguage

	service.SyncUserLanguage(ctx, userID, requestedLang)

	// Verify UpdateUserLanguage was never called
	mockQueries.AssertNotCalled(t, "UpdateUserLanguage", mock.Anything, mock.Anything)
}

func TestSyncUserLanguage_DifferentLanguage_Updates(t *testing.T) {
	mockQueries := mocks.NewMockLanguageQuerier(t)
	mockUserLoader := mocks.NewMockUserLoader(t)
	cacheInstance := newTestCache()
	logger := newTestLogger()

	service := NewLanguageService(mockQueries, cacheInstance, mockUserLoader, logger)

	ctx := context.Background()
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	requestedLang := "en"

	// Mock: User has a different language
	user := &model.User{
		ID:       userID,
		Language: "no", // Different from requested
	}

	thunk := dataloader.Thunk[*model.User](func() (*model.User, error) {
		return user, nil
	})
	mockUserLoader.On("Load", mock.Anything, userID).Return(thunk)

	// Expect UpdateUserLanguage to be called
	mockQueries.On("UpdateUserLanguage", mock.Anything, sqlc.UpdateUserLanguageParams{
		UserID:   userID,
		Language: requestedLang,
	}).Return(nil)

	service.SyncUserLanguage(ctx, userID, requestedLang)

	mockQueries.AssertExpectations(t)
}

func TestSyncUserLanguage_EmptyUserID_NoAction(t *testing.T) {
	mockQueries := mocks.NewMockLanguageQuerier(t)
	mockUserLoader := mocks.NewMockUserLoader(t)
	cacheInstance := newTestCache()
	logger := newTestLogger()

	service := NewLanguageService(mockQueries, cacheInstance, mockUserLoader, logger)

	ctx := context.Background()

	// Call with empty userID
	service.SyncUserLanguage(ctx, "", "en")

	// Nothing should be called
	mockUserLoader.AssertNotCalled(t, "Load", mock.Anything, mock.Anything)
	mockQueries.AssertNotCalled(t, "UpdateUserLanguage", mock.Anything, mock.Anything)
}

func TestSyncUserLanguage_EmptyLanguage_NoAction(t *testing.T) {
	mockQueries := mocks.NewMockLanguageQuerier(t)
	mockUserLoader := mocks.NewMockUserLoader(t)
	cacheInstance := newTestCache()
	logger := newTestLogger()

	service := NewLanguageService(mockQueries, cacheInstance, mockUserLoader, logger)

	ctx := context.Background()

	// Call with empty language
	service.SyncUserLanguage(ctx, "US01ARZ3NDEKTSV4RRFFQ69G5FAV", "")

	// Nothing should be called
	mockUserLoader.AssertNotCalled(t, "Load", mock.Anything, mock.Anything)
	mockQueries.AssertNotCalled(t, "UpdateUserLanguage", mock.Anything, mock.Anything)
}

func TestSyncUserLanguage_UserNotFound_NoUpdate(t *testing.T) {
	mockQueries := mocks.NewMockLanguageQuerier(t)
	mockUserLoader := mocks.NewMockUserLoader(t)
	cacheInstance := newTestCache()
	logger := newTestLogger()

	service := NewLanguageService(mockQueries, cacheInstance, mockUserLoader, logger)

	ctx := context.Background()
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Mock: User not found (nil)
	thunk := dataloader.Thunk[*model.User](func() (*model.User, error) {
		return nil, nil
	})
	mockUserLoader.On("Load", mock.Anything, userID).Return(thunk)

	service.SyncUserLanguage(ctx, userID, "en")

	// UpdateUserLanguage should NOT be called
	mockQueries.AssertNotCalled(t, "UpdateUserLanguage", mock.Anything, mock.Anything)
}

func TestSyncUserLanguage_CacheInvalidatedOnUpdate(t *testing.T) {
	mockQueries := mocks.NewMockLanguageQuerier(t)
	mockUserLoader := mocks.NewMockUserLoader(t)
	cacheInstance := newTestCache()
	logger := newTestLogger()

	service := NewLanguageService(mockQueries, cacheInstance, mockUserLoader, logger)

	ctx := context.Background()
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	requestedLang := "de"

	// Pre-populate cache with user data
	cacheKey := cache.UserKey(userID)
	cachedUser := &model.User{
		ID:       userID,
		Language: "no",
	}
	cacheInstance.Set(cacheKey, cachedUser)

	// Wait for Ristretto to process the Set (it's async)
	time.Sleep(10 * time.Millisecond)

	// Verify it's in cache
	_, found := cacheInstance.Get(cacheKey)
	assert.True(t, found, "User should be in cache before sync")

	// Mock user loader (returns different language)
	user := &model.User{
		ID:       userID,
		Language: "no", // Different from requested
	}
	thunk := dataloader.Thunk[*model.User](func() (*model.User, error) {
		return user, nil
	})
	mockUserLoader.On("Load", mock.Anything, userID).Return(thunk)

	// Mock update
	mockQueries.On("UpdateUserLanguage", mock.Anything, sqlc.UpdateUserLanguageParams{
		UserID:   userID,
		Language: requestedLang,
	}).Return(nil)

	service.SyncUserLanguage(ctx, userID, requestedLang)

	// Cache should be invalidated (user key should be gone)
	// Note: InvalidateUser deletes multiple related keys, the main user key should be gone
	_, found = cacheInstance.Get(cacheKey)
	assert.False(t, found, "User should be removed from cache after sync")
}
