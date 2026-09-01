package services

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemberImportServiceCreation(t *testing.T) {
	service := &MemberImportService{
		DB:            nil,
		MembersClient: nil,
	}
	require.NotNil(t, service)
}

func TestImportNewMembersResultFields(t *testing.T) {
	result := &ImportNewMembersResult{
		Fetched:  10,
		Imported: 7,
		Skipped:  3,
		Errors:   []string{"boom"},
	}

	assert.Equal(t, 10, result.Fetched)
	assert.Equal(t, 7, result.Imported)
	assert.Equal(t, 3, result.Skipped)
	assert.Equal(t, []string{"boom"}, result.Errors)
}

func TestNilIfEmpty(t *testing.T) {
	t.Run("empty string becomes nil", func(t *testing.T) {
		assert.Nil(t, nilIfEmpty(""))
	})

	t.Run("non-empty string becomes a pointer to itself", func(t *testing.T) {
		got := nilIfEmpty("Jan")
		require.NotNil(t, got)
		assert.Equal(t, "Jan", *got)
	})
}

func TestResolveChurchID(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()
	orgExcluded := uuid.New()

	maps := &churchMaps{
		orgUidToOrgID: map[uuid.UUID]int{
			orgA:        1,
			orgB:        2,
			orgExcluded: 3,
		},
		orgIDToChurchID: map[int]string{
			1: "CH01AAAAAAAAAAAAAAAAAAAAAA",
			2: "CH02BBBBBBBBBBBBBBBBBBBBBB",
		},
		excludedOrgIDs: map[int]bool{
			3: true,
		},
		defaultChurchID: "CH00DEFAULTDEFAULTDEFAULT0",
	}

	t.Run("resolves the first non-excluded known affiliation", func(t *testing.T) {
		affiliations := []members.Affiliation{
			{OrgUid: orgA, Active: true},
		}
		assert.Equal(t, "CH01AAAAAAAAAAAAAAAAAAAAAA", resolveChurchID(affiliations, maps))
	})

	t.Run("skips excluded orgs and falls through to the next affiliation", func(t *testing.T) {
		affiliations := []members.Affiliation{
			{OrgUid: orgExcluded, Active: true},
			{OrgUid: orgB, Active: true},
		}
		assert.Equal(t, "CH02BBBBBBBBBBBBBBBBBBBBBB", resolveChurchID(affiliations, maps))
	})

	t.Run("skips inactive affiliations", func(t *testing.T) {
		affiliations := []members.Affiliation{
			{OrgUid: orgA, Active: false},
		}
		assert.Equal(t, "CH00DEFAULTDEFAULTDEFAULT0", resolveChurchID(affiliations, maps))
	})

	t.Run("falls back to the default church when nothing matches", func(t *testing.T) {
		unknownOrg := uuid.New()
		affiliations := []members.Affiliation{
			{OrgUid: unknownOrg, Active: true},
		}
		assert.Equal(t, "CH00DEFAULTDEFAULTDEFAULT0", resolveChurchID(affiliations, maps))
	})

	t.Run("no affiliations at all falls back to the default church", func(t *testing.T) {
		assert.Equal(t, "CH00DEFAULTDEFAULTDEFAULT0", resolveChurchID(nil, maps))
	})
}
