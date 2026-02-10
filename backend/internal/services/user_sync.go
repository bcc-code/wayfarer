package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/ssf"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserSyncService synchronizes user data from external sources (SSF, Members API).
type UserSyncService struct {
	DB                        *database.DB
	Cache                     *cache.CacheWithRegistry
	SSFClient                 *ssf.Client
	MembersClient             *members.Client
	ChurchResolver            *ChurchResolver
	ContentAchievementService *ContentAchievementService
}

// SyncUserResult contains the results of a user sync operation.
type SyncUserResult struct {
	ContentEventsProcessed int
	GenderUpdated          bool
	ChurchUpdated          bool
	ChurchLockSkipped      bool
	PersonUUIDUpdated      bool
}

// SyncUser synchronizes a user's content events from SSF and profile data from Members API.
func (s *UserSyncService) SyncUser(ctx context.Context, userID string) (*SyncUserResult, error) {
	user, err := s.DB.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	result := &SyncUserResult{}

	// Sync content events from SSF
	if s.SSFClient != nil && user.PersonUuid.Valid {
		personUUIDStr := uuid.UUID(user.PersonUuid.Bytes).String()
		processed, err := s.syncContentEvents(ctx, user.ID, user.PersonUuid, personUUIDStr)
		if err != nil {
			slog.Error("user_sync: failed to sync content events",
				"user_id", userID,
				"error", err,
			)
			// Continue to member sync even if SSF sync fails
		} else {
			result.ContentEventsProcessed = processed
		}
	}

	// Sync member data from Members API
	if s.MembersClient != nil {
		memberResult, err := s.syncMemberData(ctx, user)
		if err != nil {
			slog.Error("user_sync: failed to sync member data",
				"user_id", userID,
				"error", err,
			)
			// Return partial result with content events count
		} else {
			result.GenderUpdated = memberResult.GenderUpdated
			result.ChurchUpdated = memberResult.ChurchUpdated
			result.ChurchLockSkipped = memberResult.ChurchLockSkipped
			result.PersonUUIDUpdated = memberResult.PersonUUIDUpdated
		}
	}

	// Invalidate all user-related cache entries after sync
	if s.Cache != nil {
		s.Cache.InvalidateUser(userID)
	}

	return result, nil
}

// SyncContentEventsFromSSF fetches content events from SSF for a user.
// Used when SSF consent is granted to backfill historical events.
func (s *UserSyncService) SyncContentEventsFromSSF(ctx context.Context, userID string) (int, error) {
	if s.SSFClient == nil {
		return 0, fmt.Errorf("SSF client not configured")
	}

	user, err := s.DB.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.PersonUuid.Valid {
		return 0, fmt.Errorf("user has no person_uuid")
	}

	personUUIDStr := uuid.UUID(user.PersonUuid.Bytes).String()
	processed, err := s.syncContentEvents(ctx, userID, user.PersonUuid, personUUIDStr)
	if err != nil {
		return processed, err
	}

	// Invalidate user cache after sync
	if s.Cache != nil {
		s.Cache.InvalidateUser(userID)
	}

	return processed, nil
}

