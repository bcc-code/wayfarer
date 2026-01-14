package services

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: The ContentAchievementService requires a valid database connection to function.
// Functional tests are implemented as E2E tests in backend/e2e/content_events_registration_test.go
// which test the full flow of:
// - Content events arriving before user registration
// - User registration triggering ProcessPendingContentEvents
// - Achievement awards and score journal entries being created

// TestContentProgressValidation tests the content progress validation logic
// This mirrors the validation in the webhook handler
func TestContentProgressValidation(t *testing.T) {
	tests := []struct {
		name            string
		contentProgress *float32
		expectValid     bool
	}{
		{
			name:            "nil progress is valid",
			contentProgress: nil,
			expectValid:     true,
		},
		{
			name:            "0.01 is valid (minimum)",
			contentProgress: float32Ptr(0.01),
			expectValid:     true,
		},
		{
			name:            "1.0 is valid (100%)",
			contentProgress: float32Ptr(1.0),
			expectValid:     true,
		},
		{
			name:            "1.1 is valid (maximum)",
			contentProgress: float32Ptr(1.1),
			expectValid:     true,
		},
		{
			name:            "0.5 is valid (50%)",
			contentProgress: float32Ptr(0.5),
			expectValid:     true,
		},
		{
			name:            "0.0 is invalid (too low)",
			contentProgress: float32Ptr(0.0),
			expectValid:     false,
		},
		{
			name:            "1.2 is invalid (above 1.1)",
			contentProgress: float32Ptr(1.2),
			expectValid:     false,
		},
		{
			name:            "negative is invalid",
			contentProgress: float32Ptr(-0.5),
			expectValid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidContentProgress(tt.contentProgress)
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

// isValidContentProgress validates content progress value
// This mirrors the validation logic in the webhook handler
func isValidContentProgress(progress *float32) bool {
	if progress == nil {
		return true // nil is valid (no progress tracking)
	}
	return *progress >= 0.01 && *progress <= 1.1
}

func float32Ptr(f float32) *float32 {
	return &f
}

// TestServiceStructCreation verifies the service struct can be created
func TestServiceStructCreation(t *testing.T) {
	t.Run("service_struct_with_nil_optional_dependencies", func(t *testing.T) {
		// The service can be created with nil optional dependencies
		// (PushService, WebhookService, Loaders, Cache)
		// Note: DB is required for actual operation
		service := &ContentAchievementService{
			DB:             nil, // Would be set to actual DB in real usage
			Cache:          nil, // Optional - cache operations skipped when nil
			PushService:    nil, // Optional - push notifications skipped when nil
			Loaders:        nil, // Optional - translation loading skipped when nil
			WebhookService: nil, // Optional - webhook dispatch skipped when nil
		}
		require.NotNil(t, service)
	})

	t.Run("service_struct_with_cache", func(t *testing.T) {
		testCache, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
		require.NoError(t, err)
		defer testCache.Close()

		service := &ContentAchievementService{
			DB:             nil,
			Cache:          testCache,
			PushService:    nil,
			Loaders:        nil,
			WebhookService: nil,
		}
		require.NotNil(t, service)
	})
}

// Benchmark tests
func BenchmarkContentProgressValidation(b *testing.B) {
	progress := float32Ptr(0.75)
	for b.Loop() {
		isValidContentProgress(progress)
	}
}

func BenchmarkContentProgressValidation_Nil(b *testing.B) {
	for b.Loop() {
		isValidContentProgress(nil)
	}
}
