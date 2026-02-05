package ssf

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncService handles syncing SSF content to the database
type SyncService struct {
	client  *Client
	queries *sqlc.Queries
	pool    *pgxpool.Pool
	logger  *slog.Logger
}

// NewSyncService creates a new sync service
func NewSyncService(client *Client, queries *sqlc.Queries, pool *pgxpool.Pool, logger *slog.Logger) *SyncService {
	return &SyncService{
		client:  client,
		queries: queries,
		pool:    pool,
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

// SyncPlanBySlug syncs all content from a plan to the database using bulk operations
func (s *SyncService) SyncPlanBySlug(ctx context.Context, slug string) (*SyncResult, error) {
	s.logger.Info("Starting SSF plan sync", "slug", slug)
	startTime := time.Now()

	// Fetch all chapters from the API
	plan, err := s.client.GetAllPlanChaptersBySlug(ctx, slug, "no")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plan chapters: %w", err)
	}

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// Query existing content for this plan to reuse their IDs
	existingContent, err := qtx.GetExternalContentByPlanID(ctx, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing content: %w", err)
	}
	existingIDsByTaskID := make(map[string]string)
	for _, content := range existingContent {
		existingIDsByTaskID[content.TaskID] = content.ID
	}

	// Use a map to deduplicate items by task_id (last occurrence wins)
	syncedAt := time.Now()
	itemsByTaskID := make(map[string]*itemData)
	var taskIDOrder []string // preserve insertion order

	// Collect all items from all chapters
	for i := range plan.Chapters {
		chapter := &plan.Chapters[i]
		chapterPublishedAt := parsePublishedDate(chapter.DatetimePublished)
		chapterCompleteBy := calculateChapterCompleteBy(chapter)

		// Collect main chapter item if present
		if chapter.MainChapterItem != nil {
			s.collectItemDataToMap(
				plan.PlanID, chapter.MainChapterItem, chapterPublishedAt, chapterCompleteBy, syncedAt,
				itemsByTaskID, &taskIDOrder, existingIDsByTaskID,
			)
		}

		// Collect all items in the chapter
		for j := range chapter.Items {
			item := &chapter.Items[j]
			s.collectItemDataToMap(
				plan.PlanID, item, chapterPublishedAt, chapterCompleteBy, syncedAt,
				itemsByTaskID, &taskIDOrder, existingIDsByTaskID,
			)
		}
	}

	// Convert map to arrays for bulk upsert
	var ids, planIDs, taskIDs, contentIDs, contentTypes, sources, urls []string
	var publishedAts, syncedAts, completeBys []pgtype.Timestamptz
	var translationContentIDs, langCodes, titles []string

	for _, taskID := range taskIDOrder {
		item := itemsByTaskID[taskID]
		ids = append(ids, item.id)
		planIDs = append(planIDs, item.planID)
		taskIDs = append(taskIDs, item.taskID)
		contentIDs = append(contentIDs, item.contentID)
		contentTypes = append(contentTypes, item.contentType)
		publishedAts = append(publishedAts, item.publishedAt)
		syncedAts = append(syncedAts, item.syncedAt)
		sources = append(sources, item.source)
		urls = append(urls, item.url)
		completeBys = append(completeBys, item.completeBy)

		// Add translations
		for langCode, title := range item.titles {
			translationContentIDs = append(translationContentIDs, item.id)
			langCodes = append(langCodes, langCode)
			titles = append(titles, title)
		}
	}

	itemCount := len(ids)

	// Bulk upsert content
	if itemCount > 0 {
		err = qtx.BulkUpsertExternalContent(ctx, sqlc.BulkUpsertExternalContentParams{
			Ids:          ids,
			Planids:      planIDs,
			Taskids:      taskIDs,
			Contentids:   contentIDs,
			Contenttypes: contentTypes,
			Publishedats: publishedAts,
			Syncedats:    syncedAts,
			Sources:      sources,
			Urls:         urls,
			Completebys:  completeBys,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to bulk upsert content: %w", err)
		}
	}

	// Bulk upsert translations
	if len(translationContentIDs) > 0 {
		err = qtx.BulkUpsertExternalContentTranslations(ctx, sqlc.BulkUpsertExternalContentTranslationsParams{
			Externalcontentids: translationContentIDs,
			Languagecodes:      langCodes,
			Titles:             titles,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to bulk upsert translations: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	duration := time.Since(startTime)
	s.logger.Info("SSF plan sync completed",
		"slug", slug,
		"plan_id", plan.PlanID,
		"chapters", len(plan.Chapters),
		"items", itemCount,
		"translations", len(translationContentIDs),
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

// itemData holds collected item data for bulk upsert
type itemData struct {
	id          string
	planID      string
	taskID      string
	contentID   string
	contentType string
	publishedAt pgtype.Timestamptz
	syncedAt    pgtype.Timestamptz
	source      string
	url         string
	completeBy  pgtype.Timestamptz
	titles      map[string]string // langCode -> title
}

// collectItemDataToMap collects item data into a map, deduplicating by task_id
func (s *SyncService) collectItemDataToMap(
	planID string, item *Item, chapterPublishedAt *time.Time, completeBy *time.Time, syncedAt time.Time,
	itemsByTaskID map[string]*itemData, taskIDOrder *[]string, existingIDsByTaskID map[string]string,
) {
	data := item.ExtractContentData(planID)
	taskID := item.PlanChapterItemID

	// Check if we've seen this task_id before in this sync
	existing, seenInSync := itemsByTaskID[taskID]
	if !seenInSync {
		*taskIDOrder = append(*taskIDOrder, taskID)
	}

	// Three-tier fallback strategy to ensure all items have a published date
	publishedAt := chapterPublishedAt
	if publishedAt == nil {
		publishedAt = data.PublishedAt
	}
	if publishedAt == nil {
		publishedAt = &syncedAt
	}

	// Determine ID: reuse existing DB ID, or existing sync ID, or generate new
	var id string
	if existingDBID, existsInDB := existingIDsByTaskID[taskID]; existsInDB {
		id = existingDBID // Reuse ID from database
	} else if seenInSync {
		id = existing.id // Reuse ID from earlier in this sync
	} else {
		id = ulid.NewExternalContentID() // Generate new ID
	}

	var completeByTS pgtype.Timestamptz
	if completeBy != nil {
		completeByTS = pgtype.Timestamptz{Time: *completeBy, Valid: true}
	}

	// Filter out empty titles
	filteredTitles := make(map[string]string)
	for langCode, title := range data.Titles {
		if title != "" {
			filteredTitles[langCode] = title
		}
	}

	itemsByTaskID[taskID] = &itemData{
		id:          id,
		planID:      planID,
		taskID:      taskID,
		contentID:   data.ContentID,
		contentType: item.ContentType,
		publishedAt: pgtype.Timestamptz{Time: *publishedAt, Valid: true},
		syncedAt:    pgtype.Timestamptz{Time: syncedAt, Valid: true},
		source:      "ssf",
		url:         "",
		completeBy:  completeByTS,
		titles:      filteredTitles,
	}
}

// parsePublishedDate parses the datetime_published string from a chapter
func parsePublishedDate(datetimePublished string) *time.Time {
	if datetimePublished == "" {
		return nil
	}
	// Try RFC3339 format first (e.g., "2025-12-04T10:30:00Z")
	t, err := time.Parse(time.RFC3339, datetimePublished)
	if err == nil {
		return &t
	}
	// Try datetime without timezone (e.g., "2026-01-13T02:00:00")
	t, err = time.Parse("2006-01-02T15:04:05", datetimePublished)
	if err == nil {
		return &t
	}
	return nil
}

// calculateChapterCompleteBy calculates the complete_by timestamp for a chapter
// based on the main chapter item's completion mode.
// Returns nil if main item is nil, doesn't have "required_24h" mode, or chapter has no published date.
func calculateChapterCompleteBy(chapter *PlanChapter) *time.Time {
	if chapter == nil || chapter.MainChapterItem == nil {
		return nil
	}
	if chapter.MainChapterItem.CompletionMode != "required_24h" {
		return nil
	}

	publishedAt := parsePublishedDate(chapter.DatetimePublished)
	if publishedAt == nil {
		return nil
	}

	completeBy := publishedAt.Add(24 * time.Hour)
	return &completeBy
}
