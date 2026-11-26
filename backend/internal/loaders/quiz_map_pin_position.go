package loaders

import (
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services/quizgrading"
	"github.com/jackc/pgx/v5/pgtype"
)

// MapPinPosition is a MAP_PIN response's pin, from its map_position column, in
// whole pixels. A missing or broken pin is logged and returned as nil:
// submittedPosition is non-null, so reading it then fails, rather than showing
// a pin that nobody placed.
func MapPinPosition(responseID string, position pgtype.Point) *model.MapPoint {
	p, err := quizgrading.PointFromPG(position)
	if err != nil || p == nil {
		slog.Error("MAP_PIN response has no readable pin", "response_id", responseID, "error", err)
		return nil
	}
	return &model.MapPoint{X: p.X, Y: p.Y}
}
