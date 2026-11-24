package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// consentBlockedIDsByTeamBatchFunc batches the members of each team hidden by a
// required consent of its project (00116). Deliberately not cached in
// Ristretto: accepting a consent changes it, so it is only batched per request
// window.
func consentBlockedIDsByTeamBatchFunc(db *database.DB) func(context.Context, []string) []*dataloader.Result[[]string] {
	return func(ctx context.Context, teamIDs []string) []*dataloader.Result[[]string] {
		rows, err := db.Queries.GetConsentBlockedUserIDsInTeams(ctx, teamIDs)
		if err != nil {
			results := make([]*dataloader.Result[[]string], len(teamIDs))
			for i := range results {
				results[i] = &dataloader.Result[[]string]{Error: err}
			}
			return results
		}

		return mapConsentBlockedIDsByTeam(teamIDs, rows)
	}
}

// mapConsentBlockedIDsByTeam groups the rows by team, in the order of teamIDs;
// a team without blocked members gets an empty slice
func mapConsentBlockedIDsByTeam(teamIDs []string, rows []*sqlc.GetConsentBlockedUserIDsInTeamsRow) []*dataloader.Result[[]string] {
	byTeam := make(map[string][]string, len(teamIDs))
	for _, row := range rows {
		byTeam[row.TeamID] = append(byTeam[row.TeamID], row.UserID)
	}

	results := make([]*dataloader.Result[[]string], len(teamIDs))
	for i, teamID := range teamIDs {
		ids := byTeam[teamID]
		if ids == nil {
			ids = []string{}
		}
		results[i] = &dataloader.Result[[]string]{Data: ids}
	}
	return results
}
