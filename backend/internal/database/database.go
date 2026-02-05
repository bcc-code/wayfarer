package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx stdlib driver for database/sql
)

//go:embed migrations/*.sql
var Migrations embed.FS

// Migrate runs all pending database migrations
func Migrate(ctx context.Context, databaseURL string) error {
	slog.Info("Running database migrations")

	// Open a database/sql connection for goose (it requires database/sql, not pgxpool)
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database for migrations: %w", err)
	}

	// Configure goose to use embedded migrations
	goose.SetBaseFS(Migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Run migrations
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	slog.Info("Database migrations completed successfully")
	return nil
}

// DB wraps the pgxpool connection pool and sqlc queries
type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

// queryNameRegex extracts the query name from sqlc-generated SQL comments
// Format: -- name: QueryName :one
var queryNameRegex = regexp.MustCompile(`(?m)^--\s*name:\s*(\w+)`)

// extractQueryName attempts to extract the query name from the SQL statement
// If a "-- name: QueryName" comment is found, returns "QueryName"
// Otherwise returns "query" as a fallback
func extractQueryName(sql string) string {
	matches := queryNameRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: try to get first word of SQL (SELECT, INSERT, etc.)
	words := strings.Fields(strings.TrimSpace(sql))
	if len(words) > 0 {
		return strings.ToLower(words[0])
	}

	return "query"
}

// Connect creates a new database connection pool
func Connect(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Apply connection pool settings from config
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	// Add OpenTelemetry tracing for database queries
	// This automatically enables if a global tracer is set via otel.SetTracerProvider()
	// Configure with custom span name extraction from sqlc comments
	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithSpanNameFunc(func(stmt string) string {
			// Extract query name from "-- name: QueryName" comment
			queryName := extractQueryName(stmt)
			return "db." + queryName
		}),
	)
	slog.Info("OpenTelemetry database tracing configured with query name extraction")

	// Enable console query logging if configured (for debugging)
	if cfg.LogQueries {
		poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger: tracelog.LoggerFunc(func(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
				// Map tracelog levels to slog levels
				var slogLevel slog.Level
				switch level {
				case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
					slogLevel = slog.LevelDebug
				case tracelog.LogLevelInfo:
					slogLevel = slog.LevelInfo
				case tracelog.LogLevelWarn:
					slogLevel = slog.LevelWarn
				case tracelog.LogLevelError:
					slogLevel = slog.LevelError
				default:
					slogLevel = slog.LevelInfo
				}

				// Extract query information
				attrs := []any{"component", "database"}
				for k, v := range data {
					attrs = append(attrs, k, v)
				}

				slog.Log(ctx, slogLevel, msg, attrs...)
			}),
			LogLevel: tracelog.LogLevelTrace,
		}
		slog.Info("database query logging enabled (replaces OTEL tracing)")
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	slog.Info("database connection established",
		"max_conns", cfg.MaxOpenConns,
		"min_conns", cfg.MaxIdleConns,
	)

	return &DB{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		slog.Info("database connection closed")
	}
}

// Ping verifies the database connection is still alive
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
