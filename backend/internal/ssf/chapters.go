package ssf

import (
	"context"
	"fmt"
)

// GetPlanBySlug fetches a plan with all its chapters by slug
// The acceptLanguage parameter is passed to the API but SSF returns all languages regardless
func (c *Client) GetPlanBySlug(ctx context.Context, slug string, acceptLanguage string) (*PlanResponse, error) {
	endpoint := fmt.Sprintf("/v1/plans/slug/%s/chapters", slug)

	c.logger.Info("Fetching SSF plan",
		"slug", slug,
		"language", acceptLanguage,
	)

	result, err := get[PlanResponse](ctx, c, endpoint, acceptLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan for slug %s: %w", slug, err)
	}

	c.logger.Info("Successfully fetched SSF plan",
		"slug", slug,
		"plan_id", result.PlanID,
		"chapter_count", len(result.Chapters),
	)

	return result, nil
}

// GetPlanChaptersBySlug fetches chapters for a plan by slug with pagination
// The acceptLanguage parameter is passed to the API but SSF returns all languages regardless
func (c *Client) GetPlanChaptersBySlug(ctx context.Context, slug string, acceptLanguage string, offset, limit int) (*PlanResponse, error) {
	endpoint := fmt.Sprintf("/v1/plans/slug/%s/chapters?offset=%d&limit=%d", slug, offset, limit)

	c.logger.Info("Fetching SSF plan chapters",
		"slug", slug,
		"offset", offset,
		"limit", limit,
	)

	chapters, err := get[[]PlanChapter](ctx, c, endpoint, acceptLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan chapters for slug %s: %w", slug, err)
	}

	// Build response with slug as plan_id
	result := &PlanResponse{
		PlanID:       slug,
		Slug:         slug,
		Chapters:     *chapters,
		ChapterCount: len(*chapters),
	}

	c.logger.Info("Successfully fetched SSF plan chapters",
		"slug", slug,
		"chapter_count", len(*chapters),
	)

	return result, nil
}

// GetAllPlanChaptersBySlug fetches all chapters with automatic pagination
func (c *Client) GetAllPlanChaptersBySlug(ctx context.Context, slug string, acceptLanguage string) (*PlanResponse, error) {
	const pageSize = 10
	offset := 0

	c.logger.Info("Fetching all SSF plan chapters",
		"slug", slug,
	)

	// Fetch chapters in pages
	var allChapters []PlanChapter
	for {
		endpoint := fmt.Sprintf("/v1/plans/slug/%s/chapters?offset=%d&limit=%d", slug, offset, pageSize)

		c.logger.Debug("Fetching page",
			"slug", slug,
			"offset", offset,
			"limit", pageSize,
		)

		chapters, err := get[[]PlanChapter](ctx, c, endpoint, acceptLanguage)
		if err != nil {
			return nil, fmt.Errorf("failed to get plan chapters at offset %d: %w", offset, err)
		}

		c.logger.Debug("Fetched page",
			"slug", slug,
			"offset", offset,
			"chapters_count", len(*chapters),
		)

		allChapters = append(allChapters, *chapters...)

		// If we got fewer chapters than the page size, we've reached the end
		if len(*chapters) < pageSize {
			break
		}

		offset += pageSize
	}

	result := &PlanResponse{
		PlanID:       slug,
		Slug:         slug,
		Chapters:     allChapters,
		ChapterCount: len(allChapters),
	}

	c.logger.Info("Successfully fetched all SSF plan chapters",
		"slug", slug,
		"total_chapters", len(allChapters),
	)

	return result, nil
}
