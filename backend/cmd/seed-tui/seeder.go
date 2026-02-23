package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5"
	"github.com/jaswdr/faker"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("/tmp/seed-tui.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
}

func logDebug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logFile, "[%s] %s\n", time.Now().Format("15:04:05.000"), msg)
	logFile.Sync()
}

// TeamSeeder handles the seeding of teams and related data
type TeamSeeder struct {
	ctx       context.Context
	db        *database.DB
	projectID string
	config    SeedConfig
	fake      faker.Faker
}

// NewTeamSeeder creates a new TeamSeeder
func NewTeamSeeder(ctx context.Context, db *database.DB, projectID string, config SeedConfig) *TeamSeeder {
	return &TeamSeeder{
		ctx:       ctx,
		db:        db,
		projectID: projectID,
		config:    config,
		fake:      faker.New(),
	}
}

// RunSync executes the seeding process synchronously and returns the result
func (s *TeamSeeder) RunSync() (SeedResult, error) {
	var result SeedResult

	logDebug("Starting RunSync for project %s", s.projectID)

	// Step 1: Get available users
	logDebug("Step 1: Getting available users...")
	availableUserIDs, err := s.db.Queries.GetUsersNotInTeamForProject(s.ctx, s.projectID)
	if err != nil {
		logDebug("Error getting available users: %v", err)
		return result, fmt.Errorf("failed to get available users: %w", err)
	}
	logDebug("Found %d available users", len(availableUserIDs))

	usersNeeded := s.config.TeamCount * s.config.TeamSize
	logDebug("Users needed: %d (teams=%d, size=%d)", usersNeeded, s.config.TeamCount, s.config.TeamSize)

	// Step 2: Ensure we have enough users
	logDebug("Step 2: Ensuring enough users...")
	userIDs, usersCreated, err := s.ensureEnoughUsersSync(availableUserIDs, usersNeeded)
	if err != nil {
		logDebug("Error ensuring users: %v", err)
		return result, fmt.Errorf("failed to ensure users: %w", err)
	}
	logDebug("Step 2 done: %d users ready, %d created", len(userIDs), usersCreated)
	result.UsersCreated = usersCreated

	// Step 3: Add users to project
	logDebug("Step 3: Adding users to project...")
	if err := s.addUsersToProject(userIDs); err != nil {
		logDebug("Error adding users to project: %v", err)
		return result, fmt.Errorf("failed to add users to project: %w", err)
	}
	logDebug("Step 3 done: users added to project")

	// Step 4: Create teams
	logDebug("Step 4: Creating %d teams...", s.config.TeamCount)
	teamIDs, err := s.createTeamsSync()
	if err != nil {
		logDebug("Error creating teams: %v", err)
		return result, fmt.Errorf("failed to create teams: %w", err)
	}
	logDebug("Step 4 done: %d teams created", len(teamIDs))
	result.TeamsCreated = len(teamIDs)

	// Step 5: Assign users to teams
	logDebug("Step 5: Assigning %d users to teams...", len(userIDs))
	membersAssigned, teamUserMap, err := s.assignUsersToTeamsSync(userIDs, teamIDs)
	if err != nil {
		logDebug("Error assigning users to teams: %v", err)
		return result, fmt.Errorf("failed to assign users to teams: %w", err)
	}
	logDebug("Step 5 done: %d members assigned", membersAssigned)
	result.MembersAssigned = membersAssigned

	// Step 6: Assign team leads (first member of each team)
	logDebug("Step 6: Assigning team leads...")
	leadsAssigned, err := s.assignTeamLeadsSync(teamUserMap)
	if err != nil {
		logDebug("Error assigning team leads: %v", err)
		return result, fmt.Errorf("failed to assign team leads: %w", err)
	}
	logDebug("Step 6 done: %d team leads assigned", leadsAssigned)
	result.TeamLeadsAssigned = leadsAssigned

	// Step 7: Generate points for users
	logDebug("Step 7: Generating points for %d users (min=%d, max=%d)...", len(userIDs), s.config.MinPoints, s.config.MaxPoints)
	pointsGenerated, err := s.generatePointsForUsersSync(userIDs)
	if err != nil {
		logDebug("Error generating points: %v", err)
		return result, fmt.Errorf("failed to generate points: %w", err)
	}
	logDebug("Step 7 done: %d total points generated", pointsGenerated)
	result.PointsGenerated = pointsGenerated

	return result, nil
}

