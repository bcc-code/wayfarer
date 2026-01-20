package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthHandler struct {
	DB                        *database.DB
	Cfg                       *config.Config
	JWKS                      keyfunc.Keyfunc
	Auth0JWKS                 keyfunc.Keyfunc
	MembersClient             *members.Client
	RoleService               *services.RoleService
	ContentAchievementService *services.ContentAchievementService
}

// BrunstadTVClaims represents the JWT claims from Brunstad TV
type BrunstadTVClaims struct {
	ChurchID   int    `json:"church_id"`
	PersonID   string `json:"person_id"`
	PersonUUID string `json:"person_uuid"`
	FirstName  string `json:"first_name"`
	Gender     string `json:"gender"`
	jwt.RegisteredClaims
}

// Auth0AppMetadata represents the app_metadata claim from Auth0 (login.bcc.no)
type Auth0AppMetadata struct {
	HasMembership bool `json:"hasMembership"`
	PersonID      int  `json:"personId"`
}

// Auth0Claims represents the JWT claims from Auth0 (login.bcc.no)
type Auth0Claims struct {
	ChurchID    int              `json:"https://login.bcc.no/claims/churchId"`
	PersonID    int              `json:"https://login.bcc.no/claims/personId"`
	PersonUUID  string           `json:"https://login.bcc.no/claims/personUid"`
	AppMetadata Auth0AppMetadata `json:"https://members.bcc.no/app_metadata"`
	jwt.RegisteredClaims
}

// WayfarerClaims represents the JWT claims issued by Wayfarer
type WayfarerClaims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"` // All roles the user has
	jwt.RegisteredClaims
}

// CallbackResponse is the JSON response returned by the callback endpoint
type CallbackResponse struct {
	Token string `json:"token"`
}

