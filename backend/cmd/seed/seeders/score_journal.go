package seeders

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5"
)

// SeedScoreJournal creates score_journal entries for users
// Each user gets at least 3 entries distributed across their projects
// 80-90% ACHIEVEMENT type, 10-20% MANUAL type, all positive points
func (s *Seeder) SeedScoreJournal(stats *Stats) error {
	slog.Info("Starting score_journal seeding")

	// Query achievements with their details for creating ACHIEVEMENT entries
	achievementQuery := `
		SELECT id, project_id, event_id, challenge_id, points
		FROM achievements
	`
	rows, err := s.DB.Pool.Query(s.Ctx, achievementQuery)
	if err != nil {
		return fmt.Errorf("failed to query achievements: %w", err)
	}
	defer rows.Close()

	// Map of projectID -> []Achievement for easy lookup
	type Achievement struct {
		ID          string
		ProjectID   string
		EventID     *string
		ChallengeID *string
		Points      int32
	}
	achievementsByProject := make(map[string][]Achievement)

	for rows.Next() {
		var a Achievement
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.EventID, &a.ChallengeID, &a.Points); err != nil {
			return fmt.Errorf("failed to scan achievement: %w", err)
		}
		achievementsByProject[a.ProjectID] = append(achievementsByProject[a.ProjectID], a)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating achievements: %w", err)
	}

	slog.Info("Loaded achievements for score journal", "total_achievements", len(achievementsByProject))

	// Get all admin users for awarding manual entries
	adminQuery := `
		SELECT id FROM users LIMIT 10
	`
	adminRows, err := s.DB.Pool.Query(s.Ctx, adminQuery)
	if err != nil {
		return fmt.Errorf("failed to query admin users: %w", err)
	}
	defer adminRows.Close()

	adminUserIDs := []string{}
	for adminRows.Next() {
		var adminID string
		if err := adminRows.Scan(&adminID); err != nil {
			return fmt.Errorf("failed to scan admin user: %w", err)
		}
		adminUserIDs = append(adminUserIDs, adminID)
	}

	if err := adminRows.Err(); err != nil {
		return fmt.Errorf("error iterating admin users: %w", err)
	}

	// Prepare batch insert
	batchRows := [][]interface{}{}
	batchSize := 1000
	totalEntries := 0

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

	// For each user, create score_journal entries
	for _, userID := range s.Data.UserIDs {
		// Determine which projects this user participates in
		// We'll use the same logic as progress seeding
		userProjects := []string{}
		for _, projectID := range s.Data.ProjectIDs {
			if rand.Float64() < s.Config.ProjectParticipationRate {
				userProjects = append(userProjects, projectID)
			}
		}

		// Skip if user has no projects
		if len(userProjects) == 0 {
			continue
		}

		// Generate 3-10 entries for this user
		numEntries := 3 + rand.Intn(8)

		for i := 0; i < numEntries; i++ {
			// Random project from user's projects
			projectID := userProjects[rand.Intn(len(userProjects))]

			// 85% chance of ACHIEVEMENT, 15% chance of MANUAL (falls within 80-90% / 10-20%)
			isAchievement := rand.Float64() < 0.85

			// Generate entry ID and timestamp (spread over last 30 days)
			entryID := ulid.NewScoreJournalID()
			createdAt := time.Now().AddDate(0, 0, -rand.Intn(30))

			var row []interface{}

			if isAchievement && len(achievementsByProject[projectID]) > 0 {
				// ACHIEVEMENT type entry
				achievement := achievementsByProject[projectID][rand.Intn(len(achievementsByProject[projectID]))]

				row = []interface{}{
					entryID,
					projectID,
					userID,
					achievement.EventID,     // event_id (may be null)
					achievement.ChallengeID, // challenge_id (may be null)
					achievement.Points,      // points
					"ACHIEVEMENT",           // source_type
					achievement.ID,          // source_id
					nil,                     // reason (null for achievements)
					nil,                     // awarded_by (null for achievements)
					createdAt,               // created_at
				}
			} else {
				// MANUAL type entry
				points := 10 + rand.Intn(91) // 10-100 points
				reason := manualReasons[rand.Intn(len(manualReasons))]
				var awardedBy *string
				if len(adminUserIDs) > 0 {
					adminID := adminUserIDs[rand.Intn(len(adminUserIDs))]
					awardedBy = &adminID
				}

				// 40% chance to link to an event
				var eventID *string
				if rand.Float64() < 0.4 && len(s.Data.EventIDs[projectID]) > 0 {
					eID := s.Data.EventIDs[projectID][rand.Intn(len(s.Data.EventIDs[projectID]))]
					eventID = &eID
				}

				row = []interface{}{
					entryID,
					projectID,
					userID,
					eventID,       // event_id (may be null)
					nil,           // challenge_id (null for manual)
					int32(points), // points
					"MANUAL",      // source_type
					nil,           // source_id (null for manual)
					reason,        // reason
					awardedBy,     // awarded_by
					createdAt,     // created_at
				}
			}

			batchRows = append(batchRows, row)

			// Batch insert when we hit batch size
			if len(batchRows) >= batchSize {
				if err := s.insertScoreJournalBatch(batchRows); err != nil {
					return err
				}
				totalEntries += len(batchRows)
				batchRows = [][]interface{}{}
			}
		}
	}

	// Insert remaining rows
	if len(batchRows) > 0 {
		if err := s.insertScoreJournalBatch(batchRows); err != nil {
			return err
		}
		totalEntries += len(batchRows)
	}

	slog.Info("Score journal seeding completed", "total_entries", totalEntries)
	return nil
}

func (s *Seeder) insertScoreJournalBatch(rows [][]interface{}) error {
	_, err := s.DB.Pool.CopyFrom(
		s.Ctx,
		pgx.Identifier{"score_journal"},
		[]string{"id", "project_id", "user_id", "event_id", "challenge_id", "points", "source_type", "source_id", "reason", "awarded_by", "created_at"},
		pgx.CopyFromRows(rows),
	)
	return err
}