// Run executes the seeding process and reports progress
func (s *TeamSeeder) Run(progressChan chan<- SeedProgress) error {
	progress := SeedProgress{}

	// Step 1: Get available users
	progress.Stage = "Loading available users..."
	progressChan <- progress

	availableUserIDs, err := s.db.Queries.GetUsersNotInTeamForProject(s.ctx, s.projectID)
	if err != nil {
		return fmt.Errorf("failed to get available users: %w", err)
	}

	usersNeeded := s.config.TeamCount * s.config.TeamSize

	// Step 2: Ensure we have enough users
	userIDs, usersCreated, err := s.ensureEnoughUsers(availableUserIDs, usersNeeded, progressChan)
	if err != nil {
		return fmt.Errorf("failed to ensure users: %w", err)
	}
	progress.UsersCreated = usersCreated

	// Step 3: Add users to project
	progress.Stage = "Adding users to project..."
	progressChan <- progress

	if err := s.addUsersToProject(userIDs); err != nil {
		return fmt.Errorf("failed to add users to project: %w", err)
	}

	// Step 4: Create teams
	progress.Stage = "Creating teams..."
	progress.Total = s.config.TeamCount
	progressChan <- progress

	teamIDs, err := s.createTeams(func(current int) {
		progress.Current = current
		progress.TeamsCreated = current
		progressChan <- progress
	})
	if err != nil {
		return fmt.Errorf("failed to create teams: %w", err)
	}

	// Step 5: Assign users to teams
	progress.Stage = "Assigning users to teams..."
	progress.Current = 0
	progress.Total = len(userIDs)
	progressChan <- progress

	membersAssigned, err := s.assignUsersToTeams(userIDs, teamIDs, func(current int) {
		progress.Current = current
		progress.MembersAssigned = current
		progressChan <- progress
	})
	if err != nil {
		return fmt.Errorf("failed to assign users to teams: %w", err)
	}
	progress.MembersAssigned = membersAssigned

	// Step 6: Generate points for users
	progress.Stage = "Generating points..."
	progress.Current = 0
	progress.Total = len(userIDs)
	progressChan <- progress

	pointsGenerated, err := s.generatePointsForUsers(userIDs, func(current, points int) {
		progress.Current = current
		progress.PointsGenerated = points
		progressChan <- progress
	})
	if err != nil {
		return fmt.Errorf("failed to generate points: %w", err)
	}
	progress.PointsGenerated = pointsGenerated

	progress.Stage = "Complete"
	progressChan <- progress

	return nil
}

// ensureEnoughUsers returns the user IDs to use, creating new users if needed
func (s *TeamSeeder) ensureEnoughUsers(availableUserIDs []string, needed int, progressChan chan<- SeedProgress) ([]string, int, error) {
	if len(availableUserIDs) >= needed {
		// Shuffle and return the first 'needed' users
		rand.Shuffle(len(availableUserIDs), func(i, j int) {
			availableUserIDs[i], availableUserIDs[j] = availableUserIDs[j], availableUserIDs[i]
		})
		return availableUserIDs[:needed], 0, nil
	}

	// We need to create more users
	toCreate := needed - len(availableUserIDs)
	progress := SeedProgress{
		Stage: "Creating new users...",
		Total: toCreate,
	}
	progressChan <- progress

	newUserIDs, err := s.createUsers(toCreate, func(current int) {
		progress.Current = current
		progress.UsersCreated = current
		progressChan <- progress
	})
	if err != nil {
		return nil, 0, err
	}

	// Combine available and new user IDs
	allUserIDs := append(availableUserIDs, newUserIDs...)

	// Shuffle the combined list
	rand.Shuffle(len(allUserIDs), func(i, j int) {
		allUserIDs[i], allUserIDs[j] = allUserIDs[j], allUserIDs[i]
	})

	return allUserIDs[:needed], toCreate, nil
}

