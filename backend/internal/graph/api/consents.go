package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// consentProjectForNewVersion decides which project a new version of an
// existing consent key belongs to.
//
// The scoping project is a property of the consent key, not of one version of
// it, so a new version inherits it. A caller asking for a different project is
// rejected rather than silently obeyed: re-scoping a consent on a version bump
// would change who it applies to without anyone asking, and re-scoping to NULL
// would quietly widen a single project's consent to every project.
func consentProjectForNewVersion(key string, existing, requested *string) (*string, error) {
	if requested != nil && (existing == nil || *existing != *requested) {
		return nil, fmt.Errorf(
			"consent key %q already applies to %s; a new version cannot change that",
			key, describeConsentProject(existing),
		)
	}

	return existing, nil
}

func describeConsentProject(projectID *string) string {
	if projectID == nil {
		return "all projects"
	}

	return "project " + *projectID
}

// resolveConsentProject returns the project to store on a newly created
// consent: whatever the caller asked for when the key is new, and the existing
// key's project otherwise
func (r *Resolver) resolveConsentProject(ctx context.Context, key string, requested *string) (*string, error) {
	previous, err := r.DB.Queries.GetLatestConsentByKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// New key, so there is nothing to inherit from
			return requested, nil
		}

		return nil, fmt.Errorf("failed to look up existing versions of consent %q: %w", key, err)
	}

	return consentProjectForNewVersion(key, previous.ProjectID, requested)
}

// latestConsentVersionError refuses an answer to any but the latest published
// version of a consent (latestID; empty when none is published). Answers are
// per version (00117): an answer to an old version would not count, and the
// user has not seen the latest one.
func latestConsentVersionError(consentID, latestID string) error {
	if latestID == "" || latestID != consentID {
		return fmt.Errorf("only the latest version of a consent can be answered")
	}
	return nil
}

// requireLatestConsentVersion applies latestConsentVersionError to the latest
// published version of the consent's key
func (r *Resolver) requireLatestConsentVersion(ctx context.Context, consentID, key string) error {
	latest, err := r.DB.Queries.GetLatestPublishedConsentByKey(ctx, key)
	if errors.Is(err, pgx.ErrNoRows) {
		return latestConsentVersionError(consentID, "")
	}
	if err != nil {
		return fmt.Errorf("failed to load the latest consent version: %w", err)
	}
	return latestConsentVersionError(consentID, latest.ID)
}
