package database

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
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
