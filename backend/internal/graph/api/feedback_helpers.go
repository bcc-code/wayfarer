package api

import (
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/utils"
)

// feedbackRowToModel converts a sqlc UserFeedback row to a GraphQL model
func feedbackRowToModel(row *sqlc.UserFeedback) *model.UserFeedback {
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
		CreatedAt:    scalars.DateTime{Time: row.CreatedAt.Time},
	}
	if row.HandledAt.Valid {
		result.HandledAt = &scalars.DateTime{Time: row.HandledAt.Time}
	}
	return result
}
