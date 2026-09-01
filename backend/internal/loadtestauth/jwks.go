// Package loadtestauth holds the pieces shared between the server and
// cmd/loadtest/tokengen for the simulated Auth0 dance: tokengen signs
// Auth0-shaped RS256 tokens with a private key, and the server — given the
// same key via AUTH0_LOADTEST_PRIVATE_KEY — serves the matching JWKS at
// /jwks.json so AUTH0_JWKS_URL can point back at the server itself.
package loadtestauth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
)

// KeyID ties the minted tokens' kid header to the JWKS entry.
const KeyID = "wayfarer-loadtest"

// ParsePrivateKey accepts an RSA private key as PEM or base64-encoded PEM
// (PKCS#1 or PKCS#8), the same flexibility FIREBASE_SERVICE_ACCOUNT has.
func ParsePrivateKey(s string) (*rsa.PrivateKey, error) {
	data := []byte(s)
	if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
		data = decoded
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in load-test private key")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse load-test private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("load-test private key is not RSA")
	}
	return key, nil
}

// BuildJWKS renders the key's public half as a JWKS document.
func BuildJWKS(pub *rsa.PublicKey) ([]byte, error) {
	jwks := map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"use": "sig",
			"alg": "RS256",
			"kid": KeyID,
			"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
		}},
	}
	return json.MarshalIndent(jwks, "", "  ")
}

// EncodePrivateKey renders the key as base64-encoded PKCS#8 PEM, ready to
// paste into AUTH0_LOADTEST_PRIVATE_KEY.
func EncodePrivateKey(key *rsa.PrivateKey) (string, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	return base64.StdEncoding.EncodeToString(pemBytes), nil
}
