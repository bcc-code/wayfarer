package authtoken

import (
	"errors"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret:         "current-secret",
		Issuer:         "wayfarer",
		AccessTokenTTL: 7 * 24 * time.Hour,
	}
}

func TestSignAndParse_RoundTrip(t *testing.T) {
	cfg := testConfig()
	now := time.Now()

	signed, err := Sign(cfg, NewClaims(cfg, "USUSER", "ASSESSION", []string{"user"}, now))
	require.NoError(t, err)

	claims, err := Parse(signed, cfg)
	require.NoError(t, err)
	assert.Equal(t, "USUSER", claims.UserID)
	assert.Equal(t, "USUSER", claims.Subject)
	assert.Equal(t, "ASSESSION", claims.SessionID)
	assert.Equal(t, []string{"user"}, claims.UserRoles)
	assert.Equal(t, jwt.ClaimStrings{Audience}, claims.Audience)
	assert.WithinDuration(t, now.Add(7*24*time.Hour), claims.ExpiresAt.Time, time.Second)
}

func TestSign_SetsKid(t *testing.T) {
	cfg := testConfig()
	signed, err := Sign(cfg, NewClaims(cfg, "USUSER", "", nil, time.Now()))
	require.NoError(t, err)

	token, _, err := jwt.NewParser().ParseUnverified(signed, &Claims{})
	require.NoError(t, err)
	assert.Equal(t, KeyID(cfg.Secret), token.Header["kid"])
}

func TestSign_RequiresSecret(t *testing.T) {
	cfg := testConfig()
	cfg.Secret = ""
	_, err := Sign(cfg, NewClaims(cfg, "USUSER", "", nil, time.Now()))
	assert.Error(t, err)
}

func TestKeyID_StableAndDistinct(t *testing.T) {
	assert.Equal(t, KeyID("a"), KeyID("a"))
	assert.NotEqual(t, KeyID("a"), KeyID("b"))
	assert.Len(t, KeyID("a"), 8)
}

func TestParse_PreviousSecretAfterRotation(t *testing.T) {
	old := testConfig()
	signed, err := Sign(old, NewClaims(old, "USUSER", "", nil, time.Now()))
	require.NoError(t, err)

	rotated := testConfig()
	rotated.Secret = "new-secret"

	_, err = Parse(signed, rotated)
	assert.ErrorIs(t, err, ErrInvalidToken, "old token rejected once previous secret is dropped")

	rotated.PreviousSecret = old.Secret
	claims, err := Parse(signed, rotated)
	require.NoError(t, err)
	assert.Equal(t, "USUSER", claims.UserID)
}

func TestParse_LegacyTokenWithoutKidOrAudience(t *testing.T) {
	cfg := testConfig()
	legacy := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:    "USLEGACY",
		UserRoles: []string{"user"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	signed, err := legacy.SignedString([]byte(cfg.Secret))
	require.NoError(t, err)

	claims, err := Parse(signed, cfg)
	require.NoError(t, err)
	assert.Equal(t, "USLEGACY", claims.UserID)
	assert.Empty(t, claims.SessionID)
}

func TestParse_Rejections(t *testing.T) {
	cfg := testConfig()
	now := time.Now()

	sign := func(t *testing.T, c Claims, secret string) string {
		t.Helper()
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
		s, err := tok.SignedString([]byte(secret))
		require.NoError(t, err)
		return s
	}

	tests := []struct {
		name  string
		token func(t *testing.T) string
	}{
		{"expired", func(t *testing.T) string {
			return sign(t, NewClaims(cfg, "USUSER", "", nil, now.Add(-8*24*time.Hour)), cfg.Secret)
		}},
		{"wrong secret", func(t *testing.T) string {
			return sign(t, NewClaims(cfg, "USUSER", "", nil, now), "other")
		}},
		{"wrong issuer", func(t *testing.T) string {
			c := NewClaims(cfg, "USUSER", "", nil, now)
			c.Issuer = "someone-else"
			return sign(t, c, cfg.Secret)
		}},
		{"wrong audience", func(t *testing.T) string {
			c := NewClaims(cfg, "USUSER", "", nil, now)
			c.Audience = jwt.ClaimStrings{"excalibur"}
			return sign(t, c, cfg.Secret)
		}},
		{"unknown kid", func(t *testing.T) string {
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, NewClaims(cfg, "USUSER", "", nil, now))
			tok.Header["kid"] = "deadbeef"
			s, err := tok.SignedString([]byte(cfg.Secret))
			require.NoError(t, err)
			return s
		}},
		{"HS512 instead of HS256", func(t *testing.T) string {
			tok := jwt.NewWithClaims(jwt.SigningMethodHS512, NewClaims(cfg, "USUSER", "", nil, now))
			s, err := tok.SignedString([]byte(cfg.Secret))
			require.NoError(t, err)
			return s
		}},
		{"garbage", func(t *testing.T) string { return "not.a.jwt" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.token(t), cfg)
			assert.ErrorIs(t, err, ErrInvalidToken)
		})
	}
}

func TestParse_EmergencyRevocation(t *testing.T) {
	cfg := testConfig()
	issued := time.Now().Add(-time.Hour)
	signed, err := Sign(cfg, NewClaims(cfg, "USUSER", "", nil, issued))
	require.NoError(t, err)

	t.Run("blocked entirely", func(t *testing.T) {
		c := cfg
		c.RevokedUsers = map[string]time.Time{"USUSER": {}}
		_, err := Parse(signed, c)
		assert.True(t, errors.Is(err, ErrRevoked))
	})

	t.Run("issued before cutoff", func(t *testing.T) {
		c := cfg
		c.RevokedUsers = map[string]time.Time{"USUSER": time.Now()}
		_, err := Parse(signed, c)
		assert.ErrorIs(t, err, ErrRevoked)
	})

	t.Run("issued after cutoff", func(t *testing.T) {
		c := cfg
		c.RevokedUsers = map[string]time.Time{"USUSER": issued.Add(-time.Hour)}
		_, err := Parse(signed, c)
		assert.NoError(t, err)
	})

	t.Run("other user listed", func(t *testing.T) {
		c := cfg
		c.RevokedUsers = map[string]time.Time{"USOTHER": {}}
		_, err := Parse(signed, c)
		assert.NoError(t, err)
	})
}

func BenchmarkParse(b *testing.B) {
	cfg := testConfig()
	signed, err := Sign(cfg, NewClaims(cfg, "USUSER", "ASSESSION", []string{"user"}, time.Now()))
	require.NoError(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Parse(signed, cfg); err != nil {
			b.Fatal(err)
		}
	}
}
