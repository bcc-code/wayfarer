package e2e

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/bcc-media/wayfarer/e2e/testutil"
)

var (
	// testDB holds the shared PostgreSQL container
	testDB *testutil.TestDatabase
	// testDBMgr manages database operations
	testDBMgr *testutil.TestDBManager
)

// TestMain sets up the shared test environment
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start PostgreSQL container (shared across all tests)
	var err error
	testDB, err = testutil.NewTestDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to start test database: %v", err)
	}

	// Initialize database manager with migrations
	testDBMgr, err = testutil.NewTestDBManager(ctx, testDB.DSN)
	if err != nil {
		testDB.Close(ctx)
		log.Fatalf("Failed to initialize database manager: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testDBMgr.Close()
	testDB.Close(ctx)

	os.Exit(code)
}

// GetTestEnv returns the shared test environment
func GetTestEnv() (*testutil.TestDBManager, *testutil.TestDatabase) {
	return testDBMgr, testDB
}
