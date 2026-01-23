package ladder_to_heaven

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/bcc-media/wayfarer/internal/webhook"
	"github.com/gin-gonic/gin"
)

// teamNameChangedHandler handles webhook requests for team name changes.
type teamNameChangedHandler struct {
	db          *database.DB
	cache       *cache.CacheWithRegistry
	challengeID string
	secretKey   string
}

// teamNameChangedRequest matches the outbound WebhookPayload format for team_name_changed events
type teamNameChangedRequest struct {
	EventType string                 `json:"event_type" binding:"required"`
	Timestamp time.Time              `json:"timestamp" binding:"required"`
	ProjectID string                 `json:"project_id" binding:"required"`
	Data      teamNameChangedData    `json:"data" binding:"required"`
}

// teamNameChangedData contains the team name change data
type teamNameChangedData struct {
	TeamID  string `json:"team_id" binding:"required"`
	OldName string `json:"old_name" binding:"required"`
	NewName string `json:"new_name" binding:"required"`
}

const (
	// Points awarded to each team member when the team renames
	pointsTeamRename = 300
	// Source type for team rename entries in score_journal
	sourceTypeTeamRename = "TEAM_RENAME"
)

// handle processes incoming team name changed webhook requests
func (h *teamNameChangedHandler) handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if feature is enabled
	if h.challengeID == "" {
		slog.Warn("ladder_to_heaven: team rename feature disabled, PLUGIN_LADDER_TO_HEAVEN_TEAM_RENAME_CHALLENGE_ID not set")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "feature disabled"})
		return
	}

	// Read raw body for signature verification
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Warn("ladder_to_heaven: failed to read request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Verify webhook signature if secret key is configured
	if h.secretKey != "" {
		signature := c.GetHeader("X-Webhook-Signature")
		if !webhook.VerifySignature(body, signature, h.secretKey) {
			slog.Warn("ladder_to_heaven: invalid webhook signature")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Restore body for JSON binding
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	// Parse request body
	var req teamNameChangedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("ladder_to_heaven: invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	slog.Info("ladder_to_heaven: processing team name change",
		"team_id", req.Data.TeamID,
		"old_name", req.Data.OldName,
		"new_name", req.Data.NewName,
		"project_id", req.ProjectID,
		"challenge_id", h.challengeID,
	)

	// Check if we've already processed this team (using team ID as source_id)
	alreadyProcessed, err := h.db.Queries.CheckScoreJournalEntryExistsBySource(ctx, sqlc.CheckScoreJournalEntryExistsBySourceParams{
		SourceType: sourceTypeTeamRename,
		SourceID:   req.Data.TeamID,
	})
	if err != nil {
		slog.Error("ladder_to_heaven: failed to check if team already processed", "error", err, "team_id", req.Data.TeamID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if alreadyProcessed {
		slog.Info("ladder_to_heaven: team rename already processed", "team_id", req.Data.TeamID)
		c.JSON(http.StatusConflict, gin.H{"error": "team rename already processed"})
		return
	}

	// Get all team members
	teamMembers, err := h.db.Queries.GetUserIDsInTeams(ctx, []string{req.Data.TeamID})
	if err != nil {
		slog.Error("ladder_to_heaven: failed to get team members", "error", err, "team_id", req.Data.TeamID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get team members"})
		return
	}

	if len(teamMembers) == 0 {
		slog.Warn("ladder_to_heaven: no members in team", "team_id", req.Data.TeamID)
		c.JSON(http.StatusOK, gin.H{"message": "no team members to process"})
		return
	}

	slog.Info("ladder_to_heaven: processing team members", "team_id", req.Data.TeamID, "member_count", len(teamMembers))

	// Award points to each team member and complete the challenge
	reason := "Team renamed: " + req.Data.NewName
	for _, userID := range teamMembers {
		// Create score journal entry for points
		journalID := ulid.NewScoreJournalID()
		reasonPtr := &reason

		_, err := h.db.Queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
			ID:          journalID,
			ProjectID:   req.ProjectID,
			UserID:      userID,
			ChallengeID: &h.challengeID,
			Points:      pointsTeamRename,
			SourceType:  sourceTypeTeamRename,
			SourceID:    &req.Data.TeamID,
			Reason:      reasonPtr,
		})
		if err != nil {
			slog.Error("ladder_to_heaven: failed to create score journal entry",
				"error", err, "user_id", userID, "team_id", req.Data.TeamID)
			// Continue processing other members
			continue
		}

		// Invalidate user cache
		h.cache.InvalidateUser(userID)

		slog.Debug("ladder_to_heaven: awarded team rename points",
			"user_id", userID,
			"team_id", req.Data.TeamID,
			"points", pointsTeamRename)
	}

	// Mark challenge as completed for all team members
	err = h.db.Queries.BulkCompleteChallenges(ctx, sqlc.BulkCompleteChallengesParams{
		Userids:     teamMembers,
		Challengeid: h.challengeID,
	})
	if err != nil {
		slog.Error("ladder_to_heaven: failed to complete challenge for team members",
			"error", err, "team_id", req.Data.TeamID, "challenge_id", h.challengeID)
		// Don't fail - points were already awarded
	}

	// Invalidate project cache
	h.cache.InvalidateProject(req.ProjectID)

	slog.Info("ladder_to_heaven: team rename processed successfully",
		"team_id", req.Data.TeamID,
		"members_processed", len(teamMembers),
		"points_per_member", pointsTeamRename)

	c.JSON(http.StatusCreated, gin.H{
		"message":          "team rename processed",
		"members_awarded":  len(teamMembers),
		"points_per_member": pointsTeamRename,
	})
}
