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

// validateConsentScope rejects a scope the CHECK constraint would reject at
// insert time anyway, so an admin sees why instead of a raw Postgres
// constraint-violation string
func validateConsentScope(isRemote bool, projectID *string) error {
	if isRemote && projectID != nil {
		return fmt.Errorf(
			"a remote consent cannot be scoped to a project: it is managed by an external system that has no notion of projects",
		)
	}

	return nil
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
