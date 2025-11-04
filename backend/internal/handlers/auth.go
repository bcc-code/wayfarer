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
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthHandler struct {
	DB            *database.DB
	Cfg           *config.Config
	JWKS          keyfunc.Keyfunc
	MembersClient *members.Client
}

// BrunstadTVClaims represents the JWT claims from Brunstad TV
type BrunstadTVClaims struct {
	ChurchID  int    `json:"church_id"`
	PersonID  string `json:"person_id"`
	FirstName string `json:"first_name"`
	Gender    string `json:"gender"`
	jwt.RegisteredClaims
}

// WayfarerClaims represents the JWT claims issued by Wayfarer
type WayfarerClaims struct {
	UserID   string `json:"user_id"`
	UserRole string `json:"user_role"`
	jwt.RegisteredClaims
}

// CallbackResponse is the JSON response returned by the callback endpoint
type CallbackResponse struct {
	Token string `json:"token"`
}

// Callback handles the OAuth callback from Brunstad TV
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

	// 2. Validate and parse Brunstad TV JWT
	claims, err := h.validateBrunstadTVToken(token)
	if err != nil {
		slog.Warn("callback: invalid token", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	slog.Info("callback: validated Brunstad TV token",
		"person_id", claims.PersonID,
		"church_id", claims.ChurchID,
	)

	// 3. Find church by external_id
	church, err := h.findChurchByExternalID(ctx, int32(claims.ChurchID))
	if err != nil {
		slog.Error("callback: failed to find church",
			"church_id", claims.ChurchID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find church"})
		return
	}

	// 4. Find or create user
	user, err := h.findOrCreateUser(ctx, claims, church.ID)
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

	// 5. Generate Wayfarer JWT
	wayfarerToken, err := h.generateWayfarerToken(user.ID)
	if err != nil {
		slog.Error("callback: failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate authentication token"})
		return
	}

	// 6. Return the token
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

// findChurchByExternalID finds a church by its external ID
func (h *AuthHandler) findChurchByExternalID(ctx context.Context, externalID int32) (*sqlc.GetChurchByExternalIDRow, error) {
	church, err := h.DB.Queries.GetChurchByExternalID(ctx, &externalID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("church with external_id %d not found", externalID)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return church, nil
}

// findOrCreateUser finds an existing user by members_id or creates a new one
func (h *AuthHandler) findOrCreateUser(ctx context.Context, claims *BrunstadTVClaims, churchID string) (*sqlc.GetUserByMembersIDRow, error) {
	// Try to find existing user
	user, err := h.DB.Queries.GetUserByMembersID(ctx, claims.PersonID)
	if err == nil {
		return user, nil
	}

	return nil, fmt.Errorf("user %s not in DB. Please fix", claims.PersonID)

	// If user doesn't exist, create new user
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("database error while finding user: %w", err)
	}

	slog.Info("callback: creating new user",
		"members_id", claims.PersonID,
		"church_id", churchID,
	)

	// Fetch member data from Members API
	var email string
	var displayName string
	var birthdate pgtype.Date

	personID, err := strconv.Atoi(claims.PersonID)
	if err != nil {
		slog.Warn("callback: invalid person_id format", "person_id", claims.PersonID, "error", err)
	} else if h.MembersClient != nil {
		member, err := h.MembersClient.Lookup(ctx, personID)
		if err != nil {
			slog.Warn("callback: failed to fetch member data from Members API",
				"person_id", personID,
				"error", err,
			)
		} else {
			// Use member data from API
			email = member.Email
			if member.DisplayName != "" {
				displayName = member.DisplayName
			} else {
				displayName = member.FirstName
			}

			// Parse birthdate if available
			if member.BirthDate != "" {
				parsedDate, err := time.Parse("2006-01-02", member.BirthDate)
				if err != nil {
					slog.Warn("callback: invalid birthdate format",
						"birthdate", member.BirthDate,
						"error", err,
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

			slog.Info("callback: fetched member data from Members API",
				"person_id", personID,
				"email", email,
				"has_birthdate", birthdate.Valid,
			)
		}
	}

	// Fallback to JWT claims if Members API data not available
	if displayName == "" {
		displayName = claims.FirstName
	}

	// Normalize gender to match database constraint (MALE, FEMALE)
	gender := normalizeGender(claims.Gender)

	// Create new user
	newUser, err := h.DB.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:        ulid.NewUserID(),
		MembersID: claims.PersonID,
		Email:     email,
		Name:      displayName,
		Gender:    gender,
		Birthdate: birthdate,
		ChurchID:  churchID,
		AvatarUrl: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert CreateUserRow to GetUserByMembersIDRow
	return &sqlc.GetUserByMembersIDRow{
		ID:        newUser.ID,
		MembersID: newUser.MembersID,
		Gender:    newUser.Gender,
		ChurchID:  newUser.ChurchID,
		Birthdate: newUser.Birthdate,
		Email:     newUser.Email,
		Name:      newUser.Name,
		AvatarUrl: newUser.AvatarUrl,
	}, nil
}

// generateWayfarerToken creates a new JWT for authentication against Wayfarer APIs
func (h *AuthHandler) generateWayfarerToken(userID string) (string, error) {
	now := time.Now()
	claims := WayfarerClaims{
		UserID:   userID,
		UserRole: "user",
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

// normalizeGender converts gender string to database format (MALE, FEMALE)
func normalizeGender(gender string) string {
	gender = strings.ToUpper(strings.TrimSpace(gender))
	if gender == "MALE" || gender == "M" {
		return "MALE"
	}
	if gender == "FEMALE" || gender == "F" {
		return "FEMALE"
	}
	// Default to MALE if unknown (you may want to handle this differently)
	return "MALE"
}
