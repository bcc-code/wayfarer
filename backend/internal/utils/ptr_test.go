package utils

import "testing"

func TestInt32PtrToIntPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    *int32
		expected *int
	}{
		{
			name:     "nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "zero value",
			input:    int32Ptr(0),
			expected: intPtr(0),
		},
		{
			name:     "positive value",
			input:    int32Ptr(42),
			expected: intPtr(42),
		},
		{
			name:     "negative value",
			input:    int32Ptr(-100),
			expected: intPtr(-100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int32PtrToIntPtr(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", *result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected %v, got nil", *tt.expected)
				return
			}

			if *result != *tt.expected {
				t.Errorf("expected %v, got %v", *tt.expected, *result)
			}
		})
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}

func intPtr(i int) *int {
	return &i
}
