package common

import "encoding/json"

// TranslationData is the common transport format for translations
// between the database and Phrase TMS
type TranslationData struct {
	Language string          // Language code (e.g., "no", "en", "de")
	Value    json.RawMessage // JSON-encoded translation content
	ID       string          // Entity ULID with prefix (e.g., "PR...", "EV...")
}
