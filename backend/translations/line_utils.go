package translations

import "strings"

// splitLines splits a string into an array of lines.
// Empty strings return an empty slice.
// Normalizes CRLF and CR line endings to LF.
func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	// Normalize line endings: CRLF -> LF, CR -> LF
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}

// splitLinesPtr is like splitLines but handles nil pointers.
func splitLinesPtr(s *string) []string {
	if s == nil {
		return []string{}
	}
	return splitLines(*s)
}

// joinLines joins an array of lines back into a single string with newlines.
func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
