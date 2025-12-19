package translations

import (
	"testing"
)

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single line no newline",
			input:    "hello world",
			expected: []string{"hello world"},
		},
		{
			name:     "two lines",
			input:    "line1\nline2",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "three lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty line in middle",
			input:    "line1\n\nline3",
			expected: []string{"line1", "", "line3"},
		},
		{
			name:     "multiple empty lines in middle",
			input:    "line1\n\n\nline4",
			expected: []string{"line1", "", "", "line4"},
		},
		{
			name:     "trailing newline",
			input:    "line1\nline2\n",
			expected: []string{"line1", "line2", ""},
		},
		{
			name:     "leading newline",
			input:    "\nline2\nline3",
			expected: []string{"", "line2", "line3"},
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			expected: []string{"", "", "", ""},
		},
		{
			name:     "single newline",
			input:    "\n",
			expected: []string{"", ""},
		},
		{
			name:     "markdown with headers and paragraphs",
			input:    "# Title\n\nParagraph 1\n\nParagraph 2",
			expected: []string{"# Title", "", "Paragraph 1", "", "Paragraph 2"},
		},
		{
			name:     "markdown list",
			input:    "Items:\n- Item 1\n- Item 2\n- Item 3",
			expected: []string{"Items:", "- Item 1", "- Item 2", "- Item 3"},
		},
		{
			name:     "whitespace only lines",
			input:    "line1\n   \nline3",
			expected: []string{"line1", "   ", "line3"},
		},
		{
			name:     "tabs in content",
			input:    "line1\n\tindented\nline3",
			expected: []string{"line1", "\tindented", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitLines(%q) = %v (len %d), want %v (len %d)",
					tt.input, result, len(result), tt.expected, len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitLines(%q)[%d] = %q, want %q",
						tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestJoinLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: "",
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: "",
		},
		{
			name:     "single line",
			input:    []string{"hello world"},
			expected: "hello world",
		},
		{
			name:     "two lines",
			input:    []string{"line1", "line2"},
			expected: "line1\nline2",
		},
		{
			name:     "empty line in middle",
			input:    []string{"line1", "", "line3"},
			expected: "line1\n\nline3",
		},
		{
			name:     "multiple empty lines",
			input:    []string{"line1", "", "", "line4"},
			expected: "line1\n\n\nline4",
		},
		{
			name:     "trailing empty string",
			input:    []string{"line1", "line2", ""},
			expected: "line1\nline2\n",
		},
		{
			name:     "leading empty string",
			input:    []string{"", "line2", "line3"},
			expected: "\nline2\nline3",
		},
		{
			name:     "only empty strings",
			input:    []string{"", "", "", ""},
			expected: "\n\n\n",
		},
		{
			name:     "single empty string",
			input:    []string{""},
			expected: "",
		},
		{
			name:     "two empty strings",
			input:    []string{"", ""},
			expected: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinLines(tt.input)
			if result != tt.expected {
				t.Errorf("joinLines(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitJoinRoundTrip(t *testing.T) {
	// Test that splitting and joining produces the original string
	testCases := []string{
		"",
		"single line",
		"line1\nline2",
		"line1\n\nline3",
		"# Title\n\nParagraph",
		"\n",
		"\n\n",
		"trailing\n",
		"\nleading",
		"complex\n\nmulti\n\n\nline\ncontent\n",
	}

	for _, original := range testCases {
		t.Run(original, func(t *testing.T) {
			lines := splitLines(original)
			rejoined := joinLines(lines)
			if rejoined != original {
				t.Errorf("Round trip failed:\n  original: %q\n  split:    %v\n  rejoined: %q",
					original, lines, rejoined)
			}
		})
	}
}

func TestSplitLinesPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected []string
	}{
		{
			name:     "nil pointer",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "empty string pointer",
			input:    strPtr(""),
			expected: []string{},
		},
		{
			name:     "normal string pointer",
			input:    strPtr("line1\nline2"),
			expected: []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLinesPtr(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitLinesPtr(%v) = %v (len %d), want %v (len %d)",
					ptrStr(tt.input), result, len(result), tt.expected, len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitLinesPtr(%v)[%d] = %q, want %q",
						ptrStr(tt.input), i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// Helper to create string pointer
func strPtr(s string) *string {
	return &s
}

// Helper to safely dereference string pointer for error messages
func ptrStr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func TestCRLFHandling(t *testing.T) {
	// Test that CRLF (Windows line endings) are handled
	// We normalize to LF during split
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "CRLF line endings",
			input:    "line1\r\nline2\r\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "mixed line endings",
			input:    "line1\r\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "CR only",
			input:    "line1\rline2",
			expected: []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitLines(%q) = %v (len %d), want %v (len %d)",
					tt.input, result, len(result), tt.expected, len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitLines(%q)[%d] = %q, want %q",
						tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

