package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/bcc-media/wayfarer/internal/auth0"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sony/gobreaker/v2"
)

const (
	minAge       = 12
	maxAge       = 37
	apiPageSize  = 500
	dbBatchSize  = 100
	defaultLang  = "no"
)

// ExcludedChurchNames contains organization names that should not be used for church assignment
var ExcludedChurchNames = map[string]bool{
	"BCC Norge": true,
}

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	slog.Info("Starting user import from Members API",
		"min_age", minAge,
		"max_age", maxAge,
	)

	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if cfg.Auth0.Domain == "" || cfg.Auth0.ClientID == "" || cfg.Members.Domain == "" {
		slog.Error("Missing required configuration",
			"auth0_domain", cfg.Auth0.Domain,
			"members_domain", cfg.Members.Domain,
		)
		os.Exit(1)
	}

	auth0Client := auth0.New(auth0.Config{
		Domain:       cfg.Auth0.Domain,
		ClientID:     cfg.Auth0.ClientID,
		ClientSecret: cfg.Auth0.ClientSecret,
	})

	membersBreaker := gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
		Name:    "members-api-import",
		Timeout: 30 * time.Second,
	})

	membersClient := members.New(
		members.Config{Domain: cfg.Members.Domain},
		auth0Client,
		membersBreaker,
	)

	if err := importUsers(ctx, membersClient, db); err != nil {
		slog.Error("User import failed", "error", err)
		os.Exit(1)
	}

	slog.Info("User import completed successfully")
}

type churchMaps struct {
	orgUidToOrgID   map[uuid.UUID]int
	orgIDToChurchID map[int]string
	excludedOrgIDs  map[int]bool
	defaultChurchID string
}

