package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/bcc-media/wayfarer/cmd/seed/seeders"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jaswdr/faker"
	"github.com/pressly/goose/v3"
)

// TestDBManager manages database lifecycle for E2E tests
type TestDBManager struct {
	DSN    string
	DB     *database.DB
	sqlDB  *sql.DB
	seeder *seeders.Seeder
	data   *seeders.SeededData
}

// NewTestDBManager creates a new database manager and runs migrations
func NewTestDBManager(ctx context.Context, dsn string) (*TestDBManager, error) {
	// Standard sql.DB for goose migrations
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations using goose
	goose.SetBaseFS(database.Migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Connect with pgx pool for application use
	db, err := database.Connect(ctx, config.DatabaseConfig{
		URL:             dsn,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return &TestDBManager{
		DSN:   dsn,
		DB:    db,
		sqlDB: sqlDB,
	}, nil
}

// DefaultSeedConfig returns a minimal seed configuration for tests
func DefaultSeedConfig() seeders.SeedConfig {
	return seeders.SeedConfig{
		NumUsers:                  20,
		NumProjects:               2,
		NumChurches:               3,
		NumSuperTeams:             2,
		NumAchievements:           10,
		TeamSize:                  5,
		ProjectParticipationRate:  0.8,
		AchievementCompletionRate: 0.5,
	}
}

// Seed populates the database with deterministic test data
func (m *TestDBManager) Seed(ctx context.Context, seed int64, cfg seeders.SeedConfig) (*seeders.SeededData, error) {
	fake := faker.NewWithSeed(rand.NewSource(seed))

	m.data = seeders.NewSeededData()
	m.seeder = &seeders.Seeder{
		DB:     m.DB,
		Fake:   fake,
		Ctx:    ctx,
		Data:   m.data,
		Config: cfg,
	}

	stats := &seeders.Stats{}

	// Seed in order (maintaining FK relationships)
	seedFns := []func(*seeders.Stats) error{
		m.seeder.SeedChurches,
		m.seeder.SeedUsers,
		m.seeder.SeedProjects,
		m.seeder.SeedTeams,
		m.seeder.SeedChallenges,
		m.seeder.SeedAchievements,
		m.seeder.SeedProgress,
		m.seeder.SeedScoreJournal,
	}

	for _, fn := range seedFns {
		if err := fn(stats); err != nil {
			return nil, fmt.Errorf("seeding failed: %w", err)
		}
	}

	// Update settings to point to a valid project from the seeded data
	if len(m.data.ProjectIDs) > 0 {
		_, err := m.DB.Pool.Exec(ctx, `
			UPDATE settings SET value_text = $1 WHERE key = 'current_project_id'
		`, m.data.ProjectIDs[0])
		if err != nil {
			return nil, fmt.Errorf("failed to update current_project_id setting: %w", err)
		}
	}

	return m.data, nil
}

// Clean truncates all tables for test isolation
func (m *TestDBManager) Clean(ctx context.Context) error {
	// Get all user tables from the database and truncate them
	// This is more robust than maintaining a hardcoded list
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
		AND table_name NOT IN ('goose_db_version', 'settings')
		ORDER BY table_name
	`
	rows, err := m.DB.Pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating tables: %w", err)
	}

	// Truncate all tables at once with CASCADE
	if len(tables) > 0 {
		truncateSQL := "TRUNCATE TABLE "
		for i, table := range tables {
			if i > 0 {
				truncateSQL += ", "
			}
			truncateSQL += table
		}
		truncateSQL += " CASCADE"

		_, err = m.DB.Pool.Exec(ctx, truncateSQL)
		if err != nil {
			return fmt.Errorf("failed to truncate tables: %w", err)
		}
	}

	// Clear seeded data references
	m.data = nil
	m.seeder = nil

	return nil
}

// Data returns the seeded data from the last Seed call
func (m *TestDBManager) Data() *seeders.SeededData {
	return m.data
}

// Close closes all database connections
func (m *TestDBManager) Close() error {
	if m.DB != nil {
		m.DB.Close()
	}
	if m.sqlDB != nil {
		return m.sqlDB.Close()
	}
	return nil
}

// RoleType matches the database enum values
type RoleType string

const (
	RoleSuperAdmin   RoleType = "SUPERADMIN"
	RoleAdmin        RoleType = "ADMIN"
	RoleChurchAdmin  RoleType = "CHURCH_ADMIN"
	RoleProjectAdmin RoleType = "PROJECT_ADMIN"
	RoleTeamLead     RoleType = "TEAM_LEAD"
	RoleUser         RoleType = "USER"
	RoleM2M          RoleType = "M2M"
)

// AssignRole assigns a role to a user in the database
func (m *TestDBManager) AssignRole(ctx context.Context, userID string, role RoleType) error {
	// Generate a ULID for the role
	roleID := ulid.NewUserRoleID()

	query := `
		INSERT INTO user_roles (id, user_id, role, assigned_by, assigned_at)
		VALUES ($1, $2, $3, $4, now())
	`
	_, err := m.DB.Pool.Exec(ctx, query, roleID, userID, string(role), userID)
	if err != nil {
		return fmt.Errorf("failed to assign role %s to user %s: %w", role, userID, err)
	}
	return nil
}

// AssignRoleWithScope assigns a scoped role (church/project/team) to a user
func (m *TestDBManager) AssignRoleWithScope(ctx context.Context, userID string, role RoleType, churchID, projectID, teamID *string) error {
	roleID := ulid.NewUserRoleID()

	query := `
		INSERT INTO user_roles (id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
	`
	_, err := m.DB.Pool.Exec(ctx, query, roleID, userID, string(role), churchID, projectID, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to assign scoped role %s to user %s: %w", role, userID, err)
	}
	return nil
}

// GetUserLanguage returns the language stored for a user in the database
func (m *TestDBManager) GetUserLanguage(ctx context.Context, userID string) (string, error) {
	var language string
	err := m.DB.Pool.QueryRow(ctx, "SELECT language FROM users WHERE id = $1", userID).Scan(&language)
	if err != nil {
		return "", fmt.Errorf("failed to get user language: %w", err)
	}
	return language, nil
}

// SetUserLanguage sets the language for a user in the database
func (m *TestDBManager) SetUserLanguage(ctx context.Context, userID, language string) error {
	_, err := m.DB.Pool.Exec(ctx, "UPDATE users SET language = $1 WHERE id = $2", language, userID)
	if err != nil {
		return fmt.Errorf("failed to set user language: %w", err)
	}
	return nil
}

// CreateTestUser creates a user with specific birthdate and gender for testing
func (m *TestDBManager) CreateTestUser(ctx context.Context, id, name, gender string, birthdate time.Time, churchID string) error {
	membersID := "TEST-" + id
	email := id + "@test.example.com"

	query := `
		INSERT INTO users (id, members_id, email, name, gender, birthdate, church_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, membersID, email, name, gender, birthdate, churchID)
	if err != nil {
		return fmt.Errorf("failed to create test user %s: %w", id, err)
	}
	return nil
}

// CreateTestChurch creates a test church
func (m *TestDBManager) CreateTestChurch(ctx context.Context, id, name, country, category string) error {
	query := `
		INSERT INTO churches (id, name, country, category)
		VALUES ($1, $2, $3, $4)
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, name, country, category)
	if err != nil {
		return fmt.Errorf("failed to create test church %s: %w", id, err)
	}
	return nil
}

// CreateTestProject creates a test project with minimal required fields
func (m *TestDBManager) CreateTestProject(ctx context.Context, id, name string) error {
	query := `
		INSERT INTO projects (
			id, name, description, start_date, end_date,
			color_light_accent, color_light_accent_contrast, color_light_on_accent,
			color_light_background_default, color_light_background_raised, color_light_background_indent,
			color_light_text_default, color_light_text_muted, color_light_text_hint,
			color_light_shadow_default, color_light_shadow_blank, color_light_border_default,
			color_dark_accent, color_dark_accent_contrast, color_dark_on_accent,
			color_dark_background_default, color_dark_background_raised, color_dark_background_indent,
			color_dark_text_default, color_dark_text_muted, color_dark_text_hint,
			color_dark_shadow_default, color_dark_shadow_blank, color_dark_border_default,
			rounding
		)
		VALUES (
			$1, $2, 'Test project', now(), now() + interval '1 year',
			'#000', '#000', '#000', '#000', '#000', '#000',
			'#000', '#000', '#000', '#000', '#000', '#000',
			'#000', '#000', '#000', '#000', '#000', '#000',
			'#000', '#000', '#000', '#000', '#000', '#000',
			0
		)
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, name)
	if err != nil {
		return fmt.Errorf("failed to create test project %s: %w", id, err)
	}
	return nil
}

// CreateTestTeam creates a test team
func (m *TestDBManager) CreateTestTeam(ctx context.Context, id, name, projectID string) error {
	joinCode := "TEST-" + id
	query := `
		INSERT INTO teams (id, project_id, name, join_code)
		VALUES ($1, $2, $3, $4)
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, projectID, name, joinCode)
	if err != nil {
		return fmt.Errorf("failed to create test team %s: %w", id, err)
	}
	return nil
}

// EnrollUserInProject adds a user to a project
func (m *TestDBManager) EnrollUserInProject(ctx context.Context, userID, projectID string) error {
	query := `
		INSERT INTO user_projects (user_id, project_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := m.DB.Pool.Exec(ctx, query, userID, projectID)
	if err != nil {
		return fmt.Errorf("failed to enroll user %s in project %s: %w", userID, projectID, err)
	}
	return nil
}

// AddUserToTeam adds a user as a team member
func (m *TestDBManager) AddUserToTeam(ctx context.Context, userID, teamID string) error {
	query := `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := m.DB.Pool.Exec(ctx, query, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to add user %s to team %s: %w", userID, teamID, err)
	}
	return nil
}

// AddScoreForUser adds a score journal entry for a user in a project
func (m *TestDBManager) AddScoreForUser(ctx context.Context, userID, projectID string, points int) error {
	scoreID := ulid.NewScoreJournalID()
	query := `
		INSERT INTO score_journal (id, project_id, user_id, points, source_type, reason)
		VALUES ($1, $2, $3, $4, 'MANUAL', 'Test score')
	`
	_, err := m.DB.Pool.Exec(ctx, query, scoreID, projectID, userID, points)
	if err != nil {
		return fmt.Errorf("failed to add score for user %s: %w", userID, err)
	}
	return nil
}

// EnsureLeaderboardConsent ensures the leaderboard_consent consent type exists
func (m *TestDBManager) EnsureLeaderboardConsent(ctx context.Context) (string, error) {
	// Check if consent exists
	var consentID string
	err := m.DB.Pool.QueryRow(ctx, `
		SELECT id FROM consents WHERE key = 'leaderboard_consent' LIMIT 1
	`).Scan(&consentID)

	if err == nil {
		return consentID, nil
	}

	// Create consent if it doesn't exist
	consentID = ulid.NewConsentID()
	query := `
		INSERT INTO consents (id, key, version, title, body, published_at)
		VALUES ($1, 'leaderboard_consent', 1, 'Leaderboard Consent', 'Test consent for leaderboard visibility', now())
	`
	_, err = m.DB.Pool.Exec(ctx, query, consentID)
	if err != nil {
		return "", fmt.Errorf("failed to create leaderboard consent: %w", err)
	}
	return consentID, nil
}

// AddLeaderboardConsent adds leaderboard consent for a user
func (m *TestDBManager) AddLeaderboardConsent(ctx context.Context, userID string) error {
	consentID, err := m.EnsureLeaderboardConsent(ctx)
	if err != nil {
		return err
	}

	historyID := ulid.NewUserConsentHistoryID()
	query := `
		INSERT INTO user_consent_history (id, user_id, consent_id, consent_key, action, occurred_at)
		VALUES ($1, $2, $3, 'leaderboard_consent', 'ACCEPTED', now())
	`
	_, err = m.DB.Pool.Exec(ctx, query, historyID, userID, consentID)
	if err != nil {
		return fmt.Errorf("failed to add leaderboard consent for user %s: %w", userID, err)
	}
	return nil
}

// CreateExternalContent creates an external content record for testing
func (m *TestDBManager) CreateExternalContent(ctx context.Context, id, planID, taskID, contentType, source string) error {
	query := `
		INSERT INTO external_content (id, plan_id, task_id, content_type, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, now(), now())
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, planID, taskID, contentType, source)
	if err != nil {
		return fmt.Errorf("failed to create external content %s: %w", id, err)
	}
	return nil
}

// CreateExternalContentEvent creates a content event (simulating an event before user exists)
func (m *TestDBManager) CreateExternalContentEvent(ctx context.Context, id string, personUUID uuid.UUID, taskID string, planID *string, source string, progress *float32) error {
	query := `
		INSERT INTO external_content_events (id, person_id, task_id, plan_id, source, received_at, content_progress, consumed_at)
		VALUES ($1, $2, $3, $4, $5, now(), $6, now())
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, personUUID, taskID, planID, source, progress)
	if err != nil {
		return fmt.Errorf("failed to create external content event %s: %w", id, err)
	}
	return nil
}

// CreateContentAchievement creates a content achievement with linked content items
func (m *TestDBManager) CreateContentAchievement(ctx context.Context, id, projectID, name string, points int, contentIDs []string, hidden bool) error {
	// Create the achievement in the main achievements table
	achievementQuery := `
		INSERT INTO achievements (
			id, project_id, name, achievement_type, points, hidden,
			description_pending, description_completed, notification_text,
			image_pending, image_completed, sort_order
		)
		VALUES ($1, $2, $3, 'CONTENT', $4, $5,
			'Complete the content', 'Content completed!', 'Achievement unlocked!',
			'https://example.com/pending.png', 'https://example.com/completed.png', 0)
	`
	_, err := m.DB.Pool.Exec(ctx, achievementQuery, id, projectID, name, points, hidden)
	if err != nil {
		return fmt.Errorf("failed to create content achievement %s: %w", id, err)
	}

	// Create entry in content_achievements junction table
	contentAchievementQuery := `
		INSERT INTO content_achievements (achievement_id)
		VALUES ($1)
	`
	_, err = m.DB.Pool.Exec(ctx, contentAchievementQuery, id)
	if err != nil {
		return fmt.Errorf("failed to create content_achievements entry for %s: %w", id, err)
	}

	// Link content items to the achievement
	for i, contentID := range contentIDs {
		itemID := ulid.NewContentItemID()
		itemQuery := `
			INSERT INTO content_achievement_items (id, achievement_id, external_content_id, sort_order)
			VALUES ($1, $2, $3, $4)
		`
		_, err := m.DB.Pool.Exec(ctx, itemQuery, itemID, id, contentID, i)
		if err != nil {
			return fmt.Errorf("failed to link content item %s to achievement %s: %w", contentID, id, err)
		}
	}

	return nil
}

// CreateUserWithPersonUUID creates a user with a specific person_uuid for testing
func (m *TestDBManager) CreateUserWithPersonUUID(ctx context.Context, id, name, churchID string, personUUID uuid.UUID) error {
	membersID := "TEST-" + id
	email := id + "@test.example.com"

	query := `
		INSERT INTO users (id, members_id, person_uuid, email, name, gender, birthdate, church_id)
		VALUES ($1, $2, $3, $4, $5, 'UNKNOWN', '2000-01-01', $6)
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, membersID, personUUID, email, name, churchID)
	if err != nil {
		return fmt.Errorf("failed to create test user %s with person_uuid: %w", id, err)
	}
	return nil
}

// GetUserAchievements returns the achievement IDs awarded to a user
func (m *TestDBManager) GetUserAchievements(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT achievement_id FROM user_achievements WHERE user_id = $1`
	rows, err := m.DB.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user achievements: %w", err)
	}
	defer rows.Close()

	var achievementIDs []string
	for rows.Next() {
		var achievementID string
		if err := rows.Scan(&achievementID); err != nil {
			return nil, fmt.Errorf("failed to scan achievement ID: %w", err)
		}
		achievementIDs = append(achievementIDs, achievementID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating achievements: %w", err)
	}

	return achievementIDs, nil
}

// GetScoreJournalEntriesCount returns the count of score journal entries for a user in a project
func (m *TestDBManager) GetScoreJournalEntriesCount(ctx context.Context, userID, projectID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM score_journal WHERE user_id = $1 AND project_id = $2`
	err := m.DB.Pool.QueryRow(ctx, query, userID, projectID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count score journal entries: %w", err)
	}
	return count, nil
}

// GetUserContentProgress returns the count of completed content items for a user and achievement
func (m *TestDBManager) GetUserContentProgress(ctx context.Context, userID, achievementID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM user_content_progress WHERE user_id = $1 AND achievement_id = $2`
	err := m.DB.Pool.QueryRow(ctx, query, userID, achievementID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user content progress: %w", err)
	}
	return count, nil
}

// CreateTestConsent creates a test consent with the given key
func (m *TestDBManager) CreateTestConsent(ctx context.Context, key string, isRemote bool) (string, error) {
	consentID := ulid.NewConsentID()
	query := `
		INSERT INTO consents (id, key, version, title, short_text, body, published_at, is_remote)
		VALUES ($1, $2, 1, $3, 'Short text for consent', 'Body of the consent document', now(), $4)
	`
	title := "Test Consent: " + key
	_, err := m.DB.Pool.Exec(ctx, query, consentID, key, title, isRemote)
	if err != nil {
		return "", fmt.Errorf("failed to create test consent %s: %w", key, err)
	}
	return consentID, nil
}

// GetUserConsentHistoryCount returns the count of consent history entries for a user and consent key
func (m *TestDBManager) GetUserConsentHistoryCount(ctx context.Context, userID, consentKey string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM user_consent_history WHERE user_id = $1 AND consent_key = $2`
	err := m.DB.Pool.QueryRow(ctx, query, userID, consentKey).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user consent history: %w", err)
	}
	return count, nil
}

// GetLatestUserConsentAction returns the latest action and source for a user's consent
func (m *TestDBManager) GetLatestUserConsentAction(ctx context.Context, userID, consentKey string) (action string, source *string, err error) {
	query := `
		SELECT action, source FROM user_consent_history
		WHERE user_id = $1 AND consent_key = $2
		ORDER BY occurred_at DESC
		LIMIT 1
	`
	err = m.DB.Pool.QueryRow(ctx, query, userID, consentKey).Scan(&action, &source)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get latest consent action: %w", err)
	}
	return action, source, nil
}