// createUsers creates new users and returns their IDs
func (s *TeamSeeder) createUsers(count int, onProgress func(int)) ([]string, error) {
	// Get churches to assign users to
	churchCount, err := s.db.Queries.GetTotalChurchCount(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get church count: %w", err)
	}

	if churchCount == 0 {
		return nil, fmt.Errorf("no churches found in database")
	}

	// Get all church IDs (or a subset if there are many)
	limitCount := int32(100)
	if churchCount < limitCount {
		limitCount = churchCount
	}
	churchIDs, err := s.db.Queries.GetRandomChurchIDs(s.ctx, limitCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get church IDs: %w", err)
	}

	genders := []string{"MALE", "FEMALE"}
	batchSize := 1000
	userIDs := make([]string, 0, count)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		batchRows := make([][]interface{}, 0, batchEnd-batchStart)

		for i := batchStart; i < batchEnd; i++ {
			gender := genders[rand.Intn(len(genders))]

			var firstName, lastName string
			if gender == "MALE" {
				firstName = s.fake.Person().FirstNameMale()
			} else {
				firstName = s.fake.Person().FirstNameFemale()
			}
			lastName = s.fake.Person().LastName()
			displayName := firstName + " " + lastName

			// Random birthdate (age 13-80)
			age := 13 + rand.Intn(68)
			now := time.Now()
			birthdate := now.AddDate(-age, -rand.Intn(12), -rand.Intn(28))

			// Random church
			churchID := churchIDs[rand.Intn(len(churchIDs))]

			// Generate members ID
			membersID := fmt.Sprintf("SEED-%d-%d", time.Now().UnixNano(), i)

			// Generate email
			email := s.fake.Internet().Email()

			// Avatar URL
			avatarURL := fmt.Sprintf("https://i.pravatar.cc/150?img=%d", (i%70)+1)

			id := ulid.NewUserID()
			userIDs = append(userIDs, id)

			batchRows = append(batchRows, []interface{}{
				id,
				membersID,
				email,
				displayName,
				firstName,
				lastName,
				nil, // middle_name
				displayName,
				gender,
				birthdate,
				churchID,
				avatarURL,
			})
		}

		// Batch insert
		_, err := s.db.Pool.CopyFrom(
			s.ctx,
			pgx.Identifier{"users"},
			[]string{"id", "members_id", "email", "name", "first_name", "last_name", "middle_name", "display_name", "gender", "birthdate", "church_id", "avatar_url"},
			pgx.CopyFromRows(batchRows),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert users batch: %w", err)
		}

		onProgress(batchEnd)
	}

	return userIDs, nil
}

// addUsersToProject adds users to the project's user_projects table
func (s *TeamSeeder) addUsersToProject(userIDs []string) error {
	batchSize := 1000

	for batchStart := 0; batchStart < len(userIDs); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > len(userIDs) {
			batchEnd = len(userIDs)
		}

		batchRows := make([][]interface{}, 0, batchEnd-batchStart)
		for _, userID := range userIDs[batchStart:batchEnd] {
			batchRows = append(batchRows, []interface{}{userID, s.projectID})
		}

		_, err := s.db.Pool.CopyFrom(
			s.ctx,
			pgx.Identifier{"user_projects"},
			[]string{"user_id", "project_id"},
			pgx.CopyFromRows(batchRows),
		)
		if err != nil {
			// Ignore duplicate key errors (users might already be in project)
			// The ON CONFLICT clause isn't available with CopyFrom, so we just skip errors
			continue
		}
	}

	return nil
}

