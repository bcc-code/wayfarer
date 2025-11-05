package pagination

import (
	"encoding/base64"
	"fmt"
)

// EncodeCursor encodes a user ID into a base64 cursor string
func EncodeCursor(id string) string {
	if id == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(id))
}

// DecodeCursor decodes a base64 cursor string back to a user ID
func DecodeCursor(cursor string) (string, error) {
	if cursor == "" {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", fmt.Errorf("invalid cursor format: %w", err)
	}

	id := string(decoded)
	if id == "" {
		return "", fmt.Errorf("cursor decoded to empty string")
	}

	return id, nil
}
