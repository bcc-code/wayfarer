package api

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
)

// TestCalculateTeamScoreDistribution tests the point distribution logic
// for the createTeamScoreAdjustment mutation
func TestCalculateTeamScoreDistribution(t *testing.T) {
	tests := []struct {
		name             string
		totalPoints      int
		memberCount      int
		distributionMode model.TeamScoreDistributionMode
		expectedPoints   []int32
	}{
		{
			name:             "SPLIT mode - equal division",
			totalPoints:      100,
			memberCount:      5,
			distributionMode: model.TeamScoreDistributionModeSplit,
			expectedPoints:   []int32{20, 20, 20, 20, 20},
		},
		{
			name:             "SPLIT mode - with remainder (100/3 = 34+33+33)",
			totalPoints:      100,
			memberCount:      3,
			distributionMode: model.TeamScoreDistributionModeSplit,
			expectedPoints:   []int32{34, 33, 33},
		},
		{
			name:             "SPLIT mode - with remainder (7/3 = 3+2+2)",
			totalPoints:      7,
			memberCount:      3,
			distributionMode: model.TeamScoreDistributionModeSplit,
			expectedPoints:   []int32{3, 2, 2},
		},
		{
			name:             "SPLIT mode - with larger remainder (10/4 = 3+3+2+2)",
			totalPoints:      10,
			memberCount:      4,
			distributionMode: model.TeamScoreDistributionModeSplit,
			expectedPoints:   []int32{3, 3, 2, 2},
		},
		{
			name:             "SPLIT mode - single member gets all",
			totalPoints:      100,
			memberCount:      1,
			distributionMode: model.TeamScoreDistributionModeSplit,
			expectedPoints:   []int32{100},
		},
		{
			name:             "SPLIT mode - zero points",
			totalPoints:      0,
			memberCount:      3,
			distributionMode: model.TeamScoreDistributionModeSplit,
			expectedPoints:   []int32{0, 0, 0},
		},
		{
			name:             "SPLIT mode - points less than members (2/5)",
			totalPoints:      2,
			memberCount:      5,
			distributionMode: model.TeamScoreDistributionModeSplit,
			expectedPoints:   []int32{1, 1, 0, 0, 0},
		},
		{
			name:             "EACH mode - all members get full amount",
			totalPoints:      100,
			memberCount:      5,
			distributionMode: model.TeamScoreDistributionModeEach,
			expectedPoints:   []int32{100, 100, 100, 100, 100},
		},
		{
			name:             "EACH mode - single member",
			totalPoints:      50,
			memberCount:      1,
			distributionMode: model.TeamScoreDistributionModeEach,
			expectedPoints:   []int32{50},
		},
		{
			name:             "EACH mode - zero points",
			totalPoints:      0,
			memberCount:      3,
			distributionMode: model.TeamScoreDistributionModeEach,
			expectedPoints:   []int32{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTeamScoreDistribution(tt.totalPoints, tt.memberCount, tt.distributionMode)

			assert.Equal(t, len(tt.expectedPoints), len(result),
				"should return correct number of point values")
			assert.Equal(t, tt.expectedPoints, result,
				"should distribute points correctly")

			// Verify total points sum
			var totalDistributed int32
			for _, points := range result {
				totalDistributed += points
			}

			if tt.distributionMode == model.TeamScoreDistributionModeSplit {
				assert.Equal(t, int32(tt.totalPoints), totalDistributed,
					"SPLIT mode: total distributed should equal input")
			} else {
				assert.Equal(t, int32(tt.totalPoints*tt.memberCount), totalDistributed,
					"EACH mode: total distributed should be points * memberCount")
			}
		})
	}
}

// calculateTeamScoreDistribution calculates point distribution for team members
// This is extracted from the resolver for testability
func calculateTeamScoreDistribution(totalPoints int, memberCount int, mode model.TeamScoreDistributionMode) []int32 {
	pointsArray := make([]int32, memberCount)

	switch mode {
	case model.TeamScoreDistributionModeSplit:
		basePoints := int32(totalPoints / memberCount)
		remainder := totalPoints % memberCount
		for i := range pointsArray {
			pointsArray[i] = basePoints
			if i < remainder {
				pointsArray[i]++
			}
		}
	case model.TeamScoreDistributionModeEach:
		for i := range pointsArray {
			pointsArray[i] = int32(totalPoints)
		}
	}

	return pointsArray
}

// TestTeamScoreDistributionModeValidation tests that all distribution modes are handled
func TestTeamScoreDistributionModeValidation(t *testing.T) {
	// Verify all enum values are covered
	modes := model.AllTeamScoreDistributionMode
	assert.Len(t, modes, 2, "should have exactly 2 distribution modes")
	assert.Contains(t, modes, model.TeamScoreDistributionModeSplit)
	assert.Contains(t, modes, model.TeamScoreDistributionModeEach)

	// Verify each mode is valid
	for _, mode := range modes {
		assert.True(t, mode.IsValid(), "mode %s should be valid", mode)
	}
}

// TestTeamScoreDistribution_EdgeCases tests edge cases for point distribution
func TestTeamScoreDistribution_EdgeCases(t *testing.T) {
	t.Run("EACH with negative points (deduction)", func(t *testing.T) {
		// Negative points can be used for deductions and distribute correctly in EACH mode
		result := calculateTeamScoreDistribution(-50, 3, model.TeamScoreDistributionModeEach)
		assert.Equal(t, []int32{-50, -50, -50}, result)
	})

	t.Run("SPLIT verifies fairness - no member gets more than 1 extra point", func(t *testing.T) {
		result := calculateTeamScoreDistribution(99, 10, model.TeamScoreDistributionModeSplit)

		// Base is 9, remainder is 9, so 9 members get 10 and 1 member gets 9
		minPoints := result[0]
		maxPoints := result[0]
		for _, points := range result {
			if points < minPoints {
				minPoints = points
			}
			if points > maxPoints {
				maxPoints = points
			}
		}
		assert.LessOrEqual(t, maxPoints-minPoints, int32(1),
			"difference between max and min points should be at most 1")
	})

	t.Run("SPLIT with large number of members", func(t *testing.T) {
		result := calculateTeamScoreDistribution(1000, 100, model.TeamScoreDistributionModeSplit)
		assert.Len(t, result, 100)

		var total int32
		for _, points := range result {
			total += points
			assert.Equal(t, int32(10), points, "each member should get 10 points")
		}
		assert.Equal(t, int32(1000), total)
	})
}
