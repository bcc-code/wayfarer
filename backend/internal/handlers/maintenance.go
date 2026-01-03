package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// MaintenanceHandler handles maintenance tasks like syncing user data
type MaintenanceHandler struct {
	DB            *database.DB
	MembersClient *members.Client
	AuthHandler   *AuthHandler
}

// SyncUserDataResponse contains the results of a user data sync operation
type SyncUserDataResponse struct {
	Processed int      `json:"processed"`
	Updated   int      `json:"updated"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

// SyncUserData syncs incomplete user data from Members API
// This endpoint is designed to be called by a cron job
func (h *MaintenanceHandler) SyncUserData(c *gin.Context) {
	ctx := c.Request.Context()

	// Get limit from query param, default to 100
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	slog.Info("maintenance: starting user data sync", "limit", limit)

	// Get users with incomplete data
	users, err := h.DB.Queries.GetUsersWithIncompleteData(ctx, int32(limit))
	if err != nil {
		slog.Error("maintenance: failed to get users with incomplete data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
		return
	}

	if len(users) == 0 {
		c.JSON(http.StatusOK, SyncUserDataResponse{
			Processed: 0,
			Updated:   0,
			Failed:    0,
		})
		return
	}

	response := SyncUserDataResponse{
		Processed: len(users),
		Updated:   0,
		Failed:    0,
		Errors:    []string{},
	}

	for _, user := range users {
		// Get person ID from members_id
		personID, err := strconv.Atoi(user.MembersID)
		if err != nil {
			slog.Warn("maintenance: invalid members_id", "user_id", user.ID, "members_id", user.MembersID)
			response.Failed++
			response.Errors = append(response.Errors, "invalid members_id for user "+user.ID)
			continue
		}

		// Fetch member data from Members API
		member, err := h.MembersClient.Lookup(ctx, personID)
		if err != nil {
			slog.Warn("maintenance: failed to fetch member data",
				"user_id", user.ID,
				"person_id", personID,
				"error", err,
			)
			response.Failed++
			response.Errors = append(response.Errors, "failed to fetch member "+user.ID+": "+err.Error())
			continue
		}

		// Determine updates
		var newGender string
		var newChurchID string

		// Update gender if currently UNKNOWN and member has gender
		if user.Gender == "UNKNOWN" && member.Gender != "" {
			newGender = normalizeGender(member.Gender)
		}

		// Update church if using default church and member has affiliation
		if h.isDefaultChurch(ctx, user.ChurchID) {
			church, err := h.AuthHandler.findChurchFromAffiliations(ctx, member.Affiliations)
			if err != nil {
				slog.Debug("maintenance: no valid church from affiliations",
					"user_id", user.ID,
					"error", err,
				)
			} else {
				newChurchID = church.ID
			}
		}

		// Skip if nothing to update
		if newGender == "" && newChurchID == "" {
			slog.Debug("maintenance: no updates needed for user", "user_id", user.ID)
			continue
		}

		// Update user
		err = h.DB.Queries.UpdateUserGenderAndChurch(ctx, sqlc.UpdateUserGenderAndChurchParams{
			ID:       user.ID,
			Gender:   newGender,
			ChurchID: newChurchID,
		})
		if err != nil {
			slog.Error("maintenance: failed to update user",
				"user_id", user.ID,
				"error", err,
			)
			response.Failed++
			response.Errors = append(response.Errors, "failed to update user "+user.ID+": "+err.Error())
			continue
		}

		slog.Info("maintenance: updated user data",
			"user_id", user.ID,
			"new_gender", newGender,
			"new_church_id", newChurchID,
		)
		response.Updated++
	}

	slog.Info("maintenance: user data sync complete",
		"processed", response.Processed,
		"updated", response.Updated,
		"failed", response.Failed,
	)

	c.JSON(http.StatusOK, response)
}

// SyncSingleUser syncs data for a single user by ID
func (h *MaintenanceHandler) SyncSingleUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	slog.Info("maintenance: syncing single user", "user_id", userID)

	// Get user
	user, err := h.DB.Queries.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("maintenance: user not found", "user_id", userID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get person ID from members_id
	personID, err := strconv.Atoi(user.MembersID)
	if err != nil {
		slog.Error("maintenance: invalid members_id", "user_id", userID, "members_id", user.MembersID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid members_id"})
		return
	}

	// Fetch member data from Members API
	member, err := h.MembersClient.Lookup(ctx, personID)
	if err != nil {
		slog.Error("maintenance: failed to fetch member data",
			"user_id", userID,
			"person_id", personID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch member data"})
		return
	}

	// Determine updates - always sync from Members API for single user
	var newGender string
	var newChurchID string

	// Always update gender from member data
	if member.Gender != "" {
		newGender = normalizeGender(member.Gender)
	}

	// Always attempt to update church from member affiliation
	church, err := h.AuthHandler.findChurchFromAffiliations(ctx, member.Affiliations)
	if err != nil {
		slog.Debug("maintenance: no valid church from affiliations",
			"user_id", userID,
			"error", err,
		)
	} else {
		newChurchID = church.ID
	}

	// Update user
	err = h.DB.Queries.UpdateUserGenderAndChurch(ctx, sqlc.UpdateUserGenderAndChurchParams{
		ID:       userID,
		Gender:   newGender,
		ChurchID: newChurchID,
	})
	if err != nil {
		slog.Error("maintenance: failed to update user", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// Always update person_uuid from member data
	if member.Uid != uuid.Nil {
		err = h.DB.Queries.UpdateUserPersonUUID(ctx, sqlc.UpdateUserPersonUUIDParams{
			ID:         userID,
			PersonUuid: pgtype.UUID{Bytes: member.Uid, Valid: true},
		})
		if err != nil {
			slog.Warn("maintenance: failed to update person_uuid", "user_id", userID, "error", err)
		}
	}

	slog.Info("maintenance: updated user data",
		"user_id", userID,
		"new_gender", newGender,
		"new_church_id", newChurchID,
	)

	c.JSON(http.StatusOK, gin.H{
		"message":       "user updated",
		"user_id":       userID,
		"new_gender":    newGender,
		"new_church_id": newChurchID,
	})
}

// isDefaultChurch checks if the given church ID is the default church (external_id IS NULL)
func (h *MaintenanceHandler) isDefaultChurch(ctx context.Context, churchID string) bool {
	church, err := h.DB.Queries.GetChurchByID(ctx, churchID)
	if err != nil {
		return false
	}
	return church.ExternalID == nil
}
