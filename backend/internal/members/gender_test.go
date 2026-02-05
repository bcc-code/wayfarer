package members

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
		{name: "MALE uppercase", input: "MALE", expected: "MALE"},
		{name: "male lowercase", input: "male", expected: "MALE"},
		{name: "M uppercase", input: "M", expected: "MALE"},
		{name: "m lowercase", input: "m", expected: "MALE"},
		{name: "Male mixed case", input: "Male", expected: "MALE"},
		{name: "FEMALE uppercase", input: "FEMALE", expected: "FEMALE"},
		{name: "female lowercase", input: "female", expected: "FEMALE"},
		{name: "F uppercase", input: "F", expected: "FEMALE"},
		{name: "f lowercase", input: "f", expected: "FEMALE"},
		{name: "Female mixed case", input: "Female", expected: "FEMALE"},
		{name: "with spaces", input: "  male  ", expected: "MALE"},
		{name: "empty string", input: "", expected: "UNKNOWN"},
		{name: "unknown value", input: "unknown", expected: "UNKNOWN"},
		{name: "null string", input: "null", expected: "UNKNOWN"},
		{name: "other value", input: "other", expected: "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeGender(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
