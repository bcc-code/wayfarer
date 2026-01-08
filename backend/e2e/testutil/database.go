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
	// Seed both the faker and math/rand with the same seed for determinism
	rand.Seed(seed)
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
	roleID := "UR" + generateULID()

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
	roleID := "UR" + generateULID()

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

// generateULID generates a simple ULID-like ID for tests
func generateULID() string {
	// Simple timestamp + random for test purposes
	const chars = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	b := make([]byte, 26)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
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
	scoreID := "SJ" + generateULID()
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
	consentID = "CN" + generateULID()
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

	historyID := "UH" + generateULID()
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
