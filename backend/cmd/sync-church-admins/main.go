package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
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
	apiPageSize = 500
)

// Target role UUIDs from BCC Members system
var targetRoleUUIDs = []uuid.UUID{
	uuid.MustParse("81c1f1ba-7e51-4782-94b3-872e97aebe74"), // Youth Leader
	uuid.MustParse("dbeccc5f-69f2-4268-bfd5-90f84630ffbc"), // BUK Contact
}

var roleNames = map[uuid.UUID]string{
	uuid.MustParse("81c1f1ba-7e51-4782-94b3-872e97aebe74"): "Youth Leader",
	uuid.MustParse("dbeccc5f-69f2-4268-bfd5-90f84630ffbc"): "BUK Contact",
}

type churchMaps struct {
	orgUidToOrgID   map[uuid.UUID]int
	orgIDToChurchID map[int]string
}

type syncStats struct {
	processed int
	imported  int
	assigned  int
	skipped   int
	errors    int
}

// parseBirthdate parses a birthdate string in YYYY-MM-DD format to pgtype.Date
func parseBirthdate(birthDate string) pgtype.Date {
	if birthDate == "" {
		return pgtype.Date{}
	}
	t, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		return pgtype.Date{}
	}
	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}

// normalizeGender converts member gender to Wayfarer's format
func normalizeGender(gender string) string {
	switch strings.ToLower(gender) {
	case "male", "m":
		return "MALE"
	case "female", "f":
		return "FEMALE"
	default:
		return "UNKNOWN"
	}
}

// ptr returns a pointer to the given string, or nil if empty
func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Print what would be done without making changes")
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	slog.Info("Starting sync of Youth Leaders and BUK Contacts to CHURCH_ADMIN",
		"dry_run", *dryRun,
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
		Name:    "members-api-sync-church-admins",
		Timeout: 30 * time.Second,
	})

	membersClient := members.New(
		members.Config{Domain: cfg.Members.Domain},
		auth0Client,
		membersBreaker,
	)

	if err := syncChurchAdmins(ctx, membersClient, db, *dryRun); err != nil {
		slog.Error("Sync failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Sync completed successfully")
}

func syncChurchAdmins(ctx context.Context, membersClient *members.Client, db *database.DB, dryRun bool) error {
	// Build church maps
	maps, err := buildChurchMaps(ctx, membersClient, db)
	if err != nil {
		return fmt.Errorf("failed to build church maps: %w", err)
	}

	// Get a SUPERADMIN user for assigned_by field
	superadmins, err := db.Queries.GetUsersWithRole(ctx, "SUPERADMIN")
	if err != nil {
		return fmt.Errorf("failed to get SUPERADMIN users: %w", err)
	}
	if len(superadmins) == 0 {
		return fmt.Errorf("no SUPERADMIN user found - cannot assign roles")
	}
	assignedBy := superadmins[0].ID
	slog.Info("Using SUPERADMIN for role assignment", "user_id", assignedBy, "email", superadmins[0].Email)

	// Process each target role
	totalStats := syncStats{}

	for _, roleUid := range targetRoleUUIDs {
		roleName := roleNames[roleUid]
		slog.Info("Processing role", "role_uid", roleUid, "role_name", roleName)

		stats, err := processRole(ctx, membersClient, db, maps, roleUid, assignedBy, dryRun)
		if err != nil {
			slog.Error("Failed to process role", "role_uid", roleUid, "error", err)
			continue
		}

		totalStats.processed += stats.processed
		totalStats.imported += stats.imported
		totalStats.assigned += stats.assigned
		totalStats.skipped += stats.skipped
		totalStats.errors += stats.errors
	}

	// Log final summary
	slog.Info("Sync completed",
		"total_processed", totalStats.processed,
		"total_imported", totalStats.imported,
		"total_assigned", totalStats.assigned,
		"total_skipped", totalStats.skipped,
		"total_errors", totalStats.errors,
		"dry_run", dryRun,
	)

	return nil
}

func buildChurchMaps(ctx context.Context, membersClient *members.Client, db *database.DB) (*churchMaps, error) {
	slog.Info("Fetching organizations from Members API")
	orgs, err := membersClient.GetAllOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organizations: %w", err)
	}
	slog.Info("Fetched organizations", "count", len(orgs))

	orgUidToOrgID := make(map[uuid.UUID]int, len(orgs))
	for _, org := range orgs {
		orgUidToOrgID[org.Uid] = org.OrgID
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

	return &churchMaps{
		orgUidToOrgID:   orgUidToOrgID,
		orgIDToChurchID: orgIDToChurchID,
	}, nil
}

