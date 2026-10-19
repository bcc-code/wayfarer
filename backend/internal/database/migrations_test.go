package database

import (
	"io/fs"
	"regexp"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var migrationFileRe = regexp.MustCompile(`^(\d{5})_[a-z0-9_]+\.sql$`)

// migrationProblems returns malformed filenames and versions shared by more
// than one file (goose refuses to run with duplicate versions).
func migrationProblems(names []string) (malformed []string, duplicates map[string][]string) {
	byVersion := map[string][]string{}
	for _, name := range names {
		m := migrationFileRe.FindStringSubmatch(name)
		if m == nil {
			malformed = append(malformed, name)
			continue
		}
		byVersion[m[1]] = append(byVersion[m[1]], name)
	}

	duplicates = map[string][]string{}
	for version, files := range byVersion {
		if len(files) > 1 {
			sort.Strings(files)
			duplicates[version] = files
		}
	}
	return malformed, duplicates
}

func TestMigrationProblems(t *testing.T) {
	malformed, duplicates := migrationProblems([]string{
		"00001_initial.sql",
		"00002_a.sql",
		"00002_b.sql",
		"3_bad.sql",
		"00004_Upper.sql",
	})

	assert.Equal(t, []string{"3_bad.sql", "00004_Upper.sql"}, malformed)
	assert.Equal(t, map[string][]string{"00002": {"00002_a.sql", "00002_b.sql"}}, duplicates)

	malformed, duplicates = migrationProblems([]string{"00001_a.sql", "00002_b.sql"})
	assert.Empty(t, malformed)
	assert.Empty(t, duplicates)
}

func TestMigrationVersionsUnique(t *testing.T) {
	names, err := fs.Glob(Migrations, "migrations/*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, names)

	for i, name := range names {
		names[i] = name[len("migrations/"):]
	}

	malformed, duplicates := migrationProblems(names)
	for _, name := range malformed {
		t.Errorf("migration %q does not match NNNNN_snake_case.sql", name)
	}
	for version, files := range duplicates {
		t.Errorf("migration version %s is used by multiple files: %v — renumber all but one", version, files)
	}
}