// createTeams creates teams for the project and returns their IDs
func (s *TeamSeeder) createTeams(onProgress func(int)) ([]string, error) {
	batchSize := 500
	teamIDs := make([]string, 0, s.config.TeamCount)

	for batchStart := 0; batchStart < s.config.TeamCount; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > s.config.TeamCount {
			batchEnd = s.config.TeamCount
		}

		batchRows := make([][]interface{}, 0, batchEnd-batchStart)

		for i := batchStart; i < batchEnd; i++ {
			teamID := ulid.NewTeamID()
			teamIDs = append(teamIDs, teamID)

			name := fmt.Sprintf("%s %s", s.fake.Company().Name(), "Team")
			description := s.fake.Lorem().Sentence(8)
			joinCode := teamID[2:] // Use ULID part as join code

			batchRows = append(batchRows, []interface{}{
				teamID,
				s.projectID,
				nil, // super_team_id
				name,
				description,
				joinCode,
			})
		}

		_, err := s.db.Pool.CopyFrom(
			s.ctx,
			pgx.Identifier{"teams"},
			[]string{"id", "project_id", "super_team_id", "name", "description", "join_code"},
			pgx.CopyFromRows(batchRows),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert teams batch: %w", err)
		}

		onProgress(batchEnd)
	}

	return teamIDs, nil
}

// assignUsersToTeams assigns users to teams evenly
func (s *TeamSeeder) assignUsersToTeams(userIDs, teamIDs []string, onProgress func(int)) (int, error) {
	batchSize := 1000
	batchRows := make([][]interface{}, 0, batchSize)
	assigned := 0

	for i, userID := range userIDs {
		teamIdx := i / s.config.TeamSize
		if teamIdx >= len(teamIDs) {
			break
		}

		teamID := teamIDs[teamIdx]
		batchRows = append(batchRows, []interface{}{teamID, userID})
		assigned++

		if len(batchRows) >= batchSize || i == len(userIDs)-1 {
			_, err := s.db.Pool.CopyFrom(
				s.ctx,
				pgx.Identifier{"team_members"},
				[]string{"team_id", "user_id"},
				pgx.CopyFromRows(batchRows),
			)
			if err != nil {
				return assigned, fmt.Errorf("failed to insert team members batch: %w", err)
			}

			batchRows = batchRows[:0]
			onProgress(assigned)
		}
	}

	return assigned, nil
}

// generatePointsForUsers creates score_journal entries for each user
func (s *TeamSeeder) generatePointsForUsers(userIDs []string, onProgress func(current, totalPoints int)) (int, error) {
	batchSize := 1000
	batchRows := make([][]interface{}, 0, batchSize)
	totalPoints := 0

	manualReasons := []string{
		"Bonus for participation",
		"Event attendance reward",
		"Extra credit for team spirit",
		"Special recognition award",
		"Encouragement bonus",
		"Leadership contribution",
		"Helping other participants",
		"Outstanding engagement",
	}

	for i, userID := range userIDs {
		// Generate random total points for this user
		targetPoints := s.config.MinPoints + rand.Intn(s.config.MaxPoints-s.config.MinPoints+1)

		// Split into multiple entries of varying sizes
		remainingPoints := targetPoints
		for remainingPoints > 0 {
			// Random entry size between 50 and 500 (or remaining if less)
			entryPoints := 50 + rand.Intn(451)
			if entryPoints > remainingPoints {
				entryPoints = remainingPoints
			}
			remainingPoints -= entryPoints

			entryID := ulid.NewScoreJournalID()
			reason := manualReasons[rand.Intn(len(manualReasons))]

			// Random created_at in the last 30 days
			createdAt := time.Now().AddDate(0, 0, -rand.Intn(30))

			batchRows = append(batchRows, []interface{}{
				entryID,
				s.projectID,
				userID,
				nil,              // event_id
				nil,              // challenge_id
				int32(entryPoints),
				"MANUAL",         // source_type
				nil,              // source_id
				reason,           // reason
				nil,              // awarded_by
				createdAt,        // created_at
			})

			totalPoints += entryPoints

			if len(batchRows) >= batchSize {
				if err := s.insertScoreJournalBatch(batchRows); err != nil {
					return totalPoints, err
				}
				batchRows = batchRows[:0]
			}
		}

		if (i+1)%100 == 0 || i == len(userIDs)-1 {
			onProgress(i+1, totalPoints)
		}
	}

	// Insert remaining rows
	if len(batchRows) > 0 {
		if err := s.insertScoreJournalBatch(batchRows); err != nil {
			return totalPoints, err
		}
	}

	return totalPoints, nil
}

