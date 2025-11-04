package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		size     int
		expected [][]int
	}{
		{
			name:     "empty slice",
			slice:    []int{},
			size:     5,
			expected: nil,
		},
		{
			name:     "slice smaller than chunk size",
			slice:    []int{1, 2, 3},
			size:     5,
			expected: [][]int{{1, 2, 3}},
		},
		{
			name:     "slice exactly chunk size",
			slice:    []int{1, 2, 3, 4, 5},
			size:     5,
			expected: [][]int{{1, 2, 3, 4, 5}},
		},
		{
			name:     "slice larger than chunk size",
			slice:    []int{1, 2, 3, 4, 5, 6, 7, 8},
			size:     3,
			expected: [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8}},
		},
		{
			name:     "size is 1",
			slice:    []int{1, 2, 3},
			size:     1,
			expected: [][]int{{1}, {2}, {3}},
		},
		{
			name:     "size is 0",
			slice:    []int{1, 2, 3},
			size:     0,
			expected: nil,
		},
		{
			name:     "size is negative",
			slice:    []int{1, 2, 3},
			size:     -1,
			expected: nil,
		},
		{
			name:     "large slice with size 800 (Members API use case)",
			slice:    makeRange(1, 1600),
			size:     800,
			expected: [][]int{makeRange(1, 800), makeRange(801, 1600)},
		},
		{
			name:     "801 items with size 800",
			slice:    makeRange(1, 801),
			size:     800,
			expected: [][]int{makeRange(1, 800), {801}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Chunk(tt.slice, tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChunkWithStrings(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		size     int
		expected [][]string
	}{
		{
			name:     "string slice",
			slice:    []string{"a", "b", "c", "d", "e"},
			size:     2,
			expected: [][]string{{"a", "b"}, {"c", "d"}, {"e"}},
		},
		{
			name:     "empty string slice",
			slice:    []string{},
			size:     3,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Chunk(tt.slice, tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a range of integers
func makeRange(min, max int) []int {
	result := make([]int, max-min+1)
	for i := range result {
		result[i] = min + i
	}
	return result
}
