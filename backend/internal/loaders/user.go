package loaders

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// userByIDBatchFunc batches user loading by IDs
func userByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.User] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.User] {
		// Check cache first for each ID
		userMap := make(map[string]*model.User)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.UserKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if user, ok := cached.(*model.User); ok {
					userMap[id] = user
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetUsersByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.User], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.User]{Error: err}
				}
				return results
			}

			// Fetch consent data for all users
			latestConsents, userConsentsMap := fetchConsentDataForUsers(ctx, db, c, missingIDs)

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				// Convert birthdate to string (always valid since birthdate is required)
				birthdateStr := row.Birthdate.Time.Format("2006-01-02")

				// Build consent status for this user
				consentStatus := buildConsentStatus(row.ID, row.Birthdate.Time, latestConsents, userConsentsMap)

				// Use display_name if available, otherwise fall back to name
				displayName := row.Name
				if row.DisplayName != nil && *row.DisplayName != "" {
					displayName = *row.DisplayName
				}

				user := &model.User{
					ID:            row.ID,
					MembersID:     row.MembersID,
					Gender:        model.Gender(row.Gender),
					ChurchID:      row.ChurchID,
					Birthdate:     birthdateStr,
					Email:         row.Email,
					Name:          displayName,
					Image:         row.AvatarUrl,
					ConsentStatus: consentStatus,
					Language:      row.Language,
				}

				userMap[row.ID] = user
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.UserKey(row.ID), user)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.User], len(ids))
		for i, id := range ids {
			if user, ok := userMap[id]; ok {
				results[i] = &dataloader.Result[*model.User]{Data: user}
			} else {
				results[i] = &dataloader.Result[*model.User]{
					Error: fmt.Errorf("user not found: %s", id),
				}
			}
		}
		return results
	}
}

// fetchConsentDataForUsers fetches all consent data needed to build ConsentStatus for users
// Returns: latestConsents (map of consent ID -> Consent), userConsentsMap (map of userID -> map of consentID -> UserConsent)
func fetchConsentDataForUsers(ctx context.Context, db *database.DB, c *cache.CacheWithRegistry, userIDs []string) (map[string]*model.Consent, map[string]map[string]*model.UserConsent) {
	// 1. Get all latest published consents (cached globally)
	latestConsents := make(map[string]*model.Consent)
	cacheKey := cache.LatestConsentsKey()
	if cached, ok := c.Get(cacheKey); ok {
		if consents, ok := cached.(map[string]*model.Consent); ok {
			latestConsents = consents
		}
	} else {
		// Query latest consents from DB
		rows, err := db.Queries.GetAllLatestPublishedConsents(ctx)
		if err == nil {
			for _, row := range rows {
				var publishedAt *scalars.DateTime
				if row.PublishedAt.Valid {
					dt := scalars.DateTime{Time: row.PublishedAt.Time}
					publishedAt = &dt
				}

				managementType := model.ConsentManagementTypeLocal
				if row.IsRemote {
					managementType = model.ConsentManagementTypeRemote
				}

				consent := &model.Consent{
					ID:             row.ID,
					Key:            row.Key,
					Version:        int(row.Version),
					Title:          row.Title,
					ShortText:      row.ShortText,
					BodyMarkdown:   row.Body,
					URL:            row.Url,
					PublishedAt:    publishedAt,
					ManagementType: managementType,
					ManagedBy:      row.ManagedBy,
				}
				latestConsents[row.ID] = consent
			}
			// Cache for 15 minutes
			c.Set(cacheKey, latestConsents)
		}
	}

	// 2. Get current user consent statuses for all users
	userConsentsMap := make(map[string]map[string]*model.UserConsent)
	if len(userIDs) > 0 {
		rows, err := db.Queries.GetCurrentUserConsentStatusesByUsers(ctx, userIDs)
		if err == nil {
			for _, row := range rows {
				action := model.ConsentActionAccepted
				if row.Action == "REJECTED" {
					action = model.ConsentActionRejected
				}

				userConsent := &model.UserConsent{
					ID:         row.ID,
					ConsentID:  row.ConsentID,
					Action:     action,
					ActionDate: scalars.DateTime{Time: row.OccurredAt.Time},
					// Consent will be resolved via resolver using ConsentByIDLoader
				}

				if userConsentsMap[row.UserID] == nil {
					userConsentsMap[row.UserID] = make(map[string]*model.UserConsent)
				}
				// Index by consent_key, not consent_id, to handle version changes
				userConsentsMap[row.UserID][row.ConsentKey] = userConsent
			}
		}
	}

	return latestConsents, userConsentsMap
}

// calculateAge returns the age in years based on a birthdate
func calculateAge(birthdate time.Time) int {
	now := time.Now()
	age := now.Year() - birthdate.Year()
	// Adjust if birthday hasn't occurred yet this year
	// Compare month and day directly to handle leap years correctly
	if now.Month() < birthdate.Month() ||
		(now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		age--
	}
	return age
}

// buildConsentStatus builds the ConsentStatus for a user
func buildConsentStatus(userID string, birthdate time.Time, latestConsents map[string]*model.Consent, userConsentsMap map[string]map[string]*model.UserConsent) *model.ConsentStatus {
	userConsents := userConsentsMap[userID]
	if userConsents == nil {
		userConsents = make(map[string]*model.UserConsent)
	}

	userAge := calculateAge(birthdate)

	pendingConsents := make([]model.Consent, 0)
	acceptedConsents := make([]model.UserConsent, 0)
	rejectedConsents := make([]model.UserConsent, 0)

	// Iterate through all latest published consents
	for _, consent := range latestConsents {
		// Skip LOCAL (internal) consents for users under 13
		if userAge < 13 && consent.ManagementType == model.ConsentManagementTypeLocal {
			continue
		}

		// Check by consent key (to handle rejection persistence across versions)
		if userConsent, hasAction := userConsents[consent.Key]; hasAction {
			// User has taken action on this consent
			if userConsent.Action == model.ConsentActionAccepted {
				acceptedConsents = append(acceptedConsents, *userConsent)
			} else if userConsent.Action == model.ConsentActionRejected {
				rejectedConsents = append(rejectedConsents, *userConsent)
			}
		} else {
			// User hasn't taken any action yet
			pendingConsents = append(pendingConsents, *consent)
		}
	}

	return &model.ConsentStatus{
		PendingConsents:  pendingConsents,
		AcceptedConsents: acceptedConsents,
		RejectedConsents: rejectedConsents,
	}
}
