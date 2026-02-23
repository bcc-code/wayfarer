package i18n

import (
	"embed"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed *.json
var translationFiles embed.FS

// BetResultMessages contains translations for bet result notifications
type BetResultMessages struct {
	Title string `json:"title"`
	Won   string `json:"won"`
	Lost  string `json:"lost"`
}

// BetReasonMessages contains translations for bet score adjustment reasons
type BetReasonMessages struct {
	Stake    string `json:"stake"`
	Winnings string `json:"winnings"`
}

// LanguageTranslations holds all translations for a single language
type LanguageTranslations struct {
	BetResult BetResultMessages `json:"bet_result"`
	BetReason BetReasonMessages `json:"bet_reason"`
}

// translations maps language code to translations
var translations = make(map[string]LanguageTranslations)

func init() {
	entries, err := translationFiles.ReadDir(".")
	if err != nil {
		panic("failed to read i18n directory: " + err.Error())
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		lang := strings.TrimSuffix(entry.Name(), ".json")

		data, err := translationFiles.ReadFile(entry.Name())
		if err != nil {
			panic("failed to read translation file " + entry.Name() + ": " + err.Error())
		}

		var langTranslations LanguageTranslations
		if err := json.Unmarshal(data, &langTranslations); err != nil {
			panic("failed to parse translation file " + entry.Name() + ": " + err.Error())
		}

		translations[lang] = langTranslations
	}

	if _, ok := translations[DefaultLanguage]; !ok {
		panic("default language " + DefaultLanguage + " not found in translations")
	}
}

// SupportedLanguages returns a list of all supported language codes
func SupportedLanguages() []string {
	langs := make([]string, 0, len(translations))
	for lang := range translations {
		langs = append(langs, lang)
	}
	return langs
}

// AddLanguage allows adding translations at runtime (useful for testing)
func AddLanguage(lang string, trans LanguageTranslations) {
	translations[lang] = trans
}

// GetLanguageFile returns the expected filename for a language
func GetLanguageFile(lang string) string {
	return filepath.Join("i18n", lang+".json")
}

// DefaultLanguage is the fallback language
const DefaultLanguage = "nb"

// GetBetResultMessages returns the bet result messages for a given language
func GetBetResultMessages(lang string) BetResultMessages {
	// Normalize language code (handle variations like "en_US" -> "en")
	lang = normalizeLanguage(lang)

	if trans, ok := translations[lang]; ok {
		return trans.BetResult
	}
	// Fallback to default language
	return translations[DefaultLanguage].BetResult
}

// FormatBetResultMessage formats a bet result message with the points value.
// Returns empty strings if points is 0 (no notification should be sent).
func FormatBetResultMessage(lang string, points int) (title string, message string) {
	if points == 0 {
		return "", ""
	}

	msgs := GetBetResultMessages(lang)
	title = msgs.Title

	if points > 0 {
		message = replacePlaceholder(msgs.Won, "{points}", strconv.Itoa(points))
	} else {
		// Show absolute value for "lost" message
		absPoints := -points
		message = replacePlaceholder(msgs.Lost, "{points}", strconv.Itoa(absPoints))
	}

	return title, message
}

// GetBetReasonMessages returns the bet reason messages for a given language
func GetBetReasonMessages(lang string) BetReasonMessages {
	lang = normalizeLanguage(lang)

	if trans, ok := translations[lang]; ok {
		return trans.BetReason
	}
	return translations[DefaultLanguage].BetReason
}

// FormatBetStakeReason formats the stake reason with the challenge name
func FormatBetStakeReason(lang, challengeName string) string {
	msgs := GetBetReasonMessages(lang)
	return replacePlaceholder(msgs.Stake, "{challenge}", challengeName)
}

// FormatBetWinningsReason formats the winnings reason with the challenge name
func FormatBetWinningsReason(lang, challengeName string) string {
	msgs := GetBetReasonMessages(lang)
	return replacePlaceholder(msgs.Winnings, "{challenge}", challengeName)
}

// normalizeLanguage converts language codes to the format used in translations
func normalizeLanguage(lang string) string {
	// Convert to lowercase
	lang = strings.ToLower(lang)

	// Handle common variations
	switch lang {
	case "no", "nb", "nn":
		return "nb"
	case "en_us", "en-us", "en_gb", "en-gb":
		return "en"
	case "de_de", "de-de", "de_at", "de-at", "de_ch", "de-ch":
		return "de"
	case "pt_br", "pt-br", "pt_pt", "pt-pt":
		return "pt"
	case "zh_cn", "zh-cn", "zh_hans", "zh-hans":
		return "zh_cn"
	case "zh_tw", "zh-tw", "zh_hk", "zh-hk", "zh_hant", "zh-hant":
		return "zh_hk"
	}

	// Handle underscore/hyphen variations
	if idx := strings.IndexAny(lang, "_-"); idx != -1 {
		base := lang[:idx]
		// Check if we have a translation for the base language
		if _, ok := translations[base]; ok {
			return base
		}
	}

	return lang
}

func replacePlaceholder(template, placeholder, value string) string {
	return strings.ReplaceAll(template, placeholder, value)
}
