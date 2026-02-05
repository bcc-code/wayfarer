package api

import (
	"fmt"
	"strings"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/jackc/pgx/v5/pgtype"
)

func buildExternalContentFilterParams(filter model.ExternalContentFilter, sortBy *model.ExternalContentSortBy, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	if filter.PlanID != nil {
		params["planId"] = *filter.PlanID
	}
	if filter.TaskID != nil {
		params["taskId"] = *filter.TaskID
	}
	if filter.ContentID != nil {
		params["contentId"] = *filter.ContentID
	}
	if filter.ContentType != nil {
		params["contentType"] = string(*filter.ContentType)
	}
	if filter.Source != nil {
		params["source"] = *filter.Source
	}
	if filter.PublishedAfter != nil {
		params["publishedAfter"] = filter.PublishedAfter.String()
	}
	if filter.PublishedBefore != nil {
		params["publishedBefore"] = filter.PublishedBefore.String()
	}
	if sortBy != nil {
		params["sortBy"] = string(*sortBy)
	}
	if first != nil {
		params["first"] = fmt.Sprintf("%d", *first)
	}
	if after != nil {
		params["after"] = *after
	}
	if last != nil {
		params["last"] = fmt.Sprintf("%d", *last)
	}
	if before != nil {
		params["before"] = *before
	}

	return params
}

func buildExternalContentSearchParams(filter model.ExternalContentFilter, sortBy *model.ExternalContentSortBy, limit int, afterCursor string, beforeCursor string, backward bool) sqlc.SearchExternalContentAdminParams {
	params := sqlc.SearchExternalContentAdminParams{
		Querylimit: int32(limit),
	}

	if filter.PlanID != nil {
		params.Planid = *filter.PlanID
	}
	if filter.TaskID != nil {
		params.Taskid = *filter.TaskID
	}
	if filter.ContentID != nil {
		params.Contentid = *filter.ContentID
	}
	if filter.ContentType != nil {
		params.Contenttype = strings.ToLower(string(*filter.ContentType))
	}
	if filter.Source != nil {
		params.Source = *filter.Source
	}
	if filter.PublishedAfter != nil {
		params.Publishedafter = pgtype.Timestamptz{Time: filter.PublishedAfter.Time, Valid: true}
	}
	if filter.PublishedBefore != nil {
		params.Publishedbefore = pgtype.Timestamptz{Time: filter.PublishedBefore.Time, Valid: true}
	}

	// Set sort order
	if sortBy != nil {
		params.Sortby = strings.ToLower(string(*sortBy))
	} else {
		params.Sortby = "created_at_desc"
	}

	return params
}

func buildExternalContentCountParams(filter model.ExternalContentFilter) sqlc.CountExternalContentAdminParams {
	params := sqlc.CountExternalContentAdminParams{}

	if filter.PlanID != nil {
		params.Planid = *filter.PlanID
	}
	if filter.TaskID != nil {
		params.Taskid = *filter.TaskID
	}
	if filter.ContentID != nil {
		params.Contentid = *filter.ContentID
	}
	if filter.ContentType != nil {
		params.Contenttype = strings.ToLower(string(*filter.ContentType))
	}
	if filter.Source != nil {
		params.Source = *filter.Source
	}
	if filter.PublishedAfter != nil {
		params.Publishedafter = pgtype.Timestamptz{Time: filter.PublishedAfter.Time, Valid: true}
	}
	if filter.PublishedBefore != nil {
		params.Publishedbefore = pgtype.Timestamptz{Time: filter.PublishedBefore.Time, Valid: true}
	}

	return params
}

func convertExternalContentRow(row *sqlc.ExternalContent) *model.ExternalContent {
	content := &model.ExternalContent{
		ID:          row.ID,
		PlanID:      row.PlanID,
		TaskID:      row.TaskID,
		ContentType: model.ExternalContentType(strings.ToUpper(row.ContentType)),
		Source:      row.Source,
		SyncedAt:    scalars.DateTime{Time: row.SyncedAt.Time},
		CreatedAt:   scalars.DateTime{Time: row.CreatedAt.Time},
		UpdatedAt:   scalars.DateTime{Time: row.UpdatedAt.Time},
	}

	if row.ContentID != nil {
		content.ContentID = row.ContentID
	}
	if row.PublishedAt.Valid {
		content.PublishedAt = &scalars.DateTime{Time: row.PublishedAt.Time}
	}

	return content
}
