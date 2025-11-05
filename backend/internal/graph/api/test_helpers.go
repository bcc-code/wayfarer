package api

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/middleware"
)

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}

// createTestContext creates a context with a user ID set for testing
func createTestContext(userID string) context.Context {
	ctx := context.Background()
	if userID != "" {
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
	}
	return ctx
}
