package pagination

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeLeaderboardConfigCursor(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC)

	tests := []struct {
		name      string
		createdAt time.Time
		id        string
		wantEmpty bool
	}{
		{"encode valid cursor", testTime, "LC01ARZ3NDEKTSV4RRFFQ69G5FAV", false},
		{"encode with empty ID returns empty", testTime, "", true},
		{"encode with zero time", time.Time{}, "LC001", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeLeaderboardConfigCursor(tt.createdAt, tt.id)
			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestDecodeLeaderboardConfigCursor(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC)

	tests := []struct {
		name          string
		cursor        string
		expectedTime  time.Time
		expectedID    string
		expectError   bool
		errorContains string
	}{
		{
			name:         "decode empty cursor",
			cursor:       "",
			expectedTime: time.Time{},
			expectedID:   "",
			expectError:  false,
		},
		{
			name:          "decode invalid base64",
			cursor:        "not-valid-base64!!!",
			expectError:   true,
			errorContains: "invalid cursor format",
		},
		{
			name:          "decode cursor without separator",
			cursor:        "bm9zZXBhcmF0b3I=", // "noseparator" in base64
			expectError:   true,
			errorContains: "invalid leaderboard config cursor format",
		},
		{
			name:          "decode cursor with empty ID part",
			cursor:        "MjAyNC0wNi0xNVQxMjozMDo0NVp8", // "2024-06-15T12:30:45Z|" in base64
			expectError:   true,
			errorContains: "cursor decoded to empty ID",
		},
		{
			name:         "decode valid cursor from round trip",
			cursor:       EncodeLeaderboardConfigCursor(testTime, "LC01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			expectedTime: testTime,
			expectedID:   "LC01ARZ3NDEKTSV4RRFFQ69G5FAV",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeLeaderboardConfigCursor(tt.cursor)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, result.ID)
				if !tt.expectedTime.IsZero() {
					assert.True(t, tt.expectedTime.Equal(result.CreatedAt))
				}
			}
		})
	}
}

func TestLeaderboardConfigCursorRoundTrip(t *testing.T) {
	createdAt := time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC)
	id := "LC01ARZ3NDEKTSV4RRFFQ69G5FAV"

	encoded := EncodeLeaderboardConfigCursor(createdAt, id)
	decoded, err := DecodeLeaderboardConfigCursor(encoded)

	require.NoError(t, err)
	assert.Equal(t, id, decoded.ID)
	assert.True(t, createdAt.Equal(decoded.CreatedAt))
}
