package api

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestFilterPersonLeaderboardEntries(t *testing.T) {
	// Create test entries with ranks 1-30
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
			expectedCount:    20, // All entries with rank <= 20 pass filter
			expectedFirstVal: intPtr(5), // Can only return 5 more (15+5=20)
		},
		{
			name:             "keeps entries when after cursor is within limit",
			entries:          createEntries(30),
			first:            intPtr(5),
			after:            stringPtr("5"),
			expectedCount:    20, // All entries with rank <= 20 pass filter
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
			result := FilterPersonLeaderboardEntries(tt.entries, tt.first, tt.after)

			if tt.expectedNilResult {
				assert.Nil(t, result.Entries)
			} else {
				assert.Equal(t, tt.expectedCount, len(result.Entries))

				// Verify all returned entries have rank <= 20
				for _, entry := range result.Entries {
					assert.LessOrEqual(t, entry.Rank, int64(MaxPersonLeaderboardEntries),
						"Entry with rank %d should not be included", entry.Rank)
				}
			}

			if tt.expectedFirstVal != nil {
				assert.NotNil(t, result.AdjustedFirst)
				assert.Equal(t, *tt.expectedFirstVal, *result.AdjustedFirst)
			}
		})
	}
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

func TestMaxPersonLeaderboardEntries(t *testing.T) {
	// Verify the constant is set to 20
	assert.Equal(t, 20, MaxPersonLeaderboardEntries)
}
