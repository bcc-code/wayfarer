package database

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

//go:embed migrations/*.sql
var Migrations embed.FS

// DB wraps the pgxpool connection pool and sqlc queries
type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
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

	// Enable query logging if configured
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
		slog.Info("database query logging enabled")
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
