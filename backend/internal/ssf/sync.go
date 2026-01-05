package ssf

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5/pgtype"
)

// SyncService handles syncing SSF content to the database
type SyncService struct {
	client  *Client
	queries *sqlc.Queries
	logger  *slog.Logger
}

// NewSyncService creates a new sync service
func NewSyncService(client *Client, queries *sqlc.Queries, logger *slog.Logger) *SyncService {
	return &SyncService{
		client:  client,
		queries: queries,
		logger:  logger,
	}
}

// SyncResult contains statistics from a sync operation
type SyncResult struct {
	PlanID       string
	Slug         string
	ChapterCount int
	ItemCount    int
	Duration     time.Duration
}

// SyncPlanBySlug syncs all content from a plan to the database
// Only items published after today are synced
func (s *SyncService) SyncPlanBySlug(ctx context.Context, slug string) (*SyncResult, error) {
	s.logger.Info("Starting SSF plan sync", "slug", slug)
	startTime := time.Now()

	// Fetch all chapters from the API
	plan, err := s.client.GetAllPlanChaptersBySlug(ctx, slug, "no")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plan chapters: %w", err)
	}

	syncedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	var itemCount int

	// Process each chapter
	for i := range plan.Chapters {
		chapter := &plan.Chapters[i]
		chapterPublishedAt := parsePublishedDate(chapter.DatetimePublished)

		// Process main chapter item if present
		if chapter.MainChapterItem != nil {
			if err := s.upsertItem(ctx, plan.PlanID, chapter.MainChapterItem, chapterPublishedAt, syncedAt); err != nil {
				s.logger.Error("Failed to upsert main chapter item",
					"chapter_id", chapter.PlanChapterID,
					"error", err,
				)
				return nil, fmt.Errorf("failed to upsert main chapter item: %w", err)
			}
			itemCount++
		}

		// Process all items in the chapter
		for j := range chapter.Items {
			item := &chapter.Items[j]
			if err := s.upsertItem(ctx, plan.PlanID, item, chapterPublishedAt, syncedAt); err != nil {
				s.logger.Error("Failed to upsert SSF content",
					"task_id", item.PlanChapterItemID,
					"error", err,
				)
				return nil, fmt.Errorf("failed to upsert content %s: %w", item.PlanChapterItemID, err)
			}
			itemCount++
		}
	}

	duration := time.Since(startTime)
	s.logger.Info("SSF plan sync completed",
		"slug", slug,
		"plan_id", plan.PlanID,
		"chapters", len(plan.Chapters),
		"items", itemCount,
		"duration_ms", duration.Milliseconds(),
	)

	return &SyncResult{
		PlanID:       plan.PlanID,
		Slug:         slug,
		ChapterCount: len(plan.Chapters),
		ItemCount:    itemCount,
		Duration:     duration,
	}, nil
}

func (s *SyncService) upsertItem(ctx context.Context, planID string, item *Item, chapterPublishedAt *time.Time, syncedAt pgtype.Timestamptz) error {
	data := item.ExtractContentData(planID)

	if item.ContentType == "bible_verse" {
		s.logger.Debug("bible_verse title extraction",
			"task_id", item.PlanChapterItemID,
			"title_source", data.TitleSource,
			"title", data.Titles["nb"],
		)
	}

	// Upsert the main content record
	params := sqlc.UpsertExternalContentParams{
		ID:          ulid.NewExternalContentID(),
		Planid:      planID,
		Taskid:      item.PlanChapterItemID,
		Contentid:   data.ContentID,
		Contenttype: item.ContentType,
		Syncedat:    syncedAt,
		Source:      "ssf",
	}

	// Three-tier fallback strategy to ensure all items have a published date
	// Prefer chapter date first as it's the most authoritative
	publishedAt := chapterPublishedAt // 1. Chapter's publication date (preferred)
	if publishedAt == nil {
		publishedAt = data.PublishedAt // 2. Item's own date (from content fields)
	}
	if publishedAt == nil {
		publishedAt = &syncedAt.Time // 3. Sync timestamp (always valid)
	}

	// At this point publishedAt is guaranteed to be non-nil
	params.Publishedat = pgtype.Timestamptz{Time: *publishedAt, Valid: true}

	content, err := s.queries.UpsertExternalContent(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to upsert content: %w", err)
	}

	// Upsert translations
	for langCode, title := range data.Titles {
		if title == "" {
			continue
		}
		err := s.queries.UpsertExternalContentTranslation(ctx, sqlc.UpsertExternalContentTranslationParams{
			Externalcontentid: content.ID,
			Languagecode:      langCode,
			Title:             title,
		})
		if err != nil {
			s.logger.Warn("Failed to upsert translation",
				"content_id", content.ID,
				"language", langCode,
				"error", err,
			)
			// Don't fail the whole sync for translation errors
		}
	}

	return nil
}

// isChapterPublishedAfterDate checks if a chapter is published after the given date
// Chapters without a published date are skipped (returns false)
func isChapterPublishedAfterDate(chapter *PlanChapter, date time.Time) bool {
	if chapter.DatetimePublished == "" {
		return false
	}
	publishedAt, err := time.Parse(time.RFC3339, chapter.DatetimePublished)
	if err != nil {
		return false
	}
	return publishedAt.After(date)
}

// parsePublishedDate parses the datetime_published string from a chapter
func parsePublishedDate(datetimePublished string) *time.Time {
	if datetimePublished == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, datetimePublished)
	if err != nil {
		return nil
	}
	return &t
}
