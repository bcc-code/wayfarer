package handlers

import (
	"log/slog"
	"net/http"

	"github.com/bcc-media/wayfarer/internal/ssf"
	"github.com/gin-gonic/gin"
)

// SSFHandler handles SSF sync endpoints
type SSFHandler struct {
	SyncService *ssf.SyncService
	SyncKey     string
}

// SyncResponse represents the response from a sync operation
type SyncResponse struct {
	PlanID       string `json:"plan_id"`
	Slug         string `json:"slug"`
	ChapterCount int    `json:"chapter_count"`
	ItemCount    int    `json:"item_count"`
	DurationMs   int64  `json:"duration_ms"`
}

// HandleSyncPlan handles POST /ssf/sync/:slug
// Authenticated via X-Sync-Key header or ?key= query param
func (h *SSFHandler) HandleSyncPlan(c *gin.Context) {
	// Validate sync key (header or query param)
	key := c.GetHeader("X-Sync-Key")
	if key == "" {
		key = c.Query("key")
	}
	if key == "" || key != h.SyncKey {
		slog.Warn("SSF sync: unauthorized request",
			"has_key", key != "",
			"path", c.Request.URL.Path,
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing key (use X-Sync-Key header or ?key= param)"})
		return
	}

	ctx := c.Request.Context()
	slug := c.Param("slug")

	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug parameter is required"})
		return
	}

	slog.Info("SSF sync: starting plan sync", "slug", slug)

	result, err := h.SyncService.SyncPlanBySlug(ctx, slug)
	if err != nil {
		slog.Error("SSF sync: sync failed",
			"slug", slug,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "sync failed",
			"details": err.Error(),
		})
		return
	}

	slog.Info("SSF sync: completed",
		"slug", slug,
		"plan_id", result.PlanID,
		"items", result.ItemCount,
	)

	c.JSON(http.StatusOK, SyncResponse{
		PlanID:       result.PlanID,
		Slug:         result.Slug,
		ChapterCount: result.ChapterCount,
		ItemCount:    result.ItemCount,
		DurationMs:   result.Duration.Milliseconds(),
	})
}
