package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	// JoinCodeLength is the length of generated join codes
	JoinCodeLength = 6
	// JoinCodeChars are the allowed characters in join codes
	JoinCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ0123456789"
	// MaxRetries is the maximum number of retries for generating unique codes
	MaxRetries = 10
)

// JoinCodeChecker defines the interface for checking if a join code exists
type JoinCodeChecker interface {
	CheckJoinCodeExists(ctx context.Context, code string) (bool, error)
}

// GenerateJoinCode generates a random 6-character alphanumeric join code
func GenerateJoinCode() (string, error) {
	code := make([]byte, JoinCodeLength)
	charsLen := big.NewInt(int64(len(JoinCodeChars)))

	for i := 0; i < JoinCodeLength; i++ {
		num, err := rand.Int(rand.Reader, charsLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		code[i] = JoinCodeChars[num.Int64()]
	}

	return string(code), nil
}

// GenerateUniqueJoinCode generates a unique join code by checking against existing codes
// It will retry up to MaxRetries times if a collision occurs
func GenerateUniqueJoinCode(ctx context.Context, checker JoinCodeChecker) (string, error) {
	for attempt := 0; attempt < MaxRetries; attempt++ {
		code, err := GenerateJoinCode()
		if err != nil {
			return "", fmt.Errorf("failed to generate join code: %w", err)
		}

		exists, err := checker.CheckJoinCodeExists(ctx, code)
		if err != nil {
			return "", fmt.Errorf("failed to check join code existence: %w", err)
		}

		if !exists {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique join code after %d attempts", MaxRetries)
}
