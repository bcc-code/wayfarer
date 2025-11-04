package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeGender(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "MALE uppercase",
			input:    "MALE",
			expected: "MALE",
		},
		{
			name:     "male lowercase",
			input:    "male",
			expected: "MALE",
		},
		{
			name:     "M uppercase",
			input:    "M",
			expected: "MALE",
		},
		{
			name:     "m lowercase",
			input:    "m",
			expected: "MALE",
		},
		{
			name:     "FEMALE uppercase",
			input:    "FEMALE",
			expected: "FEMALE",
		},
		{
			name:     "female lowercase",
			input:    "female",
			expected: "FEMALE",
		},
		{
			name:     "F uppercase",
			input:    "F",
			expected: "FEMALE",
		},
		{
			name:     "f lowercase",
			input:    "f",
			expected: "FEMALE",
		},
		{
			name:     "Male mixed case",
			input:    "Male",
			expected: "MALE",
		},
		{
			name:     "Female mixed case",
			input:    "Female",
			expected: "FEMALE",
		},
		{
			name:     "with leading/trailing spaces",
			input:    "  male  ",
			expected: "MALE",
		},
		{
			name:     "empty string defaults to MALE",
			input:    "",
			expected: "MALE",
		},
		{
			name:     "unknown value defaults to MALE",
			input:    "unknown",
			expected: "MALE",
		},
		{
			name:     "null string defaults to MALE",
			input:    "null",
			expected: "MALE",
		},
		{
			name:     "other defaults to MALE",
			input:    "other",
			expected: "MALE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeGender(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
