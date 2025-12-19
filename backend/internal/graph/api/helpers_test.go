package api

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestFilterPersonLeaderboardEntries(t *testing.T) {
	// Create test entries with ranks 1-count
	createEntries := func(count int) []services.LeaderboardEntry {
		entries := make([]services.LeaderboardEntry, count)
		for i := 0; i < count; i++ {
			entries[i] = services.LeaderboardEntry{
				EntityID: "US" + string(rune('A'+i)),
				Name:     "User " + string(rune('A'+i)),
				Score:    1000 - i*10,
				Rank:     int64(i + 1),
			}
		}
		return entries
	}

	// Tests with totalCount=100 (limit=20) to test basic filtering behavior
	t.Run("with large totalCount (limit=20)", func(t *testing.T) {
		totalCount := 100 // This gives limit of 20

		tests := []struct {
			name              string
			entries           []services.LeaderboardEntry
			first             *int
			after             *string
			expectedCount     int
			expectedFirstVal  *int
			expectedNilResult bool
		}{
			{
				name:             "filters entries to only rank <= 20",
				entries:          createEntries(30),
				first:            nil,
				after:            nil,
				expectedCount:    20,
				expectedFirstVal: intPtr(20),
			},
			{
				name:             "caps first to 20 when nil",
				entries:          createEntries(10),
				first:            nil,
				after:            nil,
				expectedCount:    10,
				expectedFirstVal: intPtr(20),
			},
			{
				name:             "caps first to 20 when greater than 20",
				entries:          createEntries(15),
				first:            intPtr(50),
				after:            nil,
				expectedCount:    15,
				expectedFirstVal: intPtr(20),
			},
			{
				name:             "keeps first when <= 20",
				entries:          createEntries(15),
				first:            intPtr(10),
				after:            nil,
				expectedCount:    15,
				expectedFirstVal: intPtr(10),
			},
			{
				name:              "returns empty when after cursor >= 20",
				entries:           createEntries(30),
				first:             intPtr(10),
				after:             stringPtr("20"),
				expectedCount:     0,
				expectedNilResult: true,
			},
			{
				name:              "returns empty when after cursor > 20",
				entries:           createEntries(30),
				first:             intPtr(10),
				after:             stringPtr("25"),
				expectedCount:     0,
				expectedNilResult: true,
			},
			{
				name:             "adjusts first when after + first would exceed 20",
				entries:          createEntries(30),
				first:            intPtr(10),
				after:            stringPtr("15"),
				expectedCount:    20,        // All entries with rank <= 20 pass filter
				expectedFirstVal: intPtr(5), // Can only return 5 more (15+5=20)
			},
			{
				name:             "keeps entries when after cursor is within limit",
				entries:          createEntries(30),
				first:            intPtr(5),
				after:            stringPtr("5"),
				expectedCount:    20,        // All entries with rank <= 20 pass filter
				expectedFirstVal: intPtr(5), // 5+5=10, which is within 20
			},
			{
				name:             "empty after cursor treated as no cursor",
				entries:          createEntries(25),
				first:            intPtr(30),
				after:            stringPtr(""),
				expectedCount:    20,
				expectedFirstVal: intPtr(20),
			},
			{
				name:             "handles invalid cursor gracefully",
				entries:          createEntries(25),
				first:            intPtr(10),
				after:            stringPtr("invalid"),
				expectedCount:    20,
				expectedFirstVal: intPtr(10), // Keep original first since cursor parsing failed
			},
			{
				name:             "empty entries returns empty",
				entries:          []services.LeaderboardEntry{},
				first:            intPtr(10),
				after:            nil,
				expectedCount:    0,
				expectedFirstVal: intPtr(10),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := FilterPersonLeaderboardEntries(tt.entries, totalCount, tt.first, tt.after)

				if tt.expectedNilResult {
					assert.Nil(t, result.Entries)
				} else {
					assert.Equal(t, tt.expectedCount, len(result.Entries))

					// Verify all returned entries have rank <= 20
					maxLimit := CalculatePersonLeaderboardLimit(totalCount)
					for _, entry := range result.Entries {
						assert.LessOrEqual(t, entry.Rank, int64(maxLimit),
							"Entry with rank %d should not be included", entry.Rank)
					}
				}

				if tt.expectedFirstVal != nil {
					assert.NotNil(t, result.AdjustedFirst)
					assert.Equal(t, *tt.expectedFirstVal, *result.AdjustedFirst)
				}
			})
		}
	})

	// Test dynamic limits based on totalCount
	t.Run("dynamic limit based on totalCount", func(t *testing.T) {
		tests := []struct {
			name             string
			totalCount       int
			entries          []services.LeaderboardEntry
			expectedLimit    int
			expectedFirstVal *int
		}{
			{
				name:             "totalCount >= 50 gives limit of 20",
				totalCount:       50,
				entries:          createEntries(30),
				expectedLimit:    20,
				expectedFirstVal: intPtr(20),
			},
			{
				name:             "totalCount = 49 gives limit of 10",
				totalCount:       49,
				entries:          createEntries(30),
				expectedLimit:    10,
				expectedFirstVal: intPtr(10),
			},
			{
				name:             "totalCount = 20 gives limit of 10",
				totalCount:       20,
				entries:          createEntries(30),
				expectedLimit:    10,
				expectedFirstVal: intPtr(10),
			},
			{
				name:             "totalCount = 19 gives limit of 3",
				totalCount:       19,
				entries:          createEntries(30),
				expectedLimit:    3,
				expectedFirstVal: intPtr(3),
			},
			{
				name:             "totalCount = 1 gives limit of 3",
				totalCount:       1,
				entries:          createEntries(5),
				expectedLimit:    3,
				expectedFirstVal: intPtr(3),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := FilterPersonLeaderboardEntries(tt.entries, tt.totalCount, nil, nil)

				assert.LessOrEqual(t, len(result.Entries), tt.expectedLimit)

				// Verify all returned entries have rank <= limit
				for _, entry := range result.Entries {
					assert.LessOrEqual(t, entry.Rank, int64(tt.expectedLimit),
						"Entry with rank %d should not be included (limit: %d)", entry.Rank, tt.expectedLimit)
				}

				assert.NotNil(t, result.AdjustedFirst)
				assert.Equal(t, *tt.expectedFirstVal, *result.AdjustedFirst)
			})
		}
	})

	// Test pagination enforcement across different limits
	t.Run("pagination enforcement with dynamic limits", func(t *testing.T) {
		tests := []struct {
			name        string
			totalCount  int
			after       *string
			expectedNil bool
		}{
			{
				name:        "totalCount=15, after=3 returns empty (limit=3)",
				totalCount:  15,
				after:       stringPtr("3"),
				expectedNil: true,
			},
			{
				name:        "totalCount=30, after=10 returns empty (limit=10)",
				totalCount:  30,
				after:       stringPtr("10"),
				expectedNil: true,
			},
			{
				name:        "totalCount=100, after=20 returns empty (limit=20)",
				totalCount:  100,
				after:       stringPtr("20"),
				expectedNil: true,
			},
			{
				name:        "totalCount=15, after=2 returns entries (limit=3)",
				totalCount:  15,
				after:       stringPtr("2"),
				expectedNil: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				entries := createEntries(30)
				result := FilterPersonLeaderboardEntries(entries, tt.totalCount, intPtr(10), tt.after)

				if tt.expectedNil {
					assert.Nil(t, result.Entries)
				} else {
					assert.NotNil(t, result.Entries)
				}
			})
		}
	})
}

