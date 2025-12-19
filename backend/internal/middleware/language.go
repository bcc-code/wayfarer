package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

// DefaultLanguage is the fallback language when no Accept-Language header is provided
// or when the requested language has no translation
const DefaultLanguage = "no"

// LanguageExtractor is a middleware that extracts the preferred language
// from the Accept-Language header and adds it to the Gin context
func LanguageExtractor() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := parseAcceptLanguage(c.GetHeader("Accept-Language"))
		c.Set("language", lang)
		c.Next()
	}
}

// parseAcceptLanguage parses the Accept-Language header and returns
// the base language code (e.g., "en" from "en-US")
func parseAcceptLanguage(header string) string {
	if header == "" {
		return DefaultLanguage
	}

	// Parse Accept-Language header using golang.org/x/text/language
	tags, _, err := language.ParseAcceptLanguage(header)
	if err != nil || len(tags) == 0 {
		return DefaultLanguage
	}

	// Get the base language from the first (highest priority) tag
	// e.g., "en" from "en-US", "no" from "nb-NO"
	base, _ := tags[0].Base()
	return base.String()
}

// GetLanguage retrieves the language from the context
// Returns DefaultLanguage if not set
func GetLanguage(ctx context.Context) string {
	if lang, ok := ctx.Value(LanguageKey).(string); ok && lang != "" {
		return lang
	}
	return DefaultLanguage
}