func buildChurchMaps(ctx context.Context, membersClient *members.Client, db *database.DB) (*churchMaps, error) {
	slog.Info("Fetching organizations from Members API")
	orgs, err := membersClient.GetAllOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organizations: %w", err)
	}
	slog.Info("Fetched organizations", "count", len(orgs))

	orgUidToOrgID := make(map[uuid.UUID]int, len(orgs))
	excludedOrgIDs := make(map[int]bool)
	for _, org := range orgs {
		orgUidToOrgID[org.Uid] = org.OrgID
		if ExcludedChurchNames[org.Name] {
			excludedOrgIDs[org.OrgID] = true
		}
	}

	slog.Info("Fetching churches from database")
	churches, err := db.Queries.GetChurchesFilteredCursor(ctx, sqlc.GetChurchesFilteredCursorParams{
		Querylimit:   10000,
		Country:      "",
		Category:     "",
		Aftercursor:  "",
		Beforecursor: "",
		Isbackward:   false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch churches: %w", err)
	}
	slog.Info("Fetched churches", "count", len(churches))

	orgIDToChurchID := make(map[int]string, len(churches))
	for _, church := range churches {
		if church.ExternalID != nil {
			orgIDToChurchID[int(*church.ExternalID)] = church.ID
		}
	}

	defaultChurch, err := db.Queries.GetDefaultChurch(ctx)
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

func getExistingMembersIDs(ctx context.Context, db *database.DB) (map[string]bool, error) {
	rows, err := db.Pool.Query(ctx, "SELECT members_id FROM users")
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

func hasActiveAffiliation(affiliations []members.Affiliation) bool {
	now := time.Now()
	for _, aff := range affiliations {
		if aff.ValidFrom != nil && now.Before(*aff.ValidFrom) {
			continue
		}
		if aff.ValidTo != nil && now.After(*aff.ValidTo) {
			continue
		}
		return true
	}
	return false
}

func resolveChurchID(affiliations []members.Affiliation, maps *churchMaps) string {
	now := time.Now()
	for _, aff := range affiliations {
		if aff.Type != "Church" {
			continue
		}
		if aff.ValidFrom != nil && now.Before(*aff.ValidFrom) {
			continue
		}
		if aff.ValidTo != nil && now.After(*aff.ValidTo) {
			continue
		}

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

func normalizeGender(gender string) string {
	switch gender {
	case "Male", "male", "M":
		return "MALE"
	case "Female", "female", "F":
		return "FEMALE"
	default:
		return "UNKNOWN"
	}
}

func parseBirthdate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return t
}

type userRow struct {
	id          string
	membersID   string
	personUUID  uuid.UUID
	email       string
	name        string
	firstName   string
	lastName    string
	middleName  string
	displayName string
	gender      string
	birthdate   time.Time
	churchID    string
}

func importUsers(ctx context.Context, membersClient *members.Client, db *database.DB) error {
	now := time.Now()
	maxBirthdate := now.AddDate(-minAge, 0, 0).Format("2006-01-02")
	minBirthdate := now.AddDate(-maxAge-1, 0, 1).Format("2006-01-02")

	slog.Info("Birthdate range", "min", minBirthdate, "max", maxBirthdate)

	maps, err := buildChurchMaps(ctx, membersClient, db)
	if err != nil {
		return fmt.Errorf("failed to build church maps: %w", err)
	}

	existingMembersIDs, err := getExistingMembersIDs(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get existing members: %w", err)
	}
	slog.Info("Existing users in database", "count", len(existingMembersIDs))

	var totalFetched, totalImported, totalSkipped int
	offset := 0

	for {
		slog.Info("Fetching members batch", "offset", offset, "limit", apiPageSize)

		fetchedMembers, err := membersClient.GetMembersByBirthdateRange(ctx, minBirthdate, maxBirthdate, apiPageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch members at offset %d: %w", offset, err)
		}

		if len(fetchedMembers) == 0 {
			break
		}

		totalFetched += len(fetchedMembers)

		var batch []userRow
		for _, member := range fetchedMembers {
			if !hasActiveAffiliation(member.Affiliations) {
				totalSkipped++
				continue
			}

			membersID := strconv.Itoa(member.PersonID)
			if existingMembersIDs[membersID] {
				totalSkipped++
				continue
			}

			churchID := resolveChurchID(member.Affiliations, maps)

			batch = append(batch, userRow{
				id:          ulid.NewUserID(),
				membersID:   membersID,
				personUUID:  member.Uid,
				email:       member.Email,
				name:        member.DisplayName,
				firstName:   member.FirstName,
				lastName:    member.LastName,
				middleName:  member.MiddleName,
				displayName: member.DisplayName,
				gender:      normalizeGender(member.Gender),
				birthdate:   parseBirthdate(member.BirthDate),
				churchID:    churchID,
			})

			existingMembersIDs[membersID] = true
		}

		if len(batch) > 0 {
			inserted, err := insertUsersBatch(ctx, db, batch)
			if err != nil {
				slog.Error("Batch insert failed", "error", err, "batch_size", len(batch))
			} else {
				totalImported += inserted
				slog.Info("Batch inserted", "count", inserted)
			}
		}

		if len(fetchedMembers) < apiPageSize {
			break
		}

		offset += apiPageSize
	}

	slog.Info("Import completed",
		"total_fetched", totalFetched,
		"total_imported", totalImported,
		"total_skipped", totalSkipped,
	)

	return nil
}

func insertUsersBatch(ctx context.Context, db *database.DB, users []userRow) (int, error) {
	if len(users) == 0 {
		return 0, nil
	}

	var totalInserted int
	batches := chunkUsers(users, dbBatchSize)

	for i, batch := range batches {
		rows := make([][]any, 0, len(batch))

		for _, u := range batch {
			var personUUID pgtype.UUID
			personUUID.Bytes = u.personUUID
			personUUID.Valid = true

			rows = append(rows, []any{
				u.id,
				u.membersID,
				personUUID,
				u.email,
				u.name,
				stringPtr(u.firstName),
				stringPtr(u.lastName),
				stringPtr(u.middleName),
				stringPtr(u.displayName),
				u.gender,
				u.birthdate,
				u.churchID,
				nil,
				defaultLang,
			})
		}

		_, err := db.Pool.CopyFrom(
			ctx,
			pgx.Identifier{"users"},
			[]string{
				"id", "members_id", "person_uuid", "email", "name",
				"first_name", "last_name", "middle_name", "display_name",
				"gender", "birthdate", "church_id", "avatar_url", "language",
			},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return totalInserted, fmt.Errorf("batch %d insert failed: %w", i, err)
		}

		totalInserted += len(batch)
	}

	return totalInserted, nil
}

func chunkUsers(users []userRow, size int) [][]userRow {
	var chunks [][]userRow
	for i := 0; i < len(users); i += size {
		end := min(i+size, len(users))
		chunks = append(chunks, users[i:end])
	}
	return chunks
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
