package ssf

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LocalizedText represents a localized name/slug pair
type LocalizedText struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// PlanResponse represents the top-level API response for a plan
type PlanResponse struct {
	PlanID           string                   `json:"plan_id"`
	Name             string                   `json:"name"`
	Slug             string                   `json:"slug"`
	Description      string                   `json:"description"`
	CoverImage       string                   `json:"cover_image"`
	DatePublished    string                   `json:"date_published"`
	PlanType         string                   `json:"plan_type"`
	RecommendedOrder int                      `json:"recommended_order"`
	Languages        map[string]LocalizedText `json:"languages"`
	ChapterCount     int                      `json:"chapter_count"`
	Chapters         []PlanChapter            `json:"chapters"`
}

// PlanChapter represents a chapter within a plan
type PlanChapter struct {
	PlanChapterID     string `json:"plan_chapter_id"`
	PlanID            string `json:"plan_id"`
	IndexID           int    `json:"index_id"`
	Name              string `json:"name"`
	ItemCount         int    `json:"item_count"`
	DatetimePublished string `json:"datetime_published"`
	MainChapterItem   *Item  `json:"main_chapter_item"`
	Items             []Item `json:"items"`
}

// Item represents a plan chapter item with content
type Item struct {
	PlanChapterItemID string                  `json:"plan_chapter_item_id"`
	PlanChapterID     string                  `json:"plan_chapter_id"`
	IndexID           int                     `json:"index_id"`
	CompletionMode    string                  `json:"completion_mode"`
	ContentType       string                  `json:"content_type"`
	Name              string                  `json:"name"`
	Subtitle          string                  `json:"subtitle"`
	Description       string                  `json:"description"`
	ExternalURL       string                  `json:"external_url"`
	MediaEpisode      *MediaEpisode           `json:"media_episode,omitempty"`
	Song              *Song                   `json:"song,omitempty"`
	BookChapter       *BookChapter            `json:"book_chapter,omitempty"`
	PeriodicalArticle *PeriodicalArticle      `json:"periodical_article,omitempty"`
	BibleChapter      *BibleChapter           `json:"bible_chapter,omitempty"`
	BibleVersionsText map[string]BibleVersion `json:"bible_versions_text,omitempty"`
}

// MediaEpisode represents a media episode (video/audio)
type MediaEpisode struct {
	MediaEpisodeID string                   `json:"media_episode_id"`
	MediaSeriesID  int                      `json:"media_series_id"`
	IndexID        int                      `json:"index_id"`
	DatePublished  string                   `json:"date_published"`
	Thumbnail      string                   `json:"thumbnail"`
	YoutubeID      string                   `json:"youtube_id"`
	Duration       string                   `json:"duration"`
	Title          string                   `json:"title"`
	Slug           string                   `json:"slug"`
	Description    string                   `json:"description"`
	Audio          string                   `json:"audio"`
	IsAudioOnly    bool                     `json:"is_audio_only"`
	MediaSeries    *MediaSeries             `json:"media_series,omitempty"`
	Languages      map[string]LocalizedText `json:"languages"`
}

// MediaSeries represents a media series
type MediaSeries struct {
	MediaSeriesID    int                      `json:"media_series_id"`
	Name             string                   `json:"name"`
	Description      string                   `json:"description"`
	DatePublished    string                   `json:"date_published"`
	CoverImage       string                   `json:"cover_image"`
	RecommendedOrder int                      `json:"recommended_order"`
	Slug             string                   `json:"slug"`
	Languages        map[string]LocalizedText `json:"languages"`
	EpisodeCount     int                      `json:"episode_count"`
}

// Song represents a song from a songbook
type Song struct {
	SongID              string                   `json:"song_id"`
	Number              int                      `json:"number"`
	OriginalKey         string                   `json:"original_key"`
	OriginalMinor       bool                     `json:"original_minor"`
	TextCopyrightName   string                   `json:"text_copyright_name"`
	MelodyCopyrightName string                   `json:"melody_copyright_name"`
	YearPublished       int                      `json:"year_published"`
	Name                string                   `json:"name"`
	Description         string                   `json:"description"`
	ContentTimestamp    string                   `json:"content_timestamp"`
	HasChords           bool                     `json:"has_chords"`
	ChordsOnly          bool                     `json:"chords_only"`
	Lyrics              string                   `json:"lyrics"`
	Parent              *Songbook                `json:"parent,omitempty"`
	Categories          []SongCategory           `json:"categories"`
	Languages           map[string]LocalizedText `json:"languages"`
}

