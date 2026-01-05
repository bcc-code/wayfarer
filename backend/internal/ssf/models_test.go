package ssf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItem_ExtractContentData_MediaEpisode(t *testing.T) {
	item := &Item{
		PlanChapterItemID: "item-123",
		ContentType:       "media_episode",
		MediaEpisode: &MediaEpisode{
			MediaEpisodeID: "episode-456",
			Languages: map[string]LocalizedText{
				"en": {Name: "English Title", Slug: "english-title"},
				"no": {Name: "Norwegian Title", Slug: "norwegian-title"},
				"de": {Name: "German Title", Slug: "german-title"},
			},
		},
	}

	data := item.ExtractContentData("plan-001")

	assert.Equal(t, "plan-001", data.PlanID)
	assert.Equal(t, "item-123", data.TaskID)
	assert.Equal(t, "episode-456", data.ContentID)
	assert.Equal(t, "media_episode", data.ContentType)
	assert.Equal(t, "English Title", data.Titles["en"])
	assert.Equal(t, "Norwegian Title", data.Titles["nb"])
	assert.Equal(t, "German Title", data.Titles["de"])
}

func TestItem_ExtractContentData_Song(t *testing.T) {
	item := &Item{
		PlanChapterItemID: "item-789",
		ContentType:       "song",
		Song: &Song{
			SongID: "song-111",
			Number: 45,
			Languages: map[string]LocalizedText{
				"en": {Name: "Amazing Grace", Slug: "amazing-grace"},
				"no": {Name: "Underbar nad", Slug: "underbar-nad"},
			},
			Parent: &Songbook{
				Slug: "fmb",
			},
		},
	}

	data := item.ExtractContentData("plan-002")

	assert.Equal(t, "plan-002", data.PlanID)
	assert.Equal(t, "item-789", data.TaskID)
	assert.Equal(t, "song-111", data.ContentID)
	assert.Equal(t, "song", data.ContentType)
	assert.Equal(t, "FMB 45 - Amazing Grace", data.Titles["en"])
	assert.Equal(t, "FMB 45 - Underbar nad", data.Titles["nb"])
}

func TestItem_ExtractContentData_BookChapter(t *testing.T) {
	item := &Item{
		PlanChapterItemID: "item-222",
		ContentType:       "book_chapter",
		BookChapter: &BookChapter{
			SectionID: 42,
			Languages: map[string]LocalizedText{
				"en": {Name: "Chapter Title", Slug: "chapter-title"},
			},
		},
	}

	data := item.ExtractContentData("plan-003")

	assert.Equal(t, "plan-003", data.PlanID)
	assert.Equal(t, "item-222", data.TaskID)
	assert.Equal(t, "42", data.ContentID)
	assert.Equal(t, "book_chapter", data.ContentType)
	assert.Equal(t, "Chapter Title", data.Titles["en"])
}

func TestItem_ExtractContentData_PeriodicalArticle(t *testing.T) {
	item := &Item{
		PlanChapterItemID: "item-333",
		ContentType:       "periodical_article",
		PeriodicalArticle: &PeriodicalArticle{
			SectionID: 99,
			Languages: map[string]LocalizedText{
				"en": {Name: "Article Title", Slug: "article-title"},
				"de": {Name: "German Article", Slug: "german-article"},
			},
		},
	}

	data := item.ExtractContentData("plan-004")

	assert.Equal(t, "plan-004", data.PlanID)
	assert.Equal(t, "item-333", data.TaskID)
	assert.Equal(t, "99", data.ContentID)
	assert.Equal(t, "periodical_article", data.ContentType)
	assert.Equal(t, "Article Title", data.Titles["en"])
	assert.Equal(t, "German Article", data.Titles["de"])
}

func TestItem_ExtractContentData_BibleVerse(t *testing.T) {
	item := &Item{
		PlanChapterItemID: "item-555",
		ContentType:       "bible_verse",
		Name:              "Genesis 1:1-5",
		BibleVersionsText: map[string]BibleVersion{
			"ESV": {BookName: "Genesis", USFM: "GEN.1.1", Text: "In the beginning..."},
			"KJV": {BookName: "Genesis", USFM: "GEN.1.1", Text: "In the beginning..."},
		},
	}

	data := item.ExtractContentData("plan-006")

	assert.Equal(t, "plan-006", data.PlanID)
	assert.Equal(t, "item-555", data.TaskID)
	assert.Equal(t, "", data.ContentID) // Bible verses have no single content ID
	assert.Equal(t, "bible_verse", data.ContentType)
	assert.Equal(t, "GEN.1.1", data.Titles["nb"]) // Uses USFM from bible_versions_text
	assert.Equal(t, "usfm", data.TitleSource)
}

func TestItem_ExtractContentData_BibleVerse_FallbackToName(t *testing.T) {
	item := &Item{
		PlanChapterItemID: "item-556",
		ContentType:       "bible_verse",
		Name:              "Genesis 1:1-5",
		BibleVersionsText: map[string]BibleVersion{}, // Empty, should fall back to Name
	}

	data := item.ExtractContentData("plan-006")

	assert.Equal(t, "Genesis 1:1-5", data.Titles["nb"]) // Falls back to Name
	assert.Equal(t, "name_fallback", data.TitleSource)
}

func TestItem_ExtractContentData_Text(t *testing.T) {
	item := &Item{
		PlanChapterItemID: "item-600",
		ContentType:       "text",
		Name:              "Introduksjon til kapittelet",
	}

	data := item.ExtractContentData("plan-008")

	assert.Equal(t, "text", data.ContentType)
	assert.Equal(t, "Introduksjon til kapittelet", data.Titles["nb"])
}

func TestItem_ExtractContentData_Quiz(t *testing.T) {
	item := &Item{
		PlanChapterItemID: "item-601",
		ContentType:       "quiz",
	}

	data := item.ExtractContentData("plan-009")

	assert.Equal(t, "quiz", data.ContentType)
	assert.Equal(t, "Quiz", data.Titles["nb"])
}

func TestItem_ExtractContentData_UnknownType(t *testing.T) {
	item := &Item{
		PlanChapterItemID: "item-666",
		ContentType:       "unknown_type",
	}

	data := item.ExtractContentData("plan-007")

	assert.Equal(t, "plan-007", data.PlanID)
	assert.Equal(t, "item-666", data.TaskID)
	assert.Equal(t, "", data.ContentID)
	assert.Equal(t, "unknown_type", data.ContentType)
	assert.Empty(t, data.Titles)
}

func TestExtractTitlesFromLanguages(t *testing.T) {
	languages := map[string]LocalizedText{
		"en": {Name: "English", Slug: "english"},
		"no": {Name: "Norwegian", Slug: "norwegian"},
		"de": {Name: "", Slug: "german"}, // Empty name should be filtered
	}

	titles := extractTitlesFromLanguages(languages)

	assert.Equal(t, 2, len(titles))
	assert.Equal(t, "English", titles["en"])
	assert.Equal(t, "Norwegian", titles["nb"])
	_, hasGerman := titles["de"]
	assert.False(t, hasGerman, "Empty names should be filtered out")
}

func TestExtractTitlesFromLanguages_Nil(t *testing.T) {
	titles := extractTitlesFromLanguages(nil)
	assert.Empty(t, titles)
}

func TestIntToString(t *testing.T) {
	assert.Equal(t, "", intToString(0))
	assert.Equal(t, "1", intToString(1))
	assert.Equal(t, "42", intToString(42))
	assert.Equal(t, "12345", intToString(12345))
}