func TestParseRankCursor(t *testing.T) {
	tests := []struct {
		name        string
		cursor      string
		expected    int64
		expectError bool
	}{
		{
			name:        "valid rank",
			cursor:      "15",
			expected:    15,
			expectError: false,
		},
		{
			name:        "zero rank",
			cursor:      "0",
			expected:    0,
			expectError: false,
		},
		{
			name:        "large rank",
			cursor:      "1000",
			expected:    1000,
			expectError: false,
		},
		{
			name:        "invalid cursor",
			cursor:      "abc",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty cursor",
			cursor:      "",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRankCursor(tt.cursor)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCalculatePersonLeaderboardLimit(t *testing.T) {
	tests := []struct {
		name       string
		totalCount int
		expected   int
	}{
		{"totalCount >= 50 returns 20", 50, 20},
		{"totalCount = 100 returns 20", 100, 20},
		{"totalCount = 1000 returns 20", 1000, 20},
		{"totalCount = 49 returns 10", 49, 10},
		{"totalCount = 20 returns 10", 20, 10},
		{"totalCount = 35 returns 10", 35, 10},
		{"totalCount = 19 returns 3", 19, 3},
		{"totalCount = 10 returns 3", 10, 3},
		{"totalCount = 1 returns 3", 1, 3},
		{"totalCount = 0 returns 3", 0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePersonLeaderboardLimit(tt.totalCount)
			assert.Equal(t, tt.expected, result)
		})
	}
}