// syncContentEvents fetches content events from SSF and processes them through the pipeline.
// Returns the number of events processed.
func (s *UserSyncService) syncContentEvents(ctx context.Context, userID string, personUUID pgtype.UUID, personUUIDStr string) (int, error) {
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)
	processed := 0
	page := 1

	for {
		resp, err := s.SSFClient.GetMemberContentEvents(ctx, personUUIDStr, page)
		if err != nil {
			return processed, fmt.Errorf("failed to fetch content events page %d: %w", page, err)
		}

		if len(resp.Items) == 0 {
			break
		}

		// Filter events newer than 6-month cutoff and collect their task IDs
		type recentEvent struct {
			event  ssf.ContentEvent
			taskID string
		}
		var recent []recentEvent
		allOlderThanCutoff := true
		for _, event := range resp.Items {
			if event.Timestamp.Before(sixMonthsAgo) {
				continue
			}
			allOlderThanCutoff = false
			recent = append(recent, recentEvent{event: event, taskID: event.TaskID})
		}

		// Batch-check which task IDs already exist for this person
		if len(recent) > 0 {
			taskIDs := make([]string, len(recent))
			for i, r := range recent {
				taskIDs[i] = r.taskID
			}

			existingTaskIDs, err := s.DB.Queries.GetExistingExternalContentEventTaskIDs(ctx, sqlc.GetExistingExternalContentEventTaskIDsParams{
				Personid: personUUID,
				Taskids:  taskIDs,
			})
			if err != nil {
				return processed, fmt.Errorf("failed to check existing content events on page %d: %w", page, err)
			}

			existingSet := make(map[string]struct{}, len(existingTaskIDs))
			for _, id := range existingTaskIDs {
				existingSet[id] = struct{}{}
			}

			for _, r := range recent {
				if _, exists := existingSet[r.taskID]; exists {
					continue
				}

				contentProgress := &r.event.ContentProgress
				err := s.ContentAchievementService.StoreAndProcessContentEvent(
					ctx,
					userID,
					personUUID,
					r.event.TaskID,
					r.event.PlanID,
					contentProgress,
					r.event.Timestamp,
					"ssf-sync",
					true, // force: skip per-event dedup check since we batch-checked above
				)
				if err != nil {
					slog.Warn("user_sync: failed to store content event",
						"user_id", userID,
						"task_id", r.event.TaskID,
						"error", err,
					)
					continue
				}
				processed++
			}
		}

		// Stop paginating when all events on a page are older than 6 months
		if allOlderThanCutoff {
			break
		}

		if !resp.HasMore {
			break
		}
		page++
	}

	slog.Info("user_sync: content events sync complete",
		"user_id", userID,
		"processed", processed,
	)

	return processed, nil
}

type memberSyncResult struct {
	GenderUpdated     bool
	ChurchUpdated     bool
	ChurchLockSkipped bool
	PersonUUIDUpdated bool
}

// syncMemberData fetches member data from the Members API and force-updates user profile.
func (s *UserSyncService) syncMemberData(ctx context.Context, user *sqlc.GetUserByIDRow) (*memberSyncResult, error) {
	personID, err := strconv.Atoi(user.MembersID)
	if err != nil {
		return nil, fmt.Errorf("invalid members_id %q: %w", user.MembersID, err)
	}

	member, err := s.MembersClient.Lookup(ctx, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch member data: %w", err)
	}

	result := &memberSyncResult{}

	// Normalize gender
	newGender := ""
	if member.Gender != "" {
		newGender = members.NormalizeGender(member.Gender)
		result.GenderUpdated = newGender != user.Gender
	}

	// Resolve church from affiliations (skip if church is locked)
	newChurchID := ""
	if user.ChurchLockedUntil.Valid && user.ChurchLockedUntil.Time.After(time.Now()) {
		slog.Info("user_sync: church update skipped due to lock",
			"user_id", user.ID,
			"locked_until", user.ChurchLockedUntil.Time,
		)
		result.ChurchLockSkipped = true
	} else if s.ChurchResolver != nil {
		church, err := s.ChurchResolver.FindChurchFromAffiliations(ctx, member.Affiliations)
		if err != nil {
			slog.Debug("user_sync: no valid church from affiliations",
				"user_id", user.ID,
				"error", err,
			)
		} else {
			newChurchID = church.ID
			result.ChurchUpdated = newChurchID != user.ChurchID
		}
	}

	// Force update gender and church
	err = s.DB.Queries.UpdateUserGenderAndChurch(ctx, sqlc.UpdateUserGenderAndChurchParams{
		ID:       user.ID,
		Gender:   newGender,
		ChurchID: newChurchID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user gender and church: %w", err)
	}

	// Update person_uuid if member has it
	if member.Uid != uuid.Nil {
		newPersonUUID := pgtype.UUID{Bytes: member.Uid, Valid: true}
		if !user.PersonUuid.Valid || uuid.UUID(user.PersonUuid.Bytes) != member.Uid {
			err = s.DB.Queries.UpdateUserPersonUUID(ctx, sqlc.UpdateUserPersonUUIDParams{
				ID:         user.ID,
				PersonUuid: newPersonUUID,
			})
			if err != nil {
				slog.Warn("user_sync: failed to update person_uuid",
					"user_id", user.ID,
					"error", err,
				)
			} else {
				result.PersonUUIDUpdated = true
			}
		}
	}

	slog.Info("user_sync: member data sync complete",
		"user_id", user.ID,
		"gender_updated", result.GenderUpdated,
		"church_updated", result.ChurchUpdated,
		"person_uuid_updated", result.PersonUUIDUpdated,
	)

	return result, nil
}
