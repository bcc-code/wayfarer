package seeders

import (
	"log/slog"
	"math/rand"

	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5"
)

// SeedTeams creates super teams, teams, and assigns users based on configured participation rate
func (s *Seeder) SeedTeams(stats *Stats) error {
	superTeamQuery := `
		INSERT INTO super_teams (id, project_id, name, description)
		VALUES ($1, $2, $3, $4)
	`

	teamQuery := `
		INSERT INTO teams (id, project_id, super_team_id, name, description, join_code)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// First, assign users to projects based on participation rate
	userProjectRows := [][]interface{}{}
	for _, projectID := range s.Data.ProjectIDs {
		for _, userID := range s.Data.UserIDs {
			if rand.Float64() < s.Config.ProjectParticipationRate {
				userProjectRows = append(userProjectRows, []interface{}{userID, projectID})
			}
		}
	}

	// Batch insert user_projects
	if len(userProjectRows) > 0 {
		_, err := s.DB.Pool.CopyFrom(
			s.Ctx,
			pgx.Identifier{"user_projects"},
			[]string{"user_id", "project_id"},
			pgx.CopyFromRows(userProjectRows),
		)
		if err != nil {
			return err
		}
		slog.Info("User project assignments", "count", len(userProjectRows))
	}

	// For each project, create super teams and teams
	projectCount := 0
	for _, projectID := range s.Data.ProjectIDs {
		projectCount++

		// Create configured number of SuperTeams per project
		for i := 0; i < s.Config.NumSuperTeams; i++ {
			superTeamID := ulid.NewSuperTeamID()
			name := s.Fake.Company().Name() + " Alliance"
			description := s.Fake.Lorem().Sentence(8)

			_, err := s.DB.Pool.Exec(s.Ctx, superTeamQuery,
				superTeamID,
				projectID,
				name,
				description,
			)
			if err != nil {
				return err
			}

			s.Data.SuperTeamIDs[projectID] = append(s.Data.SuperTeamIDs[projectID], superTeamID)
			stats.SuperTeams++
		}

		// Calculate expected users in this project
		expectedUsersInProject := int(float64(len(s.Data.UserIDs)) * s.Config.ProjectParticipationRate)

		// Calculate number of teams needed to have ~TeamSize members each
		numTeams := expectedUsersInProject / s.Config.TeamSize
		if numTeams < 1 {
			numTeams = 1
		}

		// Create teams
		teamMemberRows := [][]interface{}{}
		for i := 0; i < numTeams; i++ {
			teamID := ulid.NewTeamID()
			name := s.Fake.Company().JobTitle() + " Team"
			description := s.Fake.Lorem().Sentence(10)

			// Generate unique join code - use the full team ID without prefix
			joinCode := teamID[2:]

			// 50% chance to be part of a super team
			var superTeamID *string
			if rand.Float32() < 0.5 && len(s.Data.SuperTeamIDs[projectID]) > 0 {
				stID := s.Data.SuperTeamIDs[projectID][rand.Intn(len(s.Data.SuperTeamIDs[projectID]))]
				superTeamID = &stID
			}

			_, err := s.DB.Pool.Exec(s.Ctx, teamQuery,
				teamID,
				projectID,
				superTeamID,
				name,
				description,
				joinCode,
			)
			if err != nil {
				return err
			}

			s.Data.TeamIDs[projectID] = append(s.Data.TeamIDs[projectID], teamID)
			stats.Teams++

			// Assign approximately TeamSize random users to this team
			numMembers := s.Config.TeamSize + rand.Intn(3) - 1 // TeamSize ± 1
			assignedUsers := make(map[string]bool)

			for j := 0; j < numMembers && j < len(s.Data.UserIDs); j++ {
				// Pick a random user that hasn't been assigned to this team yet
				var userID string
				for attempts := 0; attempts < 10; attempts++ {
					userID = s.Data.UserIDs[rand.Intn(len(s.Data.UserIDs))]
					if !assignedUsers[userID] {
						break
					}
				}

				if assignedUsers[userID] {
					continue
				}

				assignedUsers[userID] = true
				teamMemberRows = append(teamMemberRows, []interface{}{teamID, userID})
			}
		}

		// Batch insert team members for this project
		if len(teamMemberRows) > 0 {
			_, err := s.DB.Pool.CopyFrom(
				s.Ctx,
				pgx.Identifier{"team_members"},
				[]string{"team_id", "user_id"},
				pgx.CopyFromRows(teamMemberRows),
			)
			if err != nil {
				return err
			}
		}

		if projectCount%10 == 0 {
			slog.Info("Teams progress", "projects_completed", projectCount, "total_projects", len(s.Data.ProjectIDs), "teams_created", stats.Teams, "superteams_created", stats.SuperTeams)
		}
	}

	return nil
}
