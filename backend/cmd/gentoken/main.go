package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type WayfarerClaims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"`
	jwt.RegisteredClaims
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <user_id> (<valid days>)")
		os.Exit(1)
	}

	userID := os.Args[1]
	secret := "your-secret-key-for-signing-wayfarer-jwts"

	validDays := 5
	if len(os.Args) >= 3 {
		v, err := strconv.ParseUint(os.Args[2], 10, 32)
		if err != nil {
			panic("bad number")
		}

		validDays = int(v)
	}

	now := time.Now()
	claims := WayfarerClaims{
		UserID:    userID,
		UserRoles: []string{"user", "admin", "superadmin"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wayfarer",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(validDays) * 24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(tokenString)
}