// Callback handles the OAuth callback from Brunstad TV or Auth0
// It validates the incoming JWT, finds or creates the user, and returns a Wayfarer JWT
func (h *AuthHandler) Callback(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Extract token from query parameter
	token := c.Query("token")
	if token == "" {
		slog.Warn("callback: missing token parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "token parameter is required"})
		return
	}

	// 2. Validate and parse JWT - try Auth0 first, then fall back to Brunstad TV
	var claims *BrunstadTVClaims
	isAuth0Token := false

	auth0Claims, auth0Err := h.validateAuth0Token(token)
	if auth0Err == nil {
		// Check membership requirement
		if !auth0Claims.AppMetadata.HasMembership {
			slog.Warn("callback: user does not have membership",
				"person_id", auth0Claims.PersonID,
			)
			c.JSON(http.StatusForbidden, gin.H{"error": "membership required"})
			return
		}

		// Auth0 token validated successfully, convert to BrunstadTVClaims format
		slog.Info("callback: validated Auth0 token",
			"person_id", auth0Claims.PersonID,
			"church_id", auth0Claims.ChurchID,
		)
		claims = &BrunstadTVClaims{
			ChurchID:         auth0Claims.ChurchID,
			PersonID:         strconv.Itoa(auth0Claims.PersonID),
			PersonUUID:       auth0Claims.PersonUUID,
			FirstName:        "", // Not provided in Auth0 token, will be fetched from Members API
			Gender:           "", // Not provided in Auth0 token, will be fetched from Members API
			RegisteredClaims: auth0Claims.RegisteredClaims,
		}
		isAuth0Token = true
	} else {
		// Auth0 validation failed, try Brunstad TV
		brunstadClaims, brunstadErr := h.validateBrunstadTVToken(token)
		if brunstadErr != nil {
			slog.Warn("callback: invalid token",
				"auth0_error", auth0Err,
				"brunstad_error", brunstadErr,
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		claims = brunstadClaims
		slog.Info("callback: validated Brunstad TV token",
			"person_id", claims.PersonID,
			"church_id", claims.ChurchID,
		)
	}

	// 3. Fetch member data from Members API (needed for Auth0 tokens, optional for Brunstad TV)
	var member *members.Member
	var gender string

	personID, parseErr := strconv.Atoi(claims.PersonID)
	if parseErr != nil {
		slog.Warn("callback: invalid person_id format", "person_id", claims.PersonID, "error", parseErr)
	} else if h.MembersClient != nil {
		var err error
		member, err = h.MembersClient.Lookup(ctx, personID)
		if err != nil {
			slog.Warn("callback: failed to fetch member data from Members API",
				"person_id", personID,
				"error", err,
			)
			// Continue with fallback for Auth0 tokens
		}
	}

	// 4. Determine gender
	if member != nil && member.Gender != "" {
		gender = normalizeGender(member.Gender)
	} else if claims.Gender != "" {
		gender = normalizeGender(claims.Gender)
	} else {
		gender = "UNKNOWN"
	}

	// 5. Find church
	var church *sqlc.GetChurchByExternalIDRow
	var err error

	if isAuth0Token && claims.ChurchID == 0 {
		// Auth0 token without churchId - get from member affiliations
		if member != nil {
			church, err = h.findChurchFromAffiliations(ctx, member.Affiliations)
			if err != nil {
				slog.Warn("callback: failed to find church from affiliations, using default",
					"person_id", claims.PersonID,
					"error", err,
				)
				church, err = h.GetOrCreateDefaultChurch(ctx)
				if err != nil {
					slog.Error("callback: failed to get default church", "error", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find church"})
					return
				}
			}
		} else {
			// No member data available - use default church
			slog.Warn("callback: no member data available, using default church",
				"person_id", claims.PersonID,
			)
			church, err = h.GetOrCreateDefaultChurch(ctx)
			if err != nil {
				slog.Error("callback: failed to get default church", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find church"})
				return
			}
		}
	} else {
		// Brunstad TV token or Auth0 with churchId - use external_id from token
		church, err = h.findChurchByExternalID(ctx, int32(claims.ChurchID))
		if err != nil {
			slog.Error("callback: failed to find church",
				"church_id", claims.ChurchID,
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find church"})
			return
		}
	}

	// 6. Find or create user
	user, err := h.findOrCreateUser(ctx, claims, church.ID, member, gender)
	if err != nil {
		slog.Error("callback: failed to find or create user",
			"person_id", claims.PersonID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process user"})
		return
	}

	slog.Info("callback: user authenticated",
		"user_id", user.ID,
		"members_id", user.MembersID,
	)

	// 7. Generate Wayfarer JWT
	wayfarerToken, err := h.generateWayfarerToken(user.ID)
	if err != nil {
		slog.Error("callback: failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate authentication token"})
		return
	}

	// 8. Return the token
	c.JSON(http.StatusOK, CallbackResponse{
		Token: wayfarerToken,
	})
}

// validateBrunstadTVToken validates the JWT from Brunstad TV using JWKS
func (h *AuthHandler) validateBrunstadTVToken(tokenString string) (*BrunstadTVClaims, error) {
	// Parse and validate the token using JWKS
	token, err := jwt.ParseWithClaims(tokenString, &BrunstadTVClaims{}, h.JWKS.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	claims, ok := token.Claims.(*BrunstadTVClaims)
	if !ok {
		return nil, errors.New("failed to parse claims")
	}

	// Verify issuer
	if claims.Issuer != h.Cfg.JWT.BrunstadTVIssuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", h.Cfg.JWT.BrunstadTVIssuer, claims.Issuer)
	}

	// Verify timestamps (exp and iat are automatically validated by jwt library)

	return claims, nil
}

// validateAuth0Token validates the JWT from Auth0 (login.bcc.no) using JWKS
func (h *AuthHandler) validateAuth0Token(tokenString string) (*Auth0Claims, error) {
	if h.Auth0JWKS == nil {
		return nil, errors.New("Auth0 JWKS not configured")
	}

	// Parse and validate the token using Auth0 JWKS
	token, err := jwt.ParseWithClaims(tokenString, &Auth0Claims{}, h.Auth0JWKS.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	claims, ok := token.Claims.(*Auth0Claims)
	if !ok {
		return nil, errors.New("failed to parse claims")
	}

	// Verify issuer
	if claims.Issuer != h.Cfg.JWT.Auth0Issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", h.Cfg.JWT.Auth0Issuer, claims.Issuer)
	}

	// Verify timestamps (exp and iat are automatically validated by jwt library)

	return claims, nil
}

// findOrCreateChurch finds a church by its external ID, or creates it from the Members API if not found
func (h *AuthHandler) findChurchByExternalID(ctx context.Context, externalID int32) (*sqlc.GetChurchByExternalIDRow, error) {
	church, err := h.DB.Queries.GetChurchByExternalID(ctx, &externalID)
	if err == nil {
		return church, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Church not found, try to create it from Members API
	if h.MembersClient == nil {
		return nil, fmt.Errorf("church with external_id %d not found and Members API not configured", externalID)
	}

	slog.Info("callback: church not found, fetching from Members API", "external_id", externalID)

	org, err := h.MembersClient.GetOrganizationByOrgID(ctx, int(externalID))

	var churchName, country, category string
	if err != nil {
		slog.Warn("callback: failed to fetch organization from Members API, creating placeholder",
			"external_id", externalID,
			"error", err,
		)
		churchName = fmt.Sprintf("Church %d", externalID)
		country = "Unknown"
		category = "S"
	} else {
		churchName = org.Name
		country = org.VisitingAddress.CountryCode
		if country == "" {
			country = "Unknown"
		}
		category = "S"
	}

	newChurch, err := h.DB.Queries.CreateChurch(ctx, sqlc.CreateChurchParams{
		ID:         ulid.NewChurchID(),
		ExternalID: &externalID,
		Name:       churchName,
		Country:    country,
		Category:   category,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create church: %w", err)
	}

	slog.Info("callback: created new church from Members API",
		"church_id", newChurch.ID,
		"name", newChurch.Name,
		"external_id", externalID,
	)

	return &sqlc.GetChurchByExternalIDRow{
		ID:         newChurch.ID,
		ExternalID: newChurch.ExternalID,
		Name:       newChurch.Name,
		Country:    newChurch.Country,
		Category:   newChurch.Category,
	}, nil
}

// findOrCreateUser finds an existing user by person_uuid or members_id, or creates a new one
func (h *AuthHandler) findOrCreateUser(ctx context.Context, claims *BrunstadTVClaims, churchID string, member *members.Member, gender string) (*sqlc.GetUserByMembersIDRow, error) {
	// Parse person_uuid from claims if present
	var personUUID pgtype.UUID
	if claims.PersonUUID != "" {
		parsed, parseErr := uuid.Parse(claims.PersonUUID)
		if parseErr == nil {
			personUUID = pgtype.UUID{Bytes: parsed, Valid: true}
		} else {
			slog.Warn("callback: invalid person_uuid format in JWT", "person_uuid", claims.PersonUUID, "error", parseErr)
		}
	}

	// Try to find by person_uuid first (preferred)
	if personUUID.Valid {
		slog.Debug("auth: looking up user by person_uuid",
			"person_uuid", uuid.UUID(personUUID.Bytes).String(),
			"members_id", claims.PersonID,
		)
		user, err := h.DB.Queries.GetUserByPersonUUID(ctx, personUUID)
		if err == nil {
			slog.Debug("auth: found existing user by person_uuid",
				"user_id", user.ID,
				"person_uuid", uuid.UUID(personUUID.Bytes).String(),
			)
			// Convert GetUserByPersonUUIDRow to GetUserByMembersIDRow
			return &sqlc.GetUserByMembersIDRow{
				ID:          user.ID,
				MembersID:   user.MembersID,
				PersonUuid:  user.PersonUuid,
				Gender:      user.Gender,
				ChurchID:    user.ChurchID,
				Birthdate:   user.Birthdate,
				Email:       user.Email,
				Name:        user.Name,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				MiddleName:  user.MiddleName,
				DisplayName: user.DisplayName,
				AvatarUrl:   user.AvatarUrl,
			}, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("database error while finding user by person_uuid: %w", err)
		}
		slog.Debug("auth: user not found by person_uuid, trying members_id",
			"person_uuid", uuid.UUID(personUUID.Bytes).String(),
		)
	}

	// Fallback: Try to find by members_id (old numeric ID)
	slog.Debug("auth: looking up user by members_id",
		"members_id", claims.PersonID,
	)
	user, err := h.DB.Queries.GetUserByMembersID(ctx, claims.PersonID)
	if err == nil {
		slog.Debug("auth: found existing user by members_id",
			"user_id", user.ID,
			"members_id", claims.PersonID,
		)
		// User exists but may not have person_uuid - update if we have it
		if personUUID.Valid && !user.PersonUuid.Valid {
			updateErr := h.DB.Queries.UpdateUserPersonUUID(ctx, sqlc.UpdateUserPersonUUIDParams{
				ID:         user.ID,
				PersonUuid: personUUID,
			})
			if updateErr != nil {
				slog.Warn("callback: failed to update person_uuid for existing user", "user_id", user.ID, "error", updateErr)
			} else {
				slog.Info("callback: updated person_uuid for existing user", "user_id", user.ID, "person_uuid", uuid.UUID(personUUID.Bytes).String())
				user.PersonUuid = personUUID
			}
		}
		return user, nil
	}

	// If user doesn't exist, create new user
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("database error while finding user: %w", err)
	}

	slog.Debug("auth: user not found by members_id, will create new user",
		"members_id", claims.PersonID,
	)

	slog.Info("callback: creating new user",
		"members_id", claims.PersonID,
		"church_id", churchID,
	)

	// Extract member data if available
	var email string
	var firstName string
	var lastName string
	var middleName string
	var displayName string
	var computedName string
	var birthdate pgtype.Date

	if member != nil {
		// Use member data from API
		email = member.Email
		firstName = member.FirstName
		lastName = member.LastName
		middleName = member.MiddleName
		displayName = member.DisplayName

		// Get person_uuid from Members API if not in JWT
		if !personUUID.Valid && member.Uid != uuid.Nil {
			personUUID = pgtype.UUID{Bytes: member.Uid, Valid: true}
			slog.Info("callback: obtained person_uuid from Members API", "person_uuid", member.Uid.String())
		}

		// Compute name for backward compatibility
		if firstName != "" && lastName != "" {
			computedName = firstName + " " + lastName
		} else if firstName != "" {
			computedName = firstName
		} else if displayName != "" {
			computedName = displayName
		}

		// Generate display name if API didn't provide one
		if displayName == "" {
			displayName = generateDisplayName(firstName, lastName, computedName)
		}

		// Parse birthdate if available
		if member.BirthDate != "" {
			parsedDate, parseErr := time.Parse("2006-01-02", member.BirthDate)
			if parseErr != nil {
				slog.Warn("callback: invalid birthdate format",
					"birthdate", member.BirthDate,
					"error", parseErr,
				)
			} else {
				// Validate that the birthdate is reasonable (between 1900 and today)
				now := time.Now()
				minDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

				if parsedDate.Before(minDate) {
					slog.Warn("callback: birthdate too far in the past",
						"birthdate", member.BirthDate,
					)
				} else if parsedDate.After(now) {
					slog.Warn("callback: birthdate is in the future",
						"birthdate", member.BirthDate,
					)
				} else {
					// Valid birthdate, store it
					birthdate = pgtype.Date{
						Time:  parsedDate,
						Valid: true,
					}
				}
			}
		}

		slog.Info("callback: using member data from Members API",
			"email", email,
			"has_birthdate", birthdate.Valid,
		)
	}

	// Fallback to JWT claims if Members API data not available
	if computedName == "" {
		computedName = claims.FirstName
		firstName = claims.FirstName
	}

	// Ensure display name is set (fallback case when API wasn't available)
	if displayName == "" {
		displayName = generateDisplayName(firstName, lastName, computedName)
	}

	// Helper function to convert string to *string
	toStringPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}

	// Create new user
	newUser, err := h.DB.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:          ulid.NewUserID(),
		MembersID:   claims.PersonID,
		PersonUuid:  personUUID,
		Email:       email,
		Name:        computedName,
		FirstName:   toStringPtr(firstName),
		LastName:    toStringPtr(lastName),
		MiddleName:  toStringPtr(middleName),
		DisplayName: toStringPtr(displayName),
		Gender:      gender,
		Birthdate:   birthdate,
		ChurchID:    churchID,
		AvatarUrl:   nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	slog.Debug("auth: user created successfully",
		"user_id", newUser.ID,
		"members_id", newUser.MembersID,
		"person_uuid_valid", personUUID.Valid,
	)

	// Process any pending consent events for this user (using person_uuid if available)
	if personUUID.Valid {
		personUUIDStr := uuid.UUID(personUUID.Bytes).String()
		slog.Debug("auth: processing pending events for new user",
			"user_id", newUser.ID,
			"person_uuid", personUUIDStr,
		)
		h.ProcessPendingConsentEvents(ctx, newUser.ID, personUUIDStr)
		if h.ContentAchievementService != nil {
			slog.Debug("auth: processing pending content events",
				"user_id", newUser.ID,
				"person_uuid", personUUIDStr,
			)
			h.ContentAchievementService.ProcessPendingContentEvents(ctx, newUser.ID, personUUID)
			slog.Debug("auth: finished processing pending content events",
				"user_id", newUser.ID,
			)
		}
	} else {
		slog.Debug("auth: skipping pending event processing - no person_uuid",
			"user_id", newUser.ID,
		)
	}

	// Convert CreateUserRow to GetUserByMembersIDRow
	return &sqlc.GetUserByMembersIDRow{
		ID:          newUser.ID,
		MembersID:   newUser.MembersID,
		PersonUuid:  newUser.PersonUuid,
		Gender:      newUser.Gender,
		ChurchID:    newUser.ChurchID,
		Birthdate:   newUser.Birthdate,
		Email:       newUser.Email,
		Name:        newUser.Name,
		FirstName:   newUser.FirstName,
		LastName:    newUser.LastName,
		MiddleName:  newUser.MiddleName,
		DisplayName: newUser.DisplayName,
		AvatarUrl:   newUser.AvatarUrl,
	}, nil
}

// generateWayfarerToken creates a new JWT for authentication against Wayfarer APIs
func (h *AuthHandler) generateWayfarerToken(userID string) (string, error) {
	now := time.Now()

	// Load all user roles from database
	userRoles, err := h.RoleService.LoadUserRoles(context.Background(), userID)
	if err != nil {
		slog.Warn("Failed to load user roles, defaulting to 'user'", "user_id", userID, "error", err)
		return "", fmt.Errorf("failed to load user roles, defaulting to 'user'")
	}

	// Convert roles to string array
	roles := make([]string, 0, len(userRoles))
	if len(userRoles) > 0 {
		for _, role := range userRoles {
			roles = append(roles, strings.ToLower(role.Role))
		}
	} else {
		// Default to USER role if no roles found
		roles = append(roles, strings.ToLower(string(services.RoleUser)))
	}

	claims := WayfarerClaims{
		UserID:    userID,
		UserRoles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.Cfg.JWT.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // Token valid for 24 hours
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.Cfg.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// normalizeGender converts gender string to database format (MALE, FEMALE, UNKNOWN)
func normalizeGender(gender string) string {
	gender = strings.ToUpper(strings.TrimSpace(gender))
	if gender == "MALE" || gender == "M" {
		return "MALE"
	}
	if gender == "FEMALE" || gender == "F" {
		return "FEMALE"
	}
	return "UNKNOWN"
}

// ExcludedChurchNames contains organization names that should not be used for church assignment
var ExcludedChurchNames = []string{"BCC Norge"}

// findChurchFromAffiliations finds the first valid non-excluded church from member affiliations.
// It iterates through all active affiliations and returns the first one that is not excluded.
func (h *AuthHandler) findChurchFromAffiliations(ctx context.Context, affiliations []members.Affiliation) (*sqlc.GetChurchByExternalIDRow, error) {
	orgUIDs := members.GetActiveAffiliationOrgUIDs(affiliations)
	if len(orgUIDs) == 0 {
		return nil, fmt.Errorf("no active affiliations found")
	}

	for _, orgUID := range orgUIDs {
		church, err := h.findChurchByOrgUID(ctx, orgUID)
		if err != nil {
			slog.Debug("findChurchFromAffiliations: skipping affiliation",
				"org_uid", orgUID,
				"error", err,
			)
			continue
		}
		return church, nil
	}

	return nil, fmt.Errorf("no valid church found from %d affiliations (all excluded or invalid)", len(orgUIDs))
}

// findChurchByOrgUID finds a church by looking up the org UUID in Members API first
func (h *AuthHandler) findChurchByOrgUID(ctx context.Context, orgUID uuid.UUID) (*sqlc.GetChurchByExternalIDRow, error) {
	if h.MembersClient == nil {
		return nil, fmt.Errorf("members API not configured")
	}

	org, err := h.MembersClient.GetOrganizationByUID(ctx, orgUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization from Members API: %w", err)
	}

	// Check if the organization name is excluded
	for _, excluded := range ExcludedChurchNames {
		if org.Name == excluded {
			return nil, fmt.Errorf("organization %q is excluded from church assignment", org.Name)
		}
	}

	return h.findChurchByExternalID(ctx, int32(org.OrgID))
}

// DefaultChurchName is the name used for the fallback church
const DefaultChurchName = "Unknown Church"

// GetOrCreateDefaultChurch returns the default church, creating it if it doesn't exist
func (h *AuthHandler) GetOrCreateDefaultChurch(ctx context.Context) (*sqlc.GetChurchByExternalIDRow, error) {
	// Try to find existing default church (external_id IS NULL)
	church, err := h.DB.Queries.GetDefaultChurch(ctx)
	if err == nil {
		return &sqlc.GetChurchByExternalIDRow{
			ID:         church.ID,
			ExternalID: church.ExternalID,
			Name:       church.Name,
			Country:    church.Country,
			Category:   church.Category,
		}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("database error while finding default church: %w", err)
	}

	// Create default church
	newChurch, err := h.DB.Queries.CreateChurch(ctx, sqlc.CreateChurchParams{
		ID:         ulid.NewChurchID(),
		ExternalID: nil,
		Name:       DefaultChurchName,
		Country:    "Unknown",
		Category:   "S",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create default church: %w", err)
	}

	slog.Info("callback: created default church", "church_id", newChurch.ID)

	return &sqlc.GetChurchByExternalIDRow{
		ID:         newChurch.ID,
		ExternalID: newChurch.ExternalID,
		Name:       newChurch.Name,
		Country:    newChurch.Country,
		Category:   newChurch.Category,
	}, nil
}

// generateDisplayName creates a display name in the format "FirstName L." if both names are provided,
// otherwise returns the fallback name
func generateDisplayName(firstName, lastName, fallbackName string) string {
	if firstName != "" && lastName != "" {
		return firstName + " " + string([]rune(lastName)[0]) + "."
	}
	return fallbackName
}

// ProcessPendingConsentEvents processes any pending consent events for a newly registered user
// personUUID is the person's UUID string used to match pending events
func (h *AuthHandler) ProcessPendingConsentEvents(ctx context.Context, userID, personUUID string) {
	// Get all pending consent events for this person_uuid (stored in members_id field of pending_consent_events)
	pendingEvents, err := h.DB.Queries.GetPendingConsentEventsByMembersID(ctx, personUUID)
	if err != nil {
		slog.Error("auth: failed to get pending consent events",
			"user_id", userID,
			"person_uuid", personUUID,
			"error", err,
		)
		return
	}

	if len(pendingEvents) == 0 {
		return
	}

	slog.Info("auth: processing pending consent events for new user",
		"user_id", userID,
		"person_uuid", personUUID,
		"count", len(pendingEvents),
	)

	for _, event := range pendingEvents {
		// Get the latest published consent by key
		consent, err := h.DB.Queries.GetLatestPublishedConsentByKey(ctx, event.ConsentKey)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.Warn("auth: consent not found for pending event, skipping",
					"consent_key", event.ConsentKey,
					"pending_id", event.ID,
				)
				continue
			}
			slog.Error("auth: failed to get consent for pending event",
				"consent_key", event.ConsentKey,
				"error", err,
			)
			continue
		}

		// Create consent history record
		historyID := ulid.NewUserConsentHistoryID()

		_, err = h.DB.Queries.CreateUserConsentHistory(ctx, sqlc.CreateUserConsentHistoryParams{
			ID:         historyID,
			UserID:     userID,
			ConsentID:  consent.ID,
			ConsentKey: event.ConsentKey,
			Action:     event.Action,
			OccurredAt: event.OccurredAt,
			Source:     event.Source,
		})
		if err != nil {
			slog.Error("auth: failed to create consent history from pending event",
				"error", err,
				"user_id", userID,
				"consent_key", event.ConsentKey,
				"pending_id", event.ID,
			)
			continue
		}

		slog.Info("auth: created consent history from pending event",
			"history_id", historyID,
			"user_id", userID,
			"consent_key", event.ConsentKey,
			"action", event.Action,
		)
	}

	// Delete all processed pending events
	err = h.DB.Queries.DeletePendingConsentEventsByMembersID(ctx, personUUID)
	if err != nil {
		slog.Error("auth: failed to delete processed pending consent events",
			"person_uuid", personUUID,
			"error", err,
		)
	}
}
