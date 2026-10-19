package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// mintWayfarerToken signs a Wayfarer access token for userID with the given
// roles, in the legacy shape (no kid, no aud) the server still accepts.
func mintWayfarerToken(secret, userID string, roles []string, now time.Time, validity time.Duration) (string, error) {
	claims := WayfarerClaims{
		UserID:    userID,
		UserRoles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wayfarer",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(validity)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
