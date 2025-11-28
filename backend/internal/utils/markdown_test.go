package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "empty string",
			markdown: "",
			expected: "",
		},
		{
			name:     "plain text",
			markdown: "Hello world",
			expected: "<p>Hello world</p>\n",
		},
		{
			name:     "heading",
			markdown: "# Title",
			expected: "<h1>Title</h1>\n",
		},
		{
			name:     "bold text",
			markdown: "This is **bold** text",
			expected: "<p>This is <strong>bold</strong> text</p>\n",
		},
		{
			name:     "italic text",
			markdown: "This is *italic* text",
			expected: "<p>This is <em>italic</em> text</p>\n",
		},
		{
			name:     "link",
			markdown: "[link text](https://example.com)",
			expected: "<p><a href=\"https://example.com\">link text</a></p>\n",
		},
		{
			name:     "unordered list",
			markdown: "- item 1\n- item 2\n- item 3",
			expected: "<ul>\n<li>item 1</li>\n<li>item 2</li>\n<li>item 3</li>\n</ul>\n",
		},
		{
			name:     "ordered list",
			markdown: "1. first\n2. second\n3. third",
			expected: "<ol>\n<li>first</li>\n<li>second</li>\n<li>third</li>\n</ol>\n",
		},
		{
			name:     "code block",
			markdown: "```\ncode here\n```",
			expected: "<pre><code>code here\n</code></pre>\n",
		},
		{
			name:     "inline code",
			markdown: "Use `inline code` here",
			expected: "<p>Use <code>inline code</code> here</p>\n",
		},
		{
			name:     "multiple paragraphs",
			markdown: "First paragraph.\n\nSecond paragraph.",
			expected: "<p>First paragraph.</p>\n<p>Second paragraph.</p>\n",
		},
		{
			name:     "blockquote",
			markdown: "> This is a quote",
			expected: "<blockquote>\n<p>This is a quote</p>\n</blockquote>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderMarkdownToHTML(tt.markdown)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
