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
	"github.com/jaswdr/faker"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// TestDBManager manages database lifecycle for E2E tests
type TestDBManager struct {
	DSN     string
	DB      *database.DB
	sqlDB   *sql.DB
	seeder  *seeders.Seeder
	data    *seeders.SeededData
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
