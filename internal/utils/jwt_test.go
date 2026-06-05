package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:    "test-secret-key-for-testing",
		JWTExpiresIn: 15 * time.Minute,
	}
}

// ==================== GenerateToken ====================

func TestGenerateToken_Success(t *testing.T) {
	cfg := testConfig()
	userID := uuid.New()

	token, err := GenerateToken(cfg, userID, "testuser")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateToken_ContainsUserID(t *testing.T) {
	cfg := testConfig()
	userID := uuid.New()

	token, err := GenerateToken(cfg, userID, "testuser")
	require.NoError(t, err)

	claims, err := ValidateToken(cfg, token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestGenerateToken_ContainsUsername(t *testing.T) {
	cfg := testConfig()

	token, err := GenerateToken(cfg, uuid.New(), "john_doe")
	require.NoError(t, err)

	claims, err := ValidateToken(cfg, token)
	require.NoError(t, err)
	assert.Equal(t, "john_doe", claims.Username)
}

func TestGenerateToken_Issuer(t *testing.T) {
	cfg := testConfig()

	token, err := GenerateToken(cfg, uuid.New(), "testuser")
	require.NoError(t, err)

	claims, err := ValidateToken(cfg, token)
	require.NoError(t, err)
	assert.Equal(t, "social-forum", claims.Issuer)
}

func TestGenerateToken_EmptyUsername(t *testing.T) {
	cfg := testConfig()

	token, err := GenerateToken(cfg, uuid.New(), "")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

// ==================== ValidateToken ====================

func TestValidateToken_ValidToken(t *testing.T) {
	cfg := testConfig()
	uid := uuid.New()

	token, err := GenerateToken(cfg, uid, "testuser")
	require.NoError(t, err)

	claims, err := ValidateToken(cfg, token)
	require.NoError(t, err)
	assert.Equal(t, uid, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	cfg1 := &config.Config{JWTSecret: "secret-a", JWTExpiresIn: 15 * time.Minute}
	cfg2 := &config.Config{JWTSecret: "secret-b", JWTExpiresIn: 15 * time.Minute}

	token, err := GenerateToken(cfg1, uuid.New(), "testuser")
	require.NoError(t, err)

	_, err = ValidateToken(cfg2, token)
	assert.Error(t, err)
}

func TestValidateToken_Malformed(t *testing.T) {
	cfg := testConfig()

	_, err := ValidateToken(cfg, "not.a.token")
	assert.Error(t, err)
}

func TestValidateToken_EmptyString(t *testing.T) {
	cfg := testConfig()

	_, err := ValidateToken(cfg, "")
	assert.Error(t, err)
}

func TestValidateToken_AlgNoneBypass(t *testing.T) {
	cfg := testConfig()

	// Craft token with alg=none
	token := jwt.NewWithClaims(jwt.SigningMethodNone, &Claims{
		UserID:   uuid.New(),
		Username: "hacker",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "hacker",
		},
	})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Should be rejected — unexpected signing method
	_, err = ValidateToken(cfg, tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

func TestValidateToken_Expired(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: -1 * time.Minute, // expired
	}

	token, err := GenerateToken(cfg, uuid.New(), "testuser")
	require.NoError(t, err)

	// Validate with normal config
	cfg2 := testConfig()
	_, err = ValidateToken(cfg2, token)
	assert.Error(t, err)
}
