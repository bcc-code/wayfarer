package api

import (
	"context"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/utils"
)

// resolveFeedbackProjectID falls back to the active project when the client did
// not send one. Returns "" when no project can be determined.
func resolveFeedbackProjectID(ctx context.Context, projects services.ProjectIDProvider, submitted *string) string {
	if submitted != nil && *submitted != "" {
		return *submitted
	}
	if projects == nil {
		return ""
	}
	projectID, err := projects.GetCurrentProjectID(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to resolve current project for feedback", "error", err)
		return ""
	}
	return projectID
}

// feedbackRowToModel converts a sqlc UserFeedback row to a GraphQL model
func feedbackRowToModel(row *sqlc.UserFeedback) *model.UserFeedback {
	tags := row.Tags
	if tags == nil {
		tags = []string{}
	}
	result := &model.UserFeedback{
		ID:           row.ID,
		UserID:       row.UserID,
		Message:      row.Message,
		CanContactMe: row.CanContactMe,
		UserAgent:    row.UserAgent,
		Platform:     row.Platform,
		ScreenWidth:  utils.Int32PtrToIntPtr(row.ScreenWidth),
		ScreenHeight: utils.Int32PtrToIntPtr(row.ScreenHeight),
		AppVersion:   row.AppVersion,
		Locale:       row.Locale,
		ProjectID:    row.ProjectID,
		Timezone:     row.Timezone,
		ContextURL:   row.ContextUrl,
		Tags:         tags,
		CreatedAt:    scalars.DateTime{Time: row.CreatedAt.Time},
	}
	if row.HandledAt.Valid {
		result.HandledAt = &scalars.DateTime{Time: row.HandledAt.Time}
	}
	return result
}
