package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ChurchResolver resolves churches from external IDs and member affiliations.
// Extracted from AuthHandler so it can be reused across handlers and services.
type ChurchResolver struct {
	DB            *database.DB
	MembersClient *members.Client
}

// FindChurchFromAffiliations finds the first valid non-excluded church from member affiliations.
// It iterates through all active affiliations and returns the first one that is not excluded.
func (r *ChurchResolver) FindChurchFromAffiliations(ctx context.Context, affiliations []members.Affiliation) (*sqlc.GetChurchByExternalIDRow, error) {
	orgUIDs := members.GetActiveAffiliationOrgUIDs(affiliations)
	if len(orgUIDs) == 0 {
		return nil, fmt.Errorf("no active affiliations found")
	}

	for _, orgUID := range orgUIDs {
		church, err := r.findChurchByOrgUID(ctx, orgUID)
		if err != nil {
			slog.Debug("FindChurchFromAffiliations: skipping affiliation",
				"org_uid", orgUID,
				"error", err,
			)
			continue
		}
		return church, nil
	}

	return nil, fmt.Errorf("no valid church found from %d affiliations (all excluded or invalid)", len(orgUIDs))
}

// FindChurchByExternalID finds a church by its external ID, or creates it from the Members API if not found
func (r *ChurchResolver) FindChurchByExternalID(ctx context.Context, externalID int32) (*sqlc.GetChurchByExternalIDRow, error) {
	church, err := r.DB.Queries.GetChurchByExternalID(ctx, &externalID)
	if err == nil {
		return church, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Church not found, try to create it from Members API
	if r.MembersClient == nil {
		return nil, fmt.Errorf("church with external_id %d not found and Members API not configured", externalID)
	}

	slog.Info("church_resolver: church not found, fetching from Members API", "external_id", externalID)

	org, err := r.MembersClient.GetOrganizationByOrgID(ctx, int(externalID))

	var churchName, country, category string
	if err != nil {
		slog.Warn("church_resolver: failed to fetch organization from Members API, creating placeholder",
			"external_id", externalID,
			"error", err,
		)
		churchName = fmt.Sprintf("Church %d", externalID)
		country = "Unknown"
		category = "S"
	} else {
		churchName = org.Name
		country = org.VisitingAddress.CountryCode
		if country == "" {
			country = "Unknown"
		}
		category = "S"
	}

	newChurch, err := r.DB.Queries.CreateChurch(ctx, sqlc.CreateChurchParams{
		ID:         ulid.NewChurchID(),
		ExternalID: &externalID,
		Name:       churchName,
		Country:    country,
		Category:   category,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create church: %w", err)
	}

	slog.Info("church_resolver: created new church from Members API",
		"church_id", newChurch.ID,
		"name", newChurch.Name,
		"external_id", externalID,
	)

	return &sqlc.GetChurchByExternalIDRow{
		ID:         newChurch.ID,
		ExternalID: newChurch.ExternalID,
		Name:       newChurch.Name,
		Country:    newChurch.Country,
		Category:   newChurch.Category,
	}, nil
}

// findChurchByOrgUID finds a church by looking up the org UUID in Members API first
func (r *ChurchResolver) findChurchByOrgUID(ctx context.Context, orgUID uuid.UUID) (*sqlc.GetChurchByExternalIDRow, error) {
	if r.MembersClient == nil {
		return nil, fmt.Errorf("members API not configured")
	}

	org, err := r.MembersClient.GetOrganizationByUID(ctx, orgUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization from Members API: %w", err)
	}

	if members.ExcludedOrgNames[org.Name] {
		return nil, fmt.Errorf("organization %q is excluded from church assignment", org.Name)
	}

	return r.FindChurchByExternalID(ctx, int32(org.OrgID))
}
