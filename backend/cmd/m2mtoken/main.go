package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultYears  = 1
	m2mUserID     = "M2M_SERVICE"
	defaultIssuer = "wayfarer"
)

type WayfarerClaims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"`
	jwt.RegisteredClaims
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Fprintln(os.Stderr, "Error: JWT_SECRET environment variable is required")
		os.Exit(1)
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = defaultIssuer
	}

	validYears := defaultYears
	if len(os.Args) >= 2 {
		years, err := strconv.ParseUint(os.Args[1], 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid years argument: %v\n", err)
			os.Exit(1)
		}
		validYears = int(years)
	}

	now := time.Now()
	claims := WayfarerClaims{
		UserID:    m2mUserID,
		UserRoles: []string{"m2m"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.AddDate(validYears, 0, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(tokenString)
}
