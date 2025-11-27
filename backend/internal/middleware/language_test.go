package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "empty header returns default",
			header:   "",
			expected: "en",
		},
		{
			name:     "simple language code",
			header:   "no",
			expected: "no",
		},
		{
			name:     "language with region",
			header:   "en-US",
			expected: "en",
		},
		{
			name:     "language with region lowercase",
			header:   "nb-no",
			expected: "nb",
		},
		{
			name:     "multiple languages - first is used",
			header:   "de,en;q=0.9,fr;q=0.8",
			expected: "de",
		},
		{
			name:     "multiple languages with quality - highest used",
			header:   "en;q=0.8,de;q=0.9,fr;q=0.7",
			expected: "de",
		},
		{
			name:     "wildcard returns mul (multilingual)",
			header:   "*",
			expected: "mul",
		},
		{
			name:     "invalid header returns default",
			header:   "invalid-header-format",
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAcceptLanguage(tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLanguageExtractor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "extracts Norwegian",
			header:   "nb-NO,nb;q=0.9,no;q=0.8",
			expected: "nb",
		},
		{
			name:     "extracts German",
			header:   "de-DE,de;q=0.9,en;q=0.8",
			expected: "de",
		},
		{
			name:     "defaults to English when no header",
			header:   "",
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				c.Request.Header.Set("Accept-Language", tt.header)
			}

			middleware := LanguageExtractor()
			middleware(c)

			// Check that language was set in context
			lang, exists := c.Get("language")
			assert.True(t, exists)
			assert.Equal(t, tt.expected, lang)
		})
	}
}

func TestGetLanguage(t *testing.T) {
	tests := []struct {
		name     string
		ctxValue interface{}
		expected string
	}{
		{
			name:     "returns language from context",
			ctxValue: "de",
			expected: "de",
		},
		{
			name:     "returns default when not set",
			ctxValue: nil,
			expected: "en",
		},
		{
			name:     "returns default for empty string",
			ctxValue: "",
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.ctxValue != nil {
				ctx = context.WithValue(ctx, LanguageKey, tt.ctxValue)
			}

			result := GetLanguage(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}