func processRole(
	ctx context.Context,
	membersClient *members.Client,
	db *database.DB,
	maps *churchMaps,
	roleUid uuid.UUID,
	assignedBy string,
	dryRun bool,
) (*syncStats, error) {
	stats := &syncStats{}
	offset := 0

	for {
		slog.Info("Fetching members batch", "role_uid", roleUid, "offset", offset, "limit", apiPageSize)

		fetchedMembers, err := membersClient.GetMembersByRoleAssignment(ctx, roleUid, apiPageSize, offset)
		if err != nil {
			return stats, fmt.Errorf("failed to fetch members at offset %d: %w", offset, err)
		}

		if len(fetchedMembers) == 0 {
			break
		}

		for _, member := range fetchedMembers {
			stats.processed++

			// Find the matching role assignment to get the orgUid
			var orgUid uuid.UUID
			found := false
			for _, ra := range member.RoleAssignments {
				if ra.RoleUid == roleUid {
					orgUid = ra.OrgUid
					found = true
					break
				}
			}

			if !found {
				slog.Warn("No matching role assignment found in response",
					"person_uuid", member.Uid,
					"role_uid", roleUid,
				)
				stats.errors++
				continue
			}

			// Map orgUid -> orgID -> churchID
			orgID, ok := maps.orgUidToOrgID[orgUid]
			if !ok {
				slog.Warn("Organization not found in Members API",
					"person_uuid", member.Uid,
					"org_uid", orgUid,
				)
				stats.errors++
				continue
			}

			churchID, ok := maps.orgIDToChurchID[orgID]
			if !ok {
				slog.Warn("Church not found in Wayfarer database",
					"person_uuid", member.Uid,
					"org_uid", orgUid,
					"org_id", orgID,
				)
				stats.errors++
				continue
			}

			// Find user in Wayfarer by person_uuid
			var personUUID pgtype.UUID
			personUUID.Bytes = member.Uid
			personUUID.Valid = true

			user, err := db.Queries.GetUserByPersonUUID(ctx, personUUID)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					slog.Error("Failed to look up user",
						"person_uuid", member.Uid,
						"error", err,
					)
					stats.errors++
					continue
				}

				// User not found - auto-import
				name := member.DisplayName
				if name == "" {
					name = strings.TrimSpace(member.FirstName + " " + member.LastName)
				}
				if name == "" {
					name = member.Email
				}

				if dryRun {
					slog.Info("[DRY RUN] Would create user",
						"person_uuid", member.Uid,
						"person_id", member.PersonID,
						"email", member.Email,
						"name", name,
						"church_id", churchID,
					)
					// Create a temporary user for dry-run logging
					user = &sqlc.GetUserByPersonUUIDRow{
						ID:    "US_DRYRUN",
						Email: member.Email,
						Name:  name,
					}
					stats.imported++
				} else {
					newUserID := ulid.NewUserID()
					createdUser, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
						ID:          newUserID,
						MembersID:   strconv.Itoa(member.PersonID),
						PersonUuid:  personUUID,
						Email:       member.Email,
						Name:        name,
						FirstName:   ptr(member.FirstName),
						LastName:    ptr(member.LastName),
						MiddleName:  ptr(member.MiddleName),
						DisplayName: ptr(member.DisplayName),
						Gender:      normalizeGender(member.Gender),
						Birthdate:   parseBirthdate(member.BirthDate),
						ChurchID:    churchID,
						AvatarUrl:   nil,
						Language:    "no",
					})
					if err != nil {
						slog.Error("Failed to create user",
							"person_uuid", member.Uid,
							"email", member.Email,
							"error", err,
						)
						stats.errors++
						continue
					}

					slog.Info("Created user",
						"user_id", createdUser.ID,
						"person_uuid", member.Uid,
						"email", member.Email,
						"name", name,
						"church_id", churchID,
					)

					// Convert created user to the row type used below
					user = &sqlc.GetUserByPersonUUIDRow{
						ID:          createdUser.ID,
						MembersID:   createdUser.MembersID,
						PersonUuid:  createdUser.PersonUuid,
						Gender:      createdUser.Gender,
						ChurchID:    createdUser.ChurchID,
						Birthdate:   createdUser.Birthdate,
						Email:       createdUser.Email,
						Name:        createdUser.Name,
						FirstName:   createdUser.FirstName,
						LastName:    createdUser.LastName,
						MiddleName:  createdUser.MiddleName,
						DisplayName: createdUser.DisplayName,
						AvatarUrl:   createdUser.AvatarUrl,
						Language:    createdUser.Language,
						CreatedAt:   createdUser.CreatedAt,
					}
					stats.imported++
				}
			}

			// Check if already has CHURCH_ADMIN for this church
			hasRole, err := db.Queries.HasRoleInChurch(ctx, sqlc.HasRoleInChurchParams{
				UserID:   user.ID,
				Role:     "CHURCH_ADMIN",
				ChurchID: &churchID,
			})
			if err != nil {
				slog.Error("Failed to check existing role",
					"user_id", user.ID,
					"church_id", churchID,
					"error", err,
				)
				stats.errors++
				continue
			}

			if hasRole {
				slog.Debug("User already has CHURCH_ADMIN role",
					"user_id", user.ID,
					"church_id", churchID,
				)
				stats.skipped++
				continue
			}

			// Assign CHURCH_ADMIN role
			if dryRun {
				slog.Info("[DRY RUN] Would assign CHURCH_ADMIN role",
					"user_id", user.ID,
					"user_email", user.Email,
					"user_name", user.Name,
					"church_id", churchID,
					"role_name", roleNames[roleUid],
				)
			} else {
				_, err := db.Queries.AssignRole(ctx, sqlc.AssignRoleParams{
					ID:         ulid.NewUserRoleID(),
					UserID:     user.ID,
					Role:       "CHURCH_ADMIN",
					ChurchID:   &churchID,
					ProjectID:  nil,
					TeamID:     nil,
					AssignedBy: assignedBy,
					AssignedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				})
				if err != nil {
					slog.Error("Failed to assign role",
						"user_id", user.ID,
						"church_id", churchID,
						"error", err,
					)
					stats.errors++
					continue
				}

				slog.Info("Assigned CHURCH_ADMIN role",
					"user_id", user.ID,
					"user_email", user.Email,
					"user_name", user.Name,
					"church_id", churchID,
					"role_name", roleNames[roleUid],
				)
			}

			stats.assigned++
		}

		if len(fetchedMembers) < apiPageSize {
			break
		}

		offset += apiPageSize
	}

	slog.Info("Role processing completed",
		"role_uid", roleUid,
		"role_name", roleNames[roleUid],
		"processed", stats.processed,
		"imported", stats.imported,
		"assigned", stats.assigned,
		"skipped", stats.skipped,
		"errors", stats.errors,
	)

	return stats, nil
}
