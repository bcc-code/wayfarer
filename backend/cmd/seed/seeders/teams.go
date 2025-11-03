package seeders

import (
	"math/rand"

	"github.com/bcc-media/wayfarer/internal/ulid"
)

// SeedTeams creates super teams, teams, and assigns users
func (s *Seeder) SeedTeams(stats *Stats) error {
	superTeamQuery := `
		INSERT INTO super_teams (id, project_id, name, description)
		VALUES ($1, $2, $3, $4)
	`

	teamQuery := `
		INSERT INTO teams (id, project_id, super_team_id, name, description, join_code)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	memberQuery := `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2)
	`

	userProjectQuery := `
		INSERT INTO user_projects (user_id, project_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	// For each project, create super teams and teams
	for _, projectID := range s.Data.ProjectIDs {
		// Create 2-3 SuperTeams per project
		numSuperTeams := 2 + rand.Intn(2)
		for i := 0; i < numSuperTeams; i++ {
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

		// Create 8-12 teams per project
		numTeams := 8 + rand.Intn(5)
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

			// Assign 5-15 random users to this team
			numMembers := 5 + rand.Intn(11)
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
					continue // Skip if we couldn't find a unique user
				}

				assignedUsers[userID] = true

				// Add user to project
				_, err := s.DB.Pool.Exec(s.Ctx, userProjectQuery, userID, projectID)
				if err != nil {
					return err
				}

				// Add user to team
				_, err = s.DB.Pool.Exec(s.Ctx, memberQuery, teamID, userID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
