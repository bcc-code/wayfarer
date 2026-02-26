package pagination

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
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

// ChallengeCursor represents a decoded challenge cursor with timestamp and ID
type ChallengeCursor struct {
	PublishedAt time.Time
	ID          string
}

// EncodeChallengeCursor encodes a timestamp and ID into a composite cursor string
// Format: base64(RFC3339_timestamp|id)
func EncodeChallengeCursor(publishedAt time.Time, id string) string {
	if id == "" {
		return ""
	}
	raw := publishedAt.Format(time.RFC3339Nano) + "|" + id
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// DecodeChallengeCursor decodes a composite cursor string back to timestamp and ID
func DecodeChallengeCursor(cursor string) (ChallengeCursor, error) {
	if cursor == "" {
		return ChallengeCursor{}, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return ChallengeCursor{}, fmt.Errorf("invalid cursor format: %w", err)
	}

	raw := string(decoded)
	parts := strings.SplitN(raw, "|", 2)
	if len(parts) != 2 {
		return ChallengeCursor{}, fmt.Errorf("invalid challenge cursor format: expected timestamp|id")
	}

	publishedAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return ChallengeCursor{}, fmt.Errorf("invalid timestamp in cursor: %w", err)
	}

	if parts[1] == "" {
		return ChallengeCursor{}, fmt.Errorf("cursor decoded to empty ID")
	}

	return ChallengeCursor{
		PublishedAt: publishedAt,
		ID:          parts[1],
	}, nil
}
