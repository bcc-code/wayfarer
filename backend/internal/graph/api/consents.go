package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// projectUser is a user within one project
type projectUser struct {
	ProjectID string
	UserID    string
}

// consentBlockChanges returns the users blocked by a required consent (00116)
// in exactly one of before and after: those a consent change moved into or
// out of every count
func consentBlockChanges(before, after []projectUser) map[projectUser]bool {
	wasBlocked := make(map[projectUser]bool, len(before))
	for _, pu := range before {
		wasBlocked[pu] = true
	}
	isBlocked := make(map[projectUser]bool, len(after))
	changed := make(map[projectUser]bool)
	for _, pu := range after {
		isBlocked[pu] = true
		if !wasBlocked[pu] {
			changed[pu] = true
		}
	}
	for pu := range wasBlocked {
		if !isBlocked[pu] {
			changed[pu] = true
		}
	}
	return changed
}

// teamsOfChangedUsers returns, as team ID to project ID, the teams in which a
// changed user is a member within the project they changed in
func teamsOfChangedUsers(memberships []*sqlc.GetTeamsByUserIDsRow, changed map[projectUser]bool) map[string]string {
	teams := make(map[string]string)
	for _, m := range memberships {
		if changed[projectUser{ProjectID: m.ProjectID, UserID: m.UserID}] {
			teams[m.ID] = m.ProjectID
		}
	}
	return teams
}

// blockedSet is who a required consent blocks (00116) within some scope, as
// listed before or after a consent change. Known is false when listing them
// failed: that is logged rather than failing the consent change.
type blockedSet struct {
	Users []projectUser
	Known bool
}

// blockedProjectsOfUser lists the projects in which the user is blocked
func (r *Resolver) blockedProjectsOfUser(ctx context.Context, userID string) blockedSet {
	rows, err := r.DB.Queries.GetConsentBlockedProjectsOfUsers(ctx, []string{userID})
	if err != nil {
		slog.Error("failed to list consent-blocked projects of user", "userID", userID, "error", err)
		return blockedSet{}
	}
	blocked := make([]projectUser, len(rows))
	for i, row := range rows {
		blocked[i] = projectUser{ProjectID: row.ProjectID, UserID: row.UserID}
	}
	return blockedSet{Users: blocked, Known: true}
}

// blockedUsersInConsentScope lists the blocked users of the consent's project,
// or of every project for a global consent
func (r *Resolver) blockedUsersInConsentScope(ctx context.Context, projectID *string) blockedSet {
	var projectIDs []string // nil: all projects
	if projectID != nil {
		projectIDs = []string{*projectID}
	}
	rows, err := r.DB.Queries.GetConsentBlockedUsersInProjects(ctx, projectIDs)
	if err != nil {
		slog.Error("failed to list consent-blocked users", "error", err)
		return blockedSet{}
	}
	blocked := make([]projectUser, len(rows))
	for i, row := range rows {
		blocked[i] = projectUser{ProjectID: row.ProjectID, UserID: row.UserID}
	}
	return blockedSet{Users: blocked, Known: true}
}

// consentBlockChangesBetween returns who a consent change blocked or
// unblocked, and false when a listing failed and that is unknown
func consentBlockChangesBetween(before, after blockedSet) (map[projectUser]bool, bool) {
	if !before.Known || !after.Known {
		return nil, false
	}
	return consentBlockChanges(before.Users, after.Users), true
}

// applyConsentBlockChanges applies what a consent change did to who counts,
// from who was blocked before and after it. It runs for every consent change
// and compares rather than trusting isRequired, so it only acts on who
// actually moved.
func (r *Resolver) applyConsentBlockChanges(ctx context.Context, before, after blockedSet) {
	changed, known := consentBlockChangesBetween(before, after)
	if !known {
		// Who moved is unknown; at least don't serve boards counting them wrongly
		r.Cache.InvalidateLeaderboardScores()
		return
	}
	if len(changed) == 0 {
		return
	}
	r.Cache.InvalidateLeaderboardScores()
	r.refreshTeamsOfBlockChanges(ctx, changed)
}

// refreshTeamsOfBlockChanges re-applies the average-age exclusion to the teams
// of users a consent change blocked or unblocked, and drops the cached member
// counts. Other teams keep their flag, which an admin may have set by hand.
func (r *Resolver) refreshTeamsOfBlockChanges(ctx context.Context, changed map[projectUser]bool) {

	userIDs := make([]string, 0, len(changed))
	seen := make(map[string]bool, len(changed))
	for pu := range changed {
		if !seen[pu.UserID] {
			seen[pu.UserID] = true
			userIDs = append(userIDs, pu.UserID)
		}
	}
	memberships, err := r.DB.Queries.GetTeamsByUserIDs(ctx, userIDs)
	if err != nil {
		slog.Error("failed to load teams of users a consent change affected", "error", err)
	} else {
		for teamID, projectID := range teamsOfChangedUsers(memberships, changed) {
			r.refreshTeamLeaderboardExclusion(ctx, teamID, projectID)
		}
	}

	// teams/superTeams filter by member counts, which leave blocked users out
	r.Cache.DeletePrefix(cache.PrefixTeamsFilter)
	r.Cache.DeletePrefix(cache.PrefixTeamsCount)
	r.Cache.DeletePrefix(cache.PrefixSuperTeamsFilter)
	r.Cache.DeletePrefix(cache.PrefixSuperTeamsCount)
}

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