// insertScoreJournalBatch inserts a batch of score_journal entries
func (s *TeamSeeder) insertScoreJournalBatch(rows [][]interface{}) error {
	_, err := s.db.Pool.CopyFrom(
		s.ctx,
		pgx.Identifier{"score_journal"},
		[]string{"id", "project_id", "user_id", "event_id", "challenge_id", "points", "source_type", "source_id", "reason", "awarded_by", "created_at"},
		pgx.CopyFromRows(rows),
	)
	return err
}

// Sync versions of methods (no progress callbacks)

// ensureEnoughUsersSync returns the user IDs to use, creating new users if needed
func (s *TeamSeeder) ensureEnoughUsersSync(availableUserIDs []string, needed int) ([]string, int, error) {
	if len(availableUserIDs) >= needed {
		rand.Shuffle(len(availableUserIDs), func(i, j int) {
			availableUserIDs[i], availableUserIDs[j] = availableUserIDs[j], availableUserIDs[i]
		})
		return availableUserIDs[:needed], 0, nil
	}

	toCreate := needed - len(availableUserIDs)
	newUserIDs, err := s.createUsersSync(toCreate)
	if err != nil {
		return nil, 0, err
	}

	allUserIDs := append(availableUserIDs, newUserIDs...)
	rand.Shuffle(len(allUserIDs), func(i, j int) {
		allUserIDs[i], allUserIDs[j] = allUserIDs[j], allUserIDs[i]
	})

	return allUserIDs[:needed], toCreate, nil
}

// createUsersSync creates new users and returns their IDs
func (s *TeamSeeder) createUsersSync(count int) ([]string, error) {
	churchCount, err := s.db.Queries.GetTotalChurchCount(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get church count: %w", err)
	}

	if churchCount == 0 {
		return nil, fmt.Errorf("no churches found in database")
	}

	limitCount := int32(100)
	if churchCount < limitCount {
		limitCount = churchCount
	}
	churchIDs, err := s.db.Queries.GetRandomChurchIDs(s.ctx, limitCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get church IDs: %w", err)
	}

	genders := []string{"MALE", "FEMALE"}
	batchSize := 1000
	userIDs := make([]string, 0, count)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		batchRows := make([][]interface{}, 0, batchEnd-batchStart)

		for i := batchStart; i < batchEnd; i++ {
			gender := genders[rand.Intn(len(genders))]

			var firstName, lastName string
			if gender == "MALE" {
				firstName = s.fake.Person().FirstNameMale()
			} else {
				firstName = s.fake.Person().FirstNameFemale()
			}
			lastName = s.fake.Person().LastName()
			displayName := firstName + " " + lastName

			age := 13 + rand.Intn(68)
			now := time.Now()
			birthdate := now.AddDate(-age, -rand.Intn(12), -rand.Intn(28))

			churchID := churchIDs[rand.Intn(len(churchIDs))]
			membersID := fmt.Sprintf("SEED-%d-%d", time.Now().UnixNano(), i)
			email := s.fake.Internet().Email()
			avatarURL := fmt.Sprintf("https://i.pravatar.cc/150?img=%d", (i%70)+1)

			id := ulid.NewUserID()
			userIDs = append(userIDs, id)

			batchRows = append(batchRows, []interface{}{
				id, membersID, email, displayName, firstName, lastName, nil, displayName, gender, birthdate, churchID, avatarURL,
			})
		}

		_, err := s.db.Pool.CopyFrom(
			s.ctx,
			pgx.Identifier{"users"},
			[]string{"id", "members_id", "email", "name", "first_name", "last_name", "middle_name", "display_name", "gender", "birthdate", "church_id", "avatar_url"},
			pgx.CopyFromRows(batchRows),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert users batch: %w", err)
		}
	}

	return userIDs, nil
}