// Songbook represents a songbook (parent of songs)
type Songbook struct {
	SongbookLangID   int                      `json:"songbook_lang_id"`
	SongbookID       int                      `json:"songbook_id"`
	LanguageCode     string                   `json:"language_code"`
	Name             string                   `json:"name"`
	Slug             string                   `json:"slug"`
	ContentTimestamp string                   `json:"content_timestamp"`
	Languages        map[string]LocalizedText `json:"languages"`
}

// SongCategory represents a song category
type SongCategory struct {
	SongCategoryID string                   `json:"song_category_id"`
	Name           string                   `json:"name"`
	Slug           string                   `json:"slug"`
	Language       string                   `json:"language"`
	Languages      map[string]LocalizedText `json:"languages"`
}

// BookChapter represents a chapter from a book
type BookChapter struct {
	SectionID  int                      `json:"section_id"`
	BookID     int                      `json:"book_id"`
	Title      string                   `json:"title"`
	ChapterNo  int                      `json:"chapter_no"`
	Slug       string                   `json:"slug"`
	Languages  map[string]LocalizedText `json:"languages"`
	Parent     *Book                    `json:"parent,omitempty"`
	AudioFiles []AudioFile              `json:"audio_files"`
}

// Book represents a book (parent of book chapters)
type Book struct {
	BookID             int                      `json:"book_id"`
	Name               string                   `json:"name"`
	DatePublished      string                   `json:"date_published"`
	AuthorIDs          []int                    `json:"author_ids"`
	ISBN               string                   `json:"isbn"`
	CoverID            string                   `json:"cover_id"`
	CoverImageFileName string                   `json:"cover_image_file_name"`
	EbookFileUUID      string                   `json:"ebook_file_uuid"`
	EbookFileName      string                   `json:"ebook_file_name"`
	EbookOriginalName  string                   `json:"ebook_original_name"`
	EbookOnly          bool                     `json:"ebook_only"`
	Noindex            bool                     `json:"noindex"`
	ContentTimestamp   string                   `json:"content_timestamp"`
	LanguageCodes      []string                 `json:"language_codes"`
	RecommendedOrder   int                      `json:"recommended_order"`
	Slug               string                   `json:"slug"`
	Languages          map[string]LocalizedText `json:"languages"`
}

// PeriodicalArticle represents an article from a periodical/magazine
type PeriodicalArticle struct {
	SectionID         int                      `json:"section_id"`
	SectionLangID     int                      `json:"section_lang_id"`
	PeriodicalIssueID int                      `json:"periodical_issue_id"`
	ChapterNo         int                      `json:"chapter_no"`
	Title             string                   `json:"title"`
	Slug              string                   `json:"slug"`
	Noindex           bool                     `json:"noindex"`
	Languages         map[string]LocalizedText `json:"languages"`
	Parent            *PeriodicalIssue         `json:"parent,omitempty"`
	AudioFiles        []AudioFile              `json:"audio_files"`
}

// PeriodicalIssue represents a periodical issue (parent of articles)
type PeriodicalIssue struct {
	PeriodicalIssueID    int    `json:"periodical_issue_id"`
	PeriodicalID         int    `json:"periodical_id"`
	Year                 int    `json:"year"`
	Month                int    `json:"month"`
	NoOfMonths           string `json:"no_of_months"`
	ContentTimestamp     string `json:"content_timestamp"`
	ISSN                 string `json:"issn"`
	PrintVersionFileName string `json:"print_version_file_name"`
	PrintBookletFileName string `json:"print_booklet_file_name"`
}

// BibleChapter represents a bible chapter reference
type BibleChapter struct {
	USFM        string `json:"usfm"`
	Human       string `json:"human"`
	HTMLContent string `json:"html_content"`
}

// BibleVersion represents bible text in a specific version
type BibleVersion struct {
	BookName string `json:"book_name"`
	USFM     string `json:"usfm"`
	Text     string `json:"text"`
	HTML     string `json:"html"`
}

// AudioFile represents an audio file attachment
type AudioFile struct {
	SectionLangAudioID int    `json:"section_lang_audio_id"`
	SectionLangID      int    `json:"section_lang_id"`
	Name               string `json:"name"`
	AudioFile          string `json:"audio_file"`
	TimestampFile      string `json:"timestamp_file"`
}

// ContentData holds extracted content data for storage
type ContentData struct {
	PlanID      string
	TaskID      string
	ContentID   string
	ContentType string
	PublishedAt *time.Time
	Titles      map[string]string // language code -> title
	TitleSource string            // where the title came from (for debugging)
}