// RemoveUserFromTeam removes a user from a team
func (m *TestDBManager) RemoveUserFromTeam(ctx context.Context, userID, teamID string) error {
	query := `DELETE FROM team_members WHERE user_id = $1 AND team_id = $2`
	_, err := m.DB.Pool.Exec(ctx, query, userID, teamID)
	if err != nil {
		return fmt.Errorf("failed to remove user %s from team %s: %w", userID, teamID, err)
	}
	return nil
}

// CreateTestEvent creates a test event for a project
func (m *TestDBManager) CreateTestEvent(ctx context.Context, id, name, projectID string) error {
	query := `
		INSERT INTO events (id, project_id, name, description, start_date, end_date)
		VALUES ($1, $2, $3, 'Test event description', now(), now() + interval '1 month')
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, projectID, name)
	if err != nil {
		return fmt.Errorf("failed to create test event %s: %w", id, err)
	}
	return nil
}

// EnrollUserInEvent adds a user to an event
func (m *TestDBManager) EnrollUserInEvent(ctx context.Context, userID, eventID string) error {
	query := `
		INSERT INTO user_events (user_id, event_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := m.DB.Pool.Exec(ctx, query, userID, eventID)
	if err != nil {
		return fmt.Errorf("failed to enroll user %s in event %s: %w", userID, eventID, err)
	}
	return nil
}

// RemoveUserFromEvent removes a user from an event
func (m *TestDBManager) RemoveUserFromEvent(ctx context.Context, userID, eventID string) error {
	query := `DELETE FROM user_events WHERE user_id = $1 AND event_id = $2`
	_, err := m.DB.Pool.Exec(ctx, query, userID, eventID)
	if err != nil {
		return fmt.Errorf("failed to remove user %s from event %s: %w", userID, eventID, err)
	}
	return nil
}

// AddScoreForUserEvent adds a score journal entry for a user in an event
func (m *TestDBManager) AddScoreForUserEvent(ctx context.Context, userID, projectID, eventID string, points int) error {
	scoreID := ulid.NewScoreJournalID()
	query := `
		INSERT INTO score_journal (id, project_id, event_id, user_id, points, source_type, reason)
		VALUES ($1, $2, $3, $4, $5, 'MANUAL', 'Test event score')
	`
	_, err := m.DB.Pool.Exec(ctx, query, scoreID, projectID, eventID, userID, points)
	if err != nil {
		return fmt.Errorf("failed to add event score for user %s: %w", userID, err)
	}
	return nil
}

// RemoveUserFromProject removes a user from a project
func (m *TestDBManager) RemoveUserFromProject(ctx context.Context, userID, projectID string) error {
	query := `DELETE FROM user_projects WHERE user_id = $1 AND project_id = $2`
	_, err := m.DB.Pool.Exec(ctx, query, userID, projectID)
	if err != nil {
		return fmt.Errorf("failed to remove user %s from project %s: %w", userID, projectID, err)
	}
	return nil
}

// CreateTestSuperTeam creates a test superteam
func (m *TestDBManager) CreateTestSuperTeam(ctx context.Context, id, name, projectID string) error {
	query := `
		INSERT INTO super_teams (id, project_id, name)
		VALUES ($1, $2, $3)
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, projectID, name)
	if err != nil {
		return fmt.Errorf("failed to create test superteam %s: %w", id, err)
	}
	return nil
}

// CreateTestTeamWithSuperTeam creates a team with a superteam association
func (m *TestDBManager) CreateTestTeamWithSuperTeam(ctx context.Context, id, name, projectID, superTeamID string) error {
	joinCode := "TEST-" + id
	query := `
		INSERT INTO teams (id, project_id, name, join_code, super_team_id)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := m.DB.Pool.Exec(ctx, query, id, projectID, name, joinCode, superTeamID)
	if err != nil {
		return fmt.Errorf("failed to create test team %s with superteam: %w", id, err)
	}
	return nil
}

// GetLeaderboardProjectTeamPoints returns total_points for a team in project leaderboard
func (m *TestDBManager) GetLeaderboardProjectTeamPoints(ctx context.Context, projectID, teamID string) (int64, error) {
	var points int64
	query := `SELECT COALESCE(total_points, 0) FROM leaderboard_project_teams WHERE project_id = $1 AND team_id = $2`
	err := m.DB.Pool.QueryRow(ctx, query, projectID, teamID).Scan(&points)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get leaderboard project team points: %w", err)
	}
	return points, nil
}

