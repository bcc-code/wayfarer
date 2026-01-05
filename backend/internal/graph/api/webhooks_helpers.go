package api

import (
	"encoding/json"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/utils"
)

func webhookEventTypeToDBString(eventType model.WebhookEventType) string {
	switch eventType {
	case model.WebhookEventTypeExternalContentEvent:
		return "external_content_event"
	case model.WebhookEventTypePointsAwarded:
		return "points_awarded"
	default:
		return string(eventType)
	}
}

func dbStringToWebhookEventType(eventType string) model.WebhookEventType {
	switch eventType {
	case "external_content_event":
		return model.WebhookEventTypeExternalContentEvent
	case "points_awarded":
		return model.WebhookEventTypePointsAwarded
	default:
		return model.WebhookEventType(eventType)
	}
}

func sqlcWebhookToModel(w *sqlc.Webhook) *model.Webhook {
	return &model.Webhook{
		ID:               w.ID,
		ProjectID:        w.ProjectID,
		Name:             w.Name,
		URL:              w.Url,
		EventType:        dbStringToWebhookEventType(w.EventType),
		IncludeUserData:  w.IncludeUserData,
		IncludeEventData: w.IncludeEventData,
		Active:           w.Active,
		Secret:           w.Secret,
		CreatedAt:        scalars.DateTime{Time: w.CreatedAt.Time},
		UpdatedAt:        scalars.DateTime{Time: w.UpdatedAt.Time},
	}
}

func sqlcWebhookLogToModel(l *sqlc.WebhookLog) *model.WebhookLog {
	// Convert JSON bytes to string
	payloadStr := string(l.RequestPayload)

	// Pretty print the JSON if possible
	var prettyPayload interface{}
	if err := json.Unmarshal(l.RequestPayload, &prettyPayload); err == nil {
		if prettyBytes, err := json.MarshalIndent(prettyPayload, "", "  "); err == nil {
			payloadStr = string(prettyBytes)
		}
	}

	return &model.WebhookLog{
		ID:                 l.ID,
		WebhookID:          l.WebhookID,
		EventType:          dbStringToWebhookEventType(l.EventType),
		RequestPayload:     payloadStr,
		ResponseStatusCode: utils.Int32PtrToIntPtr(l.ResponseStatusCode),
		ResponseBody:       l.ResponseBody,
		DurationMs:         int(l.DurationMs),
		ErrorMessage:       l.ErrorMessage,
		CreatedAt:          scalars.DateTime{Time: l.CreatedAt.Time},
	}
}
