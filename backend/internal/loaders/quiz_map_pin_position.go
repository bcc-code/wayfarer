package loaders

import (
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services/quizgrading"
)

// MapPinPosition reads a MAP_PIN response's stored pin. One that can't be read
// is logged and returned as nil: submittedPosition is non-null, so reading it
// then fails, rather than showing a pin at (0, 0) that nobody placed.
func MapPinPosition(responseID string, jsonResponse []byte) *model.MapPoint {
	p, err := quizgrading.DecodeMapPoint(jsonResponse)
	if err != nil {
		slog.Error("failed to read MAP_PIN response", "response_id", responseID, "error", err)
		return nil
	}
	return &model.MapPoint{X: p.X, Y: p.Y}
}
