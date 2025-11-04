package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type WayfarerClaims struct {
	UserID   string `json:"user_id"`
	UserRole string `json:"user_role"`
	jwt.RegisteredClaims
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <user_id>")
		os.Exit(1)
	}

	userID := os.Args[1]
	secret := "your-secret-key-for-signing-wayfarer-jwts"

	now := time.Now()
	claims := WayfarerClaims{
		UserID:   userID,
		UserRole: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wayfarer",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
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
