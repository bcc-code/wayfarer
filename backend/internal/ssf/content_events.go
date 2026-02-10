package ssf

import (
	"context"
	"fmt"
)

// GetMonthlyContentEvents fetches content events for a given month with pagination.
func (c *Client) GetMonthlyContentEvents(ctx context.Context, year int, month int, page int) (*ContentEventsResponse, error) {
	endpoint := fmt.Sprintf("/v1/bcc/content-events/monthly?year=%d&month=%d&page=%d", year, month, page)

	c.logger.Info("Fetching monthly content events",
		"year", year,
		"month", month,
		"page", page,
	)

	result, err := get[ContentEventsResponse](ctx, c, endpoint, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly content events for %d-%d page %d: %w", year, month, page, err)
	}

	c.logger.Info("Successfully fetched monthly content events",
		"year", year,
		"month", month,
		"page", page,
		"item_count", len(result.Items),
		"has_more", result.HasMore,
	)

	return result, nil
}

// GetMemberContentEvents fetches content events for a specific member with pagination
func (c *Client) GetMemberContentEvents(ctx context.Context, personID string, page int) (*ContentEventsResponse, error) {
	endpoint := fmt.Sprintf("/v1/bcc/content-events/member/%s?page=%d", personID, page)

	c.logger.Info("Fetching member content events",
		"person_id", personID,
		"page", page,
	)

	result, err := get[ContentEventsResponse](ctx, c, endpoint, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get content events for member %s page %d: %w", personID, page, err)
	}

	c.logger.Info("Successfully fetched member content events",
		"person_id", personID,
		"page", page,
		"item_count", len(result.Items),
		"has_more", result.HasMore,
	)

	return result, nil
}