// GetLeaderboardProjectSuperTeamPoints returns total_points for a superteam in project leaderboard
func (m *TestDBManager) GetLeaderboardProjectSuperTeamPoints(ctx context.Context, projectID, superTeamID string) (int64, error) {
	var points int64
	query := `SELECT COALESCE(total_points, 0) FROM leaderboard_project_superteams WHERE project_id = $1 AND super_team_id = $2`
	err := m.DB.Pool.QueryRow(ctx, query, projectID, superTeamID).Scan(&points)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get leaderboard project superteam points: %w", err)
	}
	return points, nil
}

// GetLeaderboardProjectChurchPoints returns total_points for a church in project leaderboard
func (m *TestDBManager) GetLeaderboardProjectChurchPoints(ctx context.Context, projectID, churchID string) (int64, error) {
	var points int64
	query := `SELECT COALESCE(total_points, 0) FROM leaderboard_project_churches WHERE project_id = $1 AND church_id = $2`
	err := m.DB.Pool.QueryRow(ctx, query, projectID, churchID).Scan(&points)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get leaderboard project church points: %w", err)
	}
	return points, nil
}

// GetLeaderboardEventTeamPoints returns total_points for a team in event leaderboard
func (m *TestDBManager) GetLeaderboardEventTeamPoints(ctx context.Context, eventID, teamID string) (int64, error) {
	var points int64
	query := `SELECT COALESCE(total_points, 0) FROM leaderboard_event_teams WHERE event_id = $1 AND team_id = $2`
	err := m.DB.Pool.QueryRow(ctx, query, eventID, teamID).Scan(&points)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get leaderboard event team points: %w", err)
	}
	return points, nil
}

// GetLeaderboardEventSuperTeamPoints returns total_points for a superteam in event leaderboard
func (m *TestDBManager) GetLeaderboardEventSuperTeamPoints(ctx context.Context, eventID, superTeamID string) (int64, error) {
	var points int64
	query := `SELECT COALESCE(total_points, 0) FROM leaderboard_event_superteams WHERE event_id = $1 AND super_team_id = $2`
	err := m.DB.Pool.QueryRow(ctx, query, eventID, superTeamID).Scan(&points)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get leaderboard event superteam points: %w", err)
	}
	return points, nil
}

// GetLeaderboardEventChurchPoints returns total_points for a church in event leaderboard
func (m *TestDBManager) GetLeaderboardEventChurchPoints(ctx context.Context, eventID, churchID string) (int64, error) {
	var points int64
	query := `SELECT COALESCE(total_points, 0) FROM leaderboard_event_churches WHERE event_id = $1 AND church_id = $2`
	err := m.DB.Pool.QueryRow(ctx, query, eventID, churchID).Scan(&points)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get leaderboard event church points: %w", err)
	}
	return points, nil
}
