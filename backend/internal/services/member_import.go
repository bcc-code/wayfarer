package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	memberImportMinAge      = 12
	memberImportMaxAge      = 37
	memberImportAPIPageSize = 500
	memberImportDefaultLang = "no"
)

// MemberImportService creates users for newly-eligible members ahead of their first login. Safe to call repeatedly -- findOrCreateUser matches existing rows by members_id/person_uuid, never duplicates.
type MemberImportService struct {
	DB            *database.DB
	MembersClient *members.Client
}

// ImportNewMembersResult reports what happened during one import run.
type ImportNewMembersResult struct {
	Fetched  int
	Imported int
	Skipped  int
	Errors   []string
}

type churchMaps struct {
	orgUidToOrgID   map[uuid.UUID]int
	orgIDToChurchID map[int]string
	excludedOrgIDs  map[int]bool
	defaultChurchID string
}

// ImportNewMembers creates a user for everyone in the target age band with an active affiliation not already present; a failed row is counted as an error and doesn't block the rest.
func (s *MemberImportService) ImportNewMembers(ctx context.Context) (*ImportNewMembersResult, error) {
	maps, err := s.buildChurchMaps(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build church maps: %w", err)
	}

	existingMembersIDs, err := s.getExistingMembersIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing members: %w", err)
	}

	result := &ImportNewMembersResult{}

	now := time.Now()
	maxBirthdate := now.AddDate(-memberImportMinAge, 0, 0).Format("2006-01-02")
	minBirthdate := now.AddDate(-memberImportMaxAge-1, 0, 1).Format("2006-01-02")

	offset := 0
	for {
		fetchedMembers, err := s.MembersClient.GetMembersByBirthdateRange(ctx, minBirthdate, maxBirthdate, memberImportAPIPageSize, offset)
		if err != nil {
			return result, fmt.Errorf("failed to fetch members at offset %d: %w", offset, err)
		}

		if len(fetchedMembers) == 0 {
			break
		}
		result.Fetched += len(fetchedMembers)

		for _, member := range fetchedMembers {
			if !members.HasActiveAffiliation(member.Affiliations) {
				result.Skipped++
				continue
			}

			membersID := strconv.Itoa(member.PersonID)
			if existingMembersIDs[membersID] {
				result.Skipped++
				continue
			}

			var personUUID pgtype.UUID
			if member.Uid != uuid.Nil {
				personUUID = pgtype.UUID{Bytes: member.Uid, Valid: true}
			}

			var birthdate pgtype.Date
			if parsed, ok := members.ParseBirthdate(member.BirthDate); ok {
				birthdate = pgtype.Date{Time: parsed, Valid: true}
			}

			_, err := s.DB.Queries.CreateUser(ctx, sqlc.CreateUserParams{
				ID:          ulid.NewUserID(),
				MembersID:   membersID,
				PersonUuid:  personUUID,
				Email:       member.Email,
				Name:        member.DisplayName,
				FirstName:   nilIfEmpty(member.FirstName),
				LastName:    nilIfEmpty(member.LastName),
				MiddleName:  nilIfEmpty(member.MiddleName),
				DisplayName: nilIfEmpty(member.DisplayName),
				Gender:      members.NormalizeGender(member.Gender),
				Birthdate:   birthdate,
				ChurchID:    resolveChurchID(member.Affiliations, maps),
				Language:    memberImportDefaultLang,
			})
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					// A concurrent run inserted this person first -- harmless, not an error.
					result.Skipped++
					existingMembersIDs[membersID] = true
					continue
				}
				slog.Error("member_import: failed to create user", "members_id", membersID, "error", err)
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create user for members_id %s: %v", membersID, err))
				continue
			}

			existingMembersIDs[membersID] = true
			result.Imported++
		}

		if len(fetchedMembers) < memberImportAPIPageSize {
			break
		}
		offset += memberImportAPIPageSize
	}

	slog.Info("member_import: import completed",
		"fetched", result.Fetched,
		"imported", result.Imported,
		"skipped", result.Skipped,
		"errors", len(result.Errors),
	)

	return result, nil
}

func (s *MemberImportService) buildChurchMaps(ctx context.Context) (*churchMaps, error) {
	orgs, err := s.MembersClient.GetAllOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organizations: %w", err)
	}

	orgUidToOrgID := make(map[uuid.UUID]int, len(orgs))
	excludedOrgIDs := make(map[int]bool)
	for _, org := range orgs {
		orgUidToOrgID[org.Uid] = org.OrgID
		if members.ExcludedOrgNames[org.Name] {
			excludedOrgIDs[org.OrgID] = true
		}
	}

	churches, err := s.DB.Queries.GetChurchesFilteredCursor(ctx, sqlc.GetChurchesFilteredCursorParams{
		Querylimit: 10000,
		Isbackward: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch churches: %w", err)
	}

	orgIDToChurchID := make(map[int]string, len(churches))
	for _, church := range churches {
		if church.ExternalID != nil {
			orgIDToChurchID[int(*church.ExternalID)] = church.ID
		}
	}

	defaultChurch, err := s.DB.Queries.GetDefaultChurch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get default church: %w", err)
	}

	return &churchMaps{
		orgUidToOrgID:   orgUidToOrgID,
		orgIDToChurchID: orgIDToChurchID,
		excludedOrgIDs:  excludedOrgIDs,
		defaultChurchID: defaultChurch.ID,
	}, nil
}

func (s *MemberImportService) getExistingMembersIDs(ctx context.Context) (map[string]bool, error) {
	rows, err := s.DB.Pool.Query(ctx, "SELECT members_id FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var membersID string
		if err := rows.Scan(&membersID); err != nil {
			return nil, err
		}
		result[membersID] = true
	}
	return result, rows.Err()
}

// resolveChurchID picks the first non-excluded, known org from active affiliations, or the default church.
func resolveChurchID(affiliations []members.Affiliation, maps *churchMaps) string {
	for _, aff := range members.FilterActiveAffiliations(affiliations) {
		orgID, ok := maps.orgUidToOrgID[aff.OrgUid]
		if !ok {
			continue
		}
		if maps.excludedOrgIDs[orgID] {
			continue
		}
		if churchID, ok := maps.orgIDToChurchID[orgID]; ok {
			return churchID
		}
	}
	return maps.defaultChurchID
}

// nilIfEmpty maps an empty string to nil so an absent field stores as SQL NULL, not "".
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
