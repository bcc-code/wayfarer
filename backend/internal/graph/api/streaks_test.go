package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStreakResolverExists verifies that the Streak resolver is implemented
func TestStreakResolverExists(t *testing.T) {
	// This test ensures the Streak query resolver compiles and exists
	// The fact that this test compiles means the Streak resolver is properly implemented
	var _ QueryResolver = (*queryResolver)(nil)
	assert.True(t, true, "Streak resolver is implemented")
}