// createTeamsSync creates teams for the project and returns their IDs
func (s *TeamSeeder) createTeamsSync() ([]string, error) {
	batchSize := 500
	teamIDs := make([]string, 0, s.config.TeamCount)

	for batchStart := 0; batchStart < s.config.TeamCount; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > s.config.TeamCount {
			batchEnd = s.config.TeamCount
		}

		logDebug("Creating teams batch %d-%d...", batchStart, batchEnd)

		batchRows := make([][]interface{}, 0, batchEnd-batchStart)

		for i := batchStart; i < batchEnd; i++ {
			teamID := ulid.NewTeamID()
			teamIDs = append(teamIDs, teamID)

			name := fmt.Sprintf("%s %s", s.fake.Company().Name(), "Team")
			description := s.fake.Lorem().Sentence(8)
			joinCode := teamID[2:]

			batchRows = append(batchRows, []interface{}{
				teamID, s.projectID, nil, name, description, joinCode,
			})
		}

		_, err := s.db.Pool.CopyFrom(
			s.ctx,
			pgx.Identifier{"teams"},
			[]string{"id", "project_id", "super_team_id", "name", "description", "join_code"},
			pgx.CopyFromRows(batchRows),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert teams batch: %w", err)
		}
	}

	return teamIDs, nil
}

// assignUsersToTeamsSync assigns users to teams evenly and returns team -> first user map
func (s *TeamSeeder) assignUsersToTeamsSync(userIDs, teamIDs []string) (int, map[string]string, error) {
	batchSize := 1000
	batchRows := make([][]interface{}, 0, batchSize)
	assigned := 0
	teamFirstUser := make(map[string]string) // team_id -> first user_id (for team lead)

	for i, userID := range userIDs {
		teamIdx := i / s.config.TeamSize
		if teamIdx >= len(teamIDs) {
			break
		}

		teamID := teamIDs[teamIdx]

		// Track the first user assigned to each team (will be team lead)
		if _, exists := teamFirstUser[teamID]; !exists {
			teamFirstUser[teamID] = userID
		}

		batchRows = append(batchRows, []interface{}{teamID, userID})
		assigned++

		if len(batchRows) >= batchSize || i == len(userIDs)-1 {
			_, err := s.db.Pool.CopyFrom(
				s.ctx,
				pgx.Identifier{"team_members"},
				[]string{"team_id", "user_id"},
				pgx.CopyFromRows(batchRows),
			)
			if err != nil {
				return assigned, nil, fmt.Errorf("failed to insert team members batch: %w", err)
			}
			batchRows = batchRows[:0]
		}
	}

	return assigned, teamFirstUser, nil
}

// assignTeamLeadsSync assigns team leads from the team->user map
func (s *TeamSeeder) assignTeamLeadsSync(teamUserMap map[string]string) (int, error) {
	batchSize := 500
	batchRows := make([][]interface{}, 0, batchSize)
	assigned := 0
	now := time.Now()

	for teamID, userID := range teamUserMap {
		roleID := ulid.NewUserRoleID()
		batchRows = append(batchRows, []interface{}{
			roleID,
			userID,
			"TEAM_LEAD",
			nil,    // church_id
			nil,    // project_id
			teamID,
			userID, // assigned_by (self-assignment for seeding)
			now,    // assigned_at
		})
		assigned++

		if len(batchRows) >= batchSize {
			_, err := s.db.Pool.CopyFrom(
				s.ctx,
				pgx.Identifier{"user_roles"},
				[]string{"id", "user_id", "role", "church_id", "project_id", "team_id", "assigned_by", "assigned_at"},
				pgx.CopyFromRows(batchRows),
			)
			if err != nil {
				return assigned, fmt.Errorf("failed to insert team leads batch: %w", err)
			}
			batchRows = batchRows[:0]
		}
	}

	// Insert remaining rows
	if len(batchRows) > 0 {
		_, err := s.db.Pool.CopyFrom(
			s.ctx,
			pgx.Identifier{"user_roles"},
			[]string{"id", "user_id", "role", "church_id", "project_id", "team_id", "assigned_by", "assigned_at"},
			pgx.CopyFromRows(batchRows),
		)
		if err != nil {
			return assigned, fmt.Errorf("failed to insert team leads batch: %w", err)
		}
	}

	return assigned, nil
}

