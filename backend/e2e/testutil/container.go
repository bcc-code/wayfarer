package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDatabase holds the container and connection info for a test PostgreSQL instance
type TestDatabase struct {
	Container *postgres.PostgresContainer
	DSN       string
	Host      string
	Port      string
}

// NewTestDatabase creates a new PostgreSQL container for testing
func NewTestDatabase(ctx context.Context) (*TestDatabase, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("wayfarer_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/wayfarer_test?sslmode=disable",
		host, port.Port())

	return &TestDatabase{
		Container: pgContainer,
		DSN:       dsn,
		Host:      host,
		Port:      port.Port(),
	}, nil
}

// Close terminates the PostgreSQL container
func (td *TestDatabase) Close(ctx context.Context) error {
	if td.Container != nil {
		return td.Container.Terminate(ctx)
	}
	return nil
}
