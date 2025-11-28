package utils

import (
	"bytes"

	"github.com/yuin/goldmark"
)

// RenderMarkdownToHTML converts markdown text to HTML
func RenderMarkdownToHTML(markdown string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
