package pagination

import (
	"testing"

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
