package utils

import (
	"context"
	"errors"
	"testing"
)

func TestGenerateJoinCode(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Generate join code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateJoinCode()
			if err != nil {
				t.Errorf("GenerateJoinCode() error = %v", err)
				return
			}

			// Check length
			if len(code) != JoinCodeLength {
				t.Errorf("GenerateJoinCode() length = %d, want %d", len(code), JoinCodeLength)
			}

			// Check that all characters are valid
			for _, char := range code {
				found := false
				for _, validChar := range JoinCodeChars {
					if char == validChar {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GenerateJoinCode() contains invalid character: %c", char)
				}
			}
		})
	}
}

func TestGenerateJoinCode_Uniqueness(t *testing.T) {
	// Generate multiple codes and check they're different
	// This is a probabilistic test - there's a tiny chance of collision
	codes := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		code, err := GenerateJoinCode()
		if err != nil {
			t.Fatalf("GenerateJoinCode() error = %v", err)
		}
		codes[code] = true
	}

	// With 100 iterations and 36^6 possible codes, we expect no collisions
	// If we get significantly fewer unique codes, something is wrong
	if len(codes) < iterations*95/100 { // Allow 5% collision rate
		t.Errorf("GenerateJoinCode() generated too few unique codes: %d out of %d", len(codes), iterations)
	}
}

type mockJoinCodeChecker struct {
	existingCodes map[string]bool
	checkError    error
}

func (m *mockJoinCodeChecker) CheckJoinCodeExists(ctx context.Context, code string) (bool, error) {
	if m.checkError != nil {
		return false, m.checkError
	}
	return m.existingCodes[code], nil
}

func TestGenerateUniqueJoinCode(t *testing.T) {
	tests := []struct {
		name          string
		existingCodes map[string]bool
		checkError    error
		wantErr       bool
	}{
		{
			name:          "Success - no existing codes",
			existingCodes: map[string]bool{},
			checkError:    nil,
			wantErr:       false,
		},
		{
			name: "Success - some existing codes",
			existingCodes: map[string]bool{
				"ABC123": true,
				"XYZ789": true,
			},
			checkError: nil,
			wantErr:    false,
		},
		{
			name:          "Error - checker error",
			existingCodes: map[string]bool{},
			checkError:    errors.New("database error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &mockJoinCodeChecker{
				existingCodes: tt.existingCodes,
				checkError:    tt.checkError,
			}

			code, err := GenerateUniqueJoinCode(context.Background(), checker)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateUniqueJoinCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check that the generated code is valid
				if len(code) != JoinCodeLength {
					t.Errorf("GenerateUniqueJoinCode() length = %d, want %d", len(code), JoinCodeLength)
				}

				// Check that the code doesn't exist in the existing codes
				if checker.existingCodes[code] {
					t.Errorf("GenerateUniqueJoinCode() generated existing code: %s", code)
				}
			}
		})
	}
}

func TestGenerateUniqueJoinCode_MaxRetries(t *testing.T) {
	// Create a checker that always returns true (all codes exist)
	// This should cause the function to fail after MaxRetries attempts
	alwaysExistsChecker := &alwaysExistsChecker{}

	_, err := GenerateUniqueJoinCode(context.Background(), alwaysExistsChecker)
	if err == nil {
		t.Error("GenerateUniqueJoinCode() expected error when all codes exist, got nil")
	}
}

type alwaysExistsChecker struct{}

func (a *alwaysExistsChecker) CheckJoinCodeExists(ctx context.Context, code string) (bool, error) {
	return true, nil
}
