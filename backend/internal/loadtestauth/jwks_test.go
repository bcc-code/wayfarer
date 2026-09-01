package loadtestauth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeParseRoundTrip(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	encoded, err := EncodePrivateKey(key)
	require.NoError(t, err)

	// Base64 PEM (the env-var form)
	parsed, err := ParsePrivateKey(encoded)
	require.NoError(t, err)
	assert.True(t, key.Equal(parsed))

	// Raw PEM also accepted
	pemBytes, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)
	parsed, err = ParsePrivateKey(string(pemBytes))
	require.NoError(t, err)
	assert.True(t, key.Equal(parsed))
}

func TestParsePrivateKey_Invalid(t *testing.T) {
	_, err := ParsePrivateKey("not a key")
	assert.Error(t, err)
}

func TestBuildJWKS_MatchesKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwksBytes, err := BuildJWKS(&key.PublicKey)
	require.NoError(t, err)

	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	require.NoError(t, json.Unmarshal(jwksBytes, &jwks))
	require.Len(t, jwks.Keys, 1)
	assert.Equal(t, "RSA", jwks.Keys[0].Kty)
	assert.Equal(t, "RS256", jwks.Keys[0].Alg)
	assert.Equal(t, KeyID, jwks.Keys[0].Kid)

	nBytes, err := base64.RawURLEncoding.DecodeString(jwks.Keys[0].N)
	require.NoError(t, err)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwks.Keys[0].E)
	require.NoError(t, err)
	assert.Equal(t, 0, key.PublicKey.N.Cmp(new(big.Int).SetBytes(nBytes)))
	assert.Equal(t, int64(key.PublicKey.E), new(big.Int).SetBytes(eBytes).Int64())
}
