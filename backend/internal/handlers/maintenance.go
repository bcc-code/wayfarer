package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/ssf"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// maintenanceSyncConcurrency bounds how many users are synced against the Members API
// at once. Each Members lookup can take up to its client timeout; processing them
// sequentially would let a single sync request run for minutes. Bounded concurrency
// keeps wall-clock low while staying well within the DB connection pool.
const maintenanceSyncConcurrency = 10

// MaintenanceHandler handles maintenance tasks like syncing user data
type MaintenanceHandler struct {
	DB                        *database.DB
	Cache                     *cache.CacheWithRegistry
	MembersClient             *members.Client
	ChurchResolver            *services.ChurchResolver
	AuthHandler               *AuthHandler
	ContentAchievementService *services.ContentAchievementService
	SSFClient                 *ssf.Client
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

	// Process users concurrently with a bounded pool so the Members API lookups
	// don't run fully serial (a single stalled lookup would otherwise block the rest).
	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		sem = make(chan struct{}, maintenanceSyncConcurrency)
	)

	for _, user := range users {
		wg.Add(1)
		sem <- struct{}{}
		go func(user *sqlc.GetUsersWithIncompleteDataRow) {
			defer wg.Done()
			defer func() { <-sem }()

			// Get person ID from members_id
			personID, err := strconv.Atoi(user.MembersID)
			if err != nil {
				slog.Warn("maintenance: invalid members_id", "user_id", user.ID, "members_id", user.MembersID)
				mu.Lock()
				response.Failed++
				response.Errors = append(response.Errors, "invalid members_id for user "+user.ID)
				mu.Unlock()
				return
			}

			// Fetch member data from Members API
			member, err := h.MembersClient.Lookup(ctx, personID)
			if err != nil {
				slog.Warn("maintenance: failed to fetch member data",
					"user_id", user.ID,
					"person_id", personID,
					"error", err,
				)
				mu.Lock()
				response.Failed++
				response.Errors = append(response.Errors, "failed to fetch member "+user.ID+": "+err.Error())
				mu.Unlock()
				return
			}

			// Extract profile fields (email, name, birthdate, etc.) from Members API data.
			// Empty fields mean "no data" — the update query below leaves those columns alone.
			profile := members.ExtractProfile(member)

			// Determine gender/church updates
			var newGender string
			var newChurchID string

			// Update gender if currently UNKNOWN and member has gender
			if user.Gender == "UNKNOWN" && member.Gender != "" {
				newGender = members.NormalizeGender(member.Gender)
			}

			// Update church if using default church and member has affiliation (skip if locked)
			if user.ChurchLockedUntil.Valid && user.ChurchLockedUntil.Time.After(time.Now()) {
				slog.Debug("maintenance: church update skipped due to lock",
					"user_id", user.ID,
					"locked_until", user.ChurchLockedUntil.Time,
				)
			} else if h.isDefaultChurch(ctx, user.ChurchID) {
				church, err := h.ChurchResolver.FindChurchFromAffiliations(ctx, member.Affiliations)
				if err != nil {
					slog.Debug("maintenance: no valid church from affiliations",
						"user_id", user.ID,
						"error", err,
					)
				} else {
					newChurchID = church.ID
				}
			}

			var birthdate pgtype.Date
			if profile.Birthdate != nil {
				birthdate = pgtype.Date{Time: *profile.Birthdate, Valid: true}
			}

			// Skip if nothing to update
			if newGender == "" && newChurchID == "" && profile == (members.ProfileFields{}) {
				slog.Debug("maintenance: no updates needed for user", "user_id", user.ID)
				return
			}

			// Update user
			err = h.DB.Queries.UpdateUserProfileFromMembers(ctx, sqlc.UpdateUserProfileFromMembersParams{
				ID:          user.ID,
				Email:       profile.Email,
				Name:        profile.Name,
				FirstName:   profile.FirstName,
				LastName:    profile.LastName,
				MiddleName:  profile.MiddleName,
				DisplayName: profile.DisplayName,
				Gender:      newGender,
				Birthdate:   birthdate,
				ChurchID:    newChurchID,
			})
			if err != nil {
				slog.Error("maintenance: failed to update user",
					"user_id", user.ID,
					"error", err,
				)
				mu.Lock()
				response.Failed++
				response.Errors = append(response.Errors, "failed to update user "+user.ID+": "+err.Error())
				mu.Unlock()
				return
			}

			// Invalidate all user-related cache entries
			if h.Cache != nil {
				h.Cache.InvalidateUser(user.ID)
			}

			slog.Info("maintenance: updated user data",
				"user_id", user.ID,
				"new_gender", newGender,
				"new_church_id", newChurchID,
				"birthdate_updated", profile.Birthdate != nil,
			)
			mu.Lock()
			response.Updated++
			mu.Unlock()
		}(user)
	}

	wg.Wait()

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

	processPending := c.Query("process_pending") == "true"

	slog.Info("maintenance: syncing single user", "user_id", userID, "process_pending", processPending)

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
		newGender = members.NormalizeGender(member.Gender)
	}

	// Attempt to update church from member affiliation (skip if locked)
	if user.ChurchLockedUntil.Valid && user.ChurchLockedUntil.Time.After(time.Now()) {
		slog.Debug("maintenance: church update skipped due to lock",
			"user_id", userID,
			"locked_until", user.ChurchLockedUntil.Time,
		)
	} else {
		church, err := h.ChurchResolver.FindChurchFromAffiliations(ctx, member.Affiliations)
		if err != nil {
			slog.Debug("maintenance: no valid church from affiliations, keeping existing",
				"user_id", userID,
				"affiliations", member.Affiliations,
				"error", err,
			)
			// Keep existing church - don't update to default
		} else {
			newChurchID = church.ID
		}
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

	// Process pending data only when explicitly requested
	onboardingProcessed := false
	if processPending && member.Uid != uuid.Nil {
		personUUIDStr := member.Uid.String()
		personUUID := pgtype.UUID{Bytes: member.Uid, Valid: true}

		h.AuthHandler.ProcessPendingConsentEvents(ctx, userID, personUUIDStr)

		if h.ContentAchievementService != nil {
			h.ContentAchievementService.ProcessPendingContentEvents(ctx, userID, personUUID)
		}

		onboardingProcessed = true
		slog.Info("maintenance: processed onboarding events", "user_id", userID, "person_uuid", personUUIDStr)
	}

	// Invalidate all user-related cache entries
	if h.Cache != nil {
		h.Cache.InvalidateUser(userID)
	}

	slog.Info("maintenance: updated user data",
		"user_id", userID,
		"new_gender", newGender,
		"new_church_id", newChurchID,
		"onboarding_processed", onboardingProcessed,
	)

	c.JSON(http.StatusOK, gin.H{
		"message":              "user updated",
		"user_id":              userID,
		"new_gender":           newGender,
		"new_church_id":        newChurchID,
		"onboarding_processed": onboardingProcessed,
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

// BackfillSSFEventsResponse contains the results of a backfill operation
type BackfillSSFEventsResponse struct {
	Year                int      `json:"year"`
	Month               int      `json:"month"`
	Page                int      `json:"page"`
	EventsFetched       int      `json:"events_fetched"`
	EventsProcessed     int      `json:"events_processed"`
	EventsSkippedNoUser int      `json:"events_skipped_no_user"`
	EventsSkippedDupe   int      `json:"events_skipped_duplicate"`
	HasMore             bool     `json:"has_more"`
	NextPage            int      `json:"next_page,omitempty"`
	Errors              []string `json:"errors,omitempty"`
}

// BackfillSSFEvents backfills user content completion events from SSF for a given month/page.
// This endpoint is designed to be called by an external orchestrator (cron, webhook, manual).
func (h *MaintenanceHandler) BackfillSSFEvents(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate SSF client is configured
	if h.SSFClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SSF client not configured"})
		return
	}

	// Parse and validate year
	yearStr := c.Query("year")
	if yearStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "year is required"})
		return
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}

	// Parse and validate month
	monthStr := c.Query("month")
	if monthStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month is required"})
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month (1-12)"})
		return
	}

	// Parse page (default to 1)
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	slog.Info("maintenance: starting SSF events backfill",
		"year", year,
		"month", month,
		"page", page,
	)

	// Fetch events from SSF
	resp, err := h.SSFClient.GetMonthlyContentEvents(ctx, year, month, page)
	if err != nil {
		slog.Error("maintenance: failed to fetch SSF events",
			"year", year,
			"month", month,
			"page", page,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events from SSF"})
		return
	}

	response := BackfillSSFEventsResponse{
		Year:          year,
		Month:         month,
		Page:          page,
		EventsFetched: len(resp.Items),
		HasMore:       resp.HasMore,
		Errors:        []string{},
	}

	if resp.HasMore {
		response.NextPage = page + 1
	}

	if len(resp.Items) == 0 {
		slog.Info("maintenance: no events to process",
			"year", year,
			"month", month,
			"page", page,
		)
		c.JSON(http.StatusOK, response)
		return
	}

	// Collect unique person_uuids and group events by person
	eventsByPerson := make(map[string][]ssf.ContentEvent)
	personUUIDStrings := make(map[string]struct{})

	for _, event := range resp.Items {
		if event.PersonID == "" {
			continue
		}
		personUUIDStrings[event.PersonID] = struct{}{}
		eventsByPerson[event.PersonID] = append(eventsByPerson[event.PersonID], event)
	}

	// Convert to slice for batch lookup
	personUUIDs := make([]pgtype.UUID, 0, len(personUUIDStrings))
	for uuidStr := range personUUIDStrings {
		parsed, err := uuid.Parse(uuidStr)
		if err != nil {
			slog.Warn("maintenance: invalid person_id UUID",
				"person_id", uuidStr,
				"error", err,
			)
			continue
		}
		personUUIDs = append(personUUIDs, pgtype.UUID{Bytes: parsed, Valid: true})
	}

	// Batch lookup users by person_uuids
	users, err := h.DB.Queries.GetUsersByPersonUUIDs(ctx, personUUIDs)
	if err != nil {
		slog.Error("maintenance: failed to lookup users by person_uuids",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lookup users"})
		return
	}

	// Build user map: person_uuid string -> user
	userByPersonUUID := make(map[string]*sqlc.GetUsersByPersonUUIDsRow, len(users))
	for _, u := range users {
		if u.PersonUuid.Valid {
			uuidStr := uuid.UUID(u.PersonUuid.Bytes).String()
			userByPersonUUID[uuidStr] = u
		}
	}

	// Process events for each person
	for personUUIDStr, events := range eventsByPerson {
		user, found := userByPersonUUID[personUUIDStr]
		if !found {
			response.EventsSkippedNoUser += len(events)
			continue
		}

		// Collect task IDs for this person's events
		taskIDs := make([]string, len(events))
		for i, event := range events {
			taskIDs[i] = event.TaskID
		}

		// Bulk check which already exist
		existingTaskIDs, err := h.DB.Queries.GetExistingExternalContentEventTaskIDs(ctx, sqlc.GetExistingExternalContentEventTaskIDsParams{
			Personid: user.PersonUuid,
			Taskids:  taskIDs,
		})
		if err != nil {
			slog.Warn("maintenance: failed to check existing events",
				"user_id", user.ID,
				"error", err,
			)
			response.Errors = append(response.Errors, "failed to check existing events for user "+user.ID)
			continue
		}

		existingSet := make(map[string]struct{}, len(existingTaskIDs))
		for _, id := range existingTaskIDs {
			existingSet[id] = struct{}{}
		}

		// Process only new events
		for _, event := range events {
			if _, exists := existingSet[event.TaskID]; exists {
				response.EventsSkippedDupe++
				continue
			}

			contentProgress := event.ContentProgress
			err := h.ContentAchievementService.StoreAndProcessContentEvent(
				ctx,
				user.ID,
				user.PersonUuid,
				event.TaskID,
				event.PlanID,
				&contentProgress,
				event.Timestamp,
				"ssf-backfill",
				true, // force: skip per-event dedup check since we batch-checked above
			)
			if err != nil {
				slog.Warn("maintenance: failed to process content event",
					"user_id", user.ID,
					"task_id", event.TaskID,
					"error", err,
				)
				response.Errors = append(response.Errors, "failed to process event "+event.TaskID+": "+err.Error())
				continue
			}
			response.EventsProcessed++
		}
	}

	slog.Info("maintenance: SSF events backfill complete",
		"year", year,
		"month", month,
		"page", page,
		"fetched", response.EventsFetched,
		"processed", response.EventsProcessed,
		"skipped_no_user", response.EventsSkippedNoUser,
		"skipped_duplicate", response.EventsSkippedDupe,
		"has_more", response.HasMore,
	)

	c.JSON(http.StatusOK, response)
}
