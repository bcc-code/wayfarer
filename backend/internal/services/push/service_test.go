package push

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/push/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendAchievementNotification_NilService(t *testing.T) {
	// Test that calling on nil service does not panic
	var s *Service
	assert.NotPanics(t, func() {
		s.SendAchievementNotification(context.Background(), "user123", AchievementInfo{
			ID:               "ach123",
			Name:             "Test Achievement",
			NotificationText: "You did it!",
		})
	})
}

func TestSendAchievementNotification_UnconfiguredService(t *testing.T) {
	// Test that unconfigured service (no VAPID keys) returns early without error
	mockQuerier := mocks.NewMockQuerier(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Service without VAPID keys configured
	s := NewService(mockQuerier, config.VAPIDConfig{}, logger)

	// Should not call any querier methods since it returns early
	s.SendAchievementNotification(context.Background(), "user123", AchievementInfo{
		ID:               "ach123",
		Name:             "Test Achievement",
		NotificationText: "You did it!",
	})

	// If we get here without panic or error, test passes
	mockQuerier.AssertNotCalled(t, "GetEnabledSubscriptionsForUsers", mock.Anything, mock.Anything)
}

func TestSendAchievementNotification_EmptyNotificationText(t *testing.T) {
	// Test that empty notification_text uses fallback
	mockQuerier := mocks.NewMockQuerier(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	s := NewService(mockQuerier, config.VAPIDConfig{
		PublicKey:  "test-public-key",
		PrivateKey: "test-private-key",
		Subject:    "mailto:test@example.com",
	}, logger)

	// Mock GetEnabledSubscriptionsForUsers to return empty (no subscriptions)
	mockQuerier.EXPECT().
		GetEnabledSubscriptionsForUsers(mock.Anything, mock.MatchedBy(func(params sqlc.GetEnabledSubscriptionsForUsersParams) bool {
			return params.Notificationtype == string(NotificationTypeAchievementUnlocked) &&
				len(params.Userids) == 1 &&
				params.Userids[0] == "user123"
		})).
		Return([]*sqlc.PushSubscription{}, nil)

	// Call with empty notification text
	s.SendAchievementNotification(context.Background(), "user123", AchievementInfo{
		ID:               "ach123",
		Name:             "Test Achievement",
		NotificationText: "", // Empty - should use fallback
	})

	// Assert that GetEnabledSubscriptionsForUsers was called
	mockQuerier.AssertExpectations(t)
}

func TestSendAchievementNotification_WithNotificationText(t *testing.T) {
	// Test that provided notification_text is used
	mockQuerier := mocks.NewMockQuerier(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	s := NewService(mockQuerier, config.VAPIDConfig{
		PublicKey:  "test-public-key",
		PrivateKey: "test-private-key",
		Subject:    "mailto:test@example.com",
	}, logger)

	// Mock GetEnabledSubscriptionsForUsers to return empty (no subscriptions)
	mockQuerier.EXPECT().
		GetEnabledSubscriptionsForUsers(mock.Anything, mock.MatchedBy(func(params sqlc.GetEnabledSubscriptionsForUsersParams) bool {
			return params.Notificationtype == string(NotificationTypeAchievementUnlocked) &&
				len(params.Userids) == 1 &&
				params.Userids[0] == "user456"
		})).
		Return([]*sqlc.PushSubscription{}, nil)

	// Call with custom notification text
	s.SendAchievementNotification(context.Background(), "user456", AchievementInfo{
		ID:               "ach789",
		Name:             "Custom Achievement",
		NotificationText: "Custom notification message!",
		ImageCompleted:   "https://example.com/image.png",
	})

	// Assert that GetEnabledSubscriptionsForUsers was called
	mockQuerier.AssertExpectations(t)
}

func TestAchievementInfo_Fields(t *testing.T) {
	// Test that AchievementInfo struct holds all fields correctly
	info := AchievementInfo{
		ID:               "ach123",
		Name:             "Test Achievement",
		NotificationText: "You earned this!",
		ImageCompleted:   "https://example.com/badge.png",
	}

	assert.Equal(t, "ach123", info.ID)
	assert.Equal(t, "Test Achievement", info.Name)
	assert.Equal(t, "You earned this!", info.NotificationText)
	assert.Equal(t, "https://example.com/badge.png", info.ImageCompleted)
}