// ExtractContentData extracts the relevant data from an Item for storage
func (item *Item) ExtractContentData(planID string) ContentData {
	data := ContentData{
		PlanID:      planID,
		TaskID:      item.PlanChapterItemID,
		ContentType: item.ContentType,
		Titles:      make(map[string]string),
	}

	// Extract content ID and publication date based on content type
	switch item.ContentType {
	case "media_episode", "media": // SSF API uses "media" as content_type
		if item.MediaEpisode != nil {
			data.ContentID = item.MediaEpisode.MediaEpisodeID
			data.Titles = extractTitlesFromLanguages(item.MediaEpisode.Languages)
			data.PublishedAt = parseDate(item.MediaEpisode.DatePublished)
		}
	case "song":
		if item.Song != nil {
			data.ContentID = item.Song.SongID
			// Use song's own Languages, fall back to Name
			if len(item.Song.Languages) > 0 {
				data.Titles = extractTitlesFromLanguages(item.Song.Languages)
			} else if item.Song.Name != "" {
				data.Titles["en"] = item.Song.Name
			}
			// Build prefix: "SLUG NUMBER - " (e.g., "FMB 123 - ")
			var prefix string
			if item.Song.Parent != nil && item.Song.Parent.Slug != "" {
				prefix = strings.ToUpper(item.Song.Parent.Slug) + " "
			}
			if item.Song.Number > 0 {
				prefix += fmt.Sprintf("%d - ", item.Song.Number)
			}
			// Prepend prefix to titles
			if prefix != "" {
				for lang, title := range data.Titles {
					data.Titles[lang] = prefix + title
				}
			}
			// Extract year-based publication date
			if item.Song.YearPublished > 0 {
				data.PublishedAt = parseYear(item.Song.YearPublished)
			}
		}
	case "book_chapter":
		if item.BookChapter != nil {
			data.ContentID = intToString(item.BookChapter.SectionID)
			data.Titles = extractTitlesFromLanguages(item.BookChapter.Languages)
			// Extract book publication date
			if item.BookChapter.Parent != nil {
				data.PublishedAt = parseDate(item.BookChapter.Parent.DatePublished)
			}
		}
	case "periodical_article":
		if item.PeriodicalArticle != nil {
			data.ContentID = intToString(item.PeriodicalArticle.SectionID)
			data.Titles = extractTitlesFromLanguages(item.PeriodicalArticle.Languages)
			// Extract periodical issue date from year/month
			if item.PeriodicalArticle.Parent != nil && item.PeriodicalArticle.Parent.Year > 0 {
				data.PublishedAt = parseYearMonth(item.PeriodicalArticle.Parent.Year, item.PeriodicalArticle.Parent.Month)
			}
		}
	case "bible_verse":
		// Bible verses have multiple versions, no single content ID
		data.ContentID = ""
		// Extract USFM from first available version, fall back to Item.Name
		for _, version := range item.BibleVersionsText {
			if version.USFM != "" {
				data.Titles["nb"] = version.USFM
				data.TitleSource = "usfm"
				break
			}
		}
		if data.TitleSource == "" && item.Name != "" {
			data.Titles["nb"] = item.Name
			data.TitleSource = "name_fallback"
		}
		if data.TitleSource == "" {
			data.TitleSource = "none"
		}
	case "text":
		data.Titles["nb"] = item.Name
	case "quiz":
		data.Titles["nb"] = "Quiz"
	}

	return data
}

func extractTitlesFromLanguages(languages map[string]LocalizedText) map[string]string {
	titles := make(map[string]string)
	for lang, text := range languages {
		if text.Name != "" {
			// Map "no" to "nb" for Norwegian Bokmal
			if lang == "no" {
				lang = "nb"
			}
			titles[lang] = text.Name
		}
	}
	return titles
}

func intToString(i int) string {
	if i == 0 {
		return ""
	}
	return strconv.Itoa(i)
}

// parseDate parses a date string in RFC3339 or date-only format
// Returns nil if the date is empty or invalid
func parseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}

	// Try RFC3339 format first (e.g., "2025-12-04T10:30:00Z")
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return &t
	}

	// Try date-only format (e.g., "2025-12-04")
	t, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return &t
	}

	return nil
}

// parseYear converts a year integer to a date (YYYY-01-01 at midnight UTC)
// Returns nil if year is invalid
func parseYear(year int) *time.Time {
	if year <= 0 || year > 9999 {
		return nil
	}
	t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	return &t
}

// parseYearMonth converts year/month integers to a date (YYYY-MM-01 at midnight UTC)
// Returns nil if year is invalid. Month is clamped to valid range 1-12
func parseYearMonth(year, month int) *time.Time {
	if year <= 0 || year > 9999 {
		return nil
	}
	if month < 1 {
		month = 1
	}
	if month > 12 {
		month = 12
	}
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return &t
}