// ReassignTeamLeadsRandomly removes all team leads for a project and assigns new random ones
func (s *TeamSeeder) ReassignTeamLeadsRandomly() (int, error) {
	logDebug("Removing existing team leads for project %s...", s.projectID)

	// Delete all existing team leads for this project
	if err := s.db.Queries.DeleteTeamLeadsForProject(s.ctx, s.projectID); err != nil {
		return 0, fmt.Errorf("failed to delete existing team leads: %w", err)
	}

	// Get all team IDs for the project
	teamIDs, err := s.db.Queries.GetTeamIDsForProject(s.ctx, s.projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get team IDs: %w", err)
	}
	logDebug("Found %d teams to assign leads to", len(teamIDs))

	if len(teamIDs) == 0 {
		return 0, nil
	}

	// Get a random member for each team
	randomMembers, err := s.db.Queries.GetRandomMemberForTeams(s.ctx, teamIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to get random members: %w", err)
	}
	logDebug("Got %d random members for teams", len(randomMembers))

	// Build team -> user map
	teamUserMap := make(map[string]string)
	for _, m := range randomMembers {
		teamUserMap[m.TeamID] = m.UserID
	}

	// Assign the new team leads
	assigned, err := s.assignTeamLeadsSync(teamUserMap)
	if err != nil {
		return 0, err
	}

	logDebug("Assigned %d team leads", assigned)
	return assigned, nil
}

// generatePointsForUsersSync creates score_journal entries for each user
func (s *TeamSeeder) generatePointsForUsersSync(userIDs []string) (int, error) {
	batchSize := 5000 // Larger batches for speed
	batchRows := make([][]interface{}, 0, batchSize)
	totalPoints := 0
	entriesCreated := 0

	manualReasons := []string{
		"Bonus for participation",
		"Event attendance reward",
		"Extra credit for team spirit",
		"Special recognition award",
		"Encouragement bonus",
		"Leadership contribution",
		"Helping other participants",
		"Outstanding engagement",
	}

	for i, userID := range userIDs {
		targetPoints := s.config.MinPoints + rand.Intn(s.config.MaxPoints-s.config.MinPoints+1)

		remainingPoints := targetPoints
		for remainingPoints > 0 {
			entryPoints := 50 + rand.Intn(451)
			if entryPoints > remainingPoints {
				entryPoints = remainingPoints
			}
			remainingPoints -= entryPoints

			entryID := ulid.NewScoreJournalID()
			reason := manualReasons[rand.Intn(len(manualReasons))]
			createdAt := time.Now().AddDate(0, 0, -rand.Intn(30))

			batchRows = append(batchRows, []interface{}{
				entryID, s.projectID, userID, nil, nil, int32(entryPoints), "MANUAL", nil, reason, nil, createdAt,
			})

			totalPoints += entryPoints
			entriesCreated++

			if len(batchRows) >= batchSize {
				logDebug("Inserting batch of %d score entries (user %d/%d, %d entries so far)...", len(batchRows), i+1, len(userIDs), entriesCreated)
				if err := s.insertScoreJournalBatch(batchRows); err != nil {
					return totalPoints, err
				}
				batchRows = batchRows[:0]
			}
		}
	}

	if len(batchRows) > 0 {
		logDebug("Inserting final batch of %d score entries...", len(batchRows))
		if err := s.insertScoreJournalBatch(batchRows); err != nil {
			return totalPoints, err
		}
	}

	logDebug("Total score entries created: %d", entriesCreated)
	return totalPoints, nil
}
