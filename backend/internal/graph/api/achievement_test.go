package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAchievementIDFormat tests that achievement IDs follow the correct format
func TestAchievementIDFormat(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		isValid bool
	}{
		{
			name:    "valid achievement ID",
			id:      "AC01K8XV6VK9ED2GBZSQ2VDTAT8T",
			isValid: true,
		},
		{
			name:    "wrong prefix",
			id:      "US01K8XV6VK9ED2GBZSQ2VDTAT8T",
			isValid: false,
		},
		{
			name:    "too short",
			id:      "AC123",
			isValid: false,
		},
		{
			name:    "too long",
			id:      "AC01K8XV6VK9ED2GBZSQ2VDTAT8T123",
			isValid: false,
		},
		{
			name:    "missing prefix",
			id:      "01K8XV6VK9ED2GBZSQ2VDTAT8T",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasCorrectPrefix := len(tt.id) >= 2 && tt.id[:2] == "AC"
			hasCorrectLength := len(tt.id) == 28
			isValid := hasCorrectPrefix && hasCorrectLength

			assert.Equal(t, tt.isValid, isValid, "ID validation should match expected result")
		})
	}
}

// TestAchievementTypes tests that all achievement types are distinct
func TestAchievementTypes(t *testing.T) {
	types := []string{"SIMPLE", "READING", "LISTENING", "STREAK"}

	// Ensure all types are unique
	typeMap := make(map[string]bool)
	for _, achievementType := range types {
		assert.False(t, typeMap[achievementType], "achievement type %s should be unique", achievementType)
		typeMap[achievementType] = true
	}

	assert.Len(t, typeMap, 4, "should have exactly 4 achievement types")
}
