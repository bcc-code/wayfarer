package pagination

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "encode valid ULID",
			id:       "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			expected: "VVMwMUFSWjNOREVLVFNWNFJSRkZRNjlHNUZBVg==",
		},
		{
			name:     "encode empty string",
			id:       "",
			expected: "",
		},
		{
			name:     "encode short ID",
			id:       "US123",
			expected: "VVMxMjM=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeCursor(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	tests := []struct {
		name        string
		cursor      string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "decode valid cursor",
			cursor:      "VVMwMUFSWjNOREVLVFNWNFJSRkZRNjlHNUZBVg==",
			expected:    "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			expectError: false,
		},
		{
			name:        "decode empty cursor",
			cursor:      "",
			expected:    "",
			expectError: false,
		},
		{
			name:        "decode invalid base64",
			cursor:      "invalid!!!cursor",
			expected:    "",
			expectError: true,
			errorMsg:    "invalid cursor format",
		},
		{
			name:        "decode cursor that decodes to empty",
			cursor:      "",
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeCursor(tt.cursor)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCursorRoundTrip(t *testing.T) {
	testIDs := []string{
		"US01ARZ3NDEKTSV4RRFFQ69G5FAV",
		"US01H0G5N7VHKWT9ZQXCPD4YR5",
		"CH01ARZ3NDEKTSV4RRFFQ69G5FAV",
	}

	for _, id := range testIDs {
		t.Run(id, func(t *testing.T) {
			encoded := EncodeCursor(id)
			decoded, err := DecodeCursor(encoded)

			require.NoError(t, err)
			assert.Equal(t, id, decoded)
		})
	}
}

func TestEncodeChallengeCursor(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC)

	tests := []struct {
		name        string
		publishedAt time.Time
		id          string
		wantEmpty   bool
	}{
		{
			name:        "encode valid challenge cursor",
			publishedAt: testTime,
			id:          "CL01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantEmpty:   false,
		},
		{
			name:        "encode with empty ID returns empty",
			publishedAt: testTime,
			id:          "",
			wantEmpty:   true,
		},
		{
			name:        "encode with zero time",
			publishedAt: time.Time{},
			id:          "CL001",
			wantEmpty:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeChallengeCursor(tt.publishedAt, tt.id)
			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestDecodeChallengeCursor(t *testing.T) {
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
			errorContains: "invalid challenge cursor format",
		},
		{
			name:          "decode cursor with invalid timestamp",
			cursor:        "aW52YWxpZC10aW1lfENMMDAxMjM=", // "invalid-time|CL00123" in base64
			expectError:   true,
			errorContains: "invalid timestamp",
		},
		{
			name:          "decode cursor with empty ID part",
			cursor:        "MjAyNC0wNi0xNVQxMjozMDo0NVp8", // "2024-06-15T12:30:45Z|" in base64
			expectError:   true,
			errorContains: "cursor decoded to empty ID",
		},
		{
			name:         "decode valid cursor from round trip",
			cursor:       EncodeChallengeCursor(testTime, "CL01ARZ3NDEKTSV4RRFFQ69G5FAV"),
			expectedTime: testTime,
			expectedID:   "CL01ARZ3NDEKTSV4RRFFQ69G5FAV",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeChallengeCursor(tt.cursor)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, result.ID)
				if !tt.expectedTime.IsZero() {
					assert.True(t, tt.expectedTime.Equal(result.PublishedAt),
						"expected time %v, got %v", tt.expectedTime, result.PublishedAt)
				}
			}
		})
	}
}

func TestChallengeCursorRoundTrip(t *testing.T) {
	testCases := []struct {
		name        string
		publishedAt time.Time
		id          string
	}{
		{
			name:        "standard challenge cursor",
			publishedAt: time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC),
			id:          "CL01ARZ3NDEKTSV4RRFFQ69G5FAV",
		},
		{
			name:        "challenge cursor with nanoseconds",
			publishedAt: time.Date(2024, 1, 1, 0, 0, 0, 123456789, time.UTC),
			id:          "CL01H0G5N7VHKWT9ZQXCPD4YR5",
		},
		{
			name:        "challenge cursor with short ID",
			publishedAt: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			id:          "CL001",
		},
		{
			name:        "challenge cursor with future date",
			publishedAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			id:          "CL_FUTURE_TEST",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := EncodeChallengeCursor(tc.publishedAt, tc.id)
			require.NotEmpty(t, encoded, "encoded cursor should not be empty")

			decoded, err := DecodeChallengeCursor(encoded)
			require.NoError(t, err)

			assert.Equal(t, tc.id, decoded.ID)
			assert.True(t, tc.publishedAt.Equal(decoded.PublishedAt),
				"timestamps don't match: expected %v, got %v", tc.publishedAt, decoded.PublishedAt)
		})
	}
}
