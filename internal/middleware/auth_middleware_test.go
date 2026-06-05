package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/config"
	"github.com/satria/obrolan-api/internal/utils"
	"github.com/stretchr/testify/assert"
)

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:    "test-secret-for-middleware",
		JWTExpiresIn: 15 * time.Minute,
	}
}

func TestAuth_ValidToken(t *testing.T) {
	cfg := testConfig()
	uid := uuid.New()

	token, err := utils.GenerateToken(cfg, uid, "testuser")
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	AuthMiddleware(cfg)(c)

	assert.False(t, c.IsAborted(), "should not be aborted")
	assert.Equal(t, uid, GetUserID(c))
	assert.Equal(t, "testuser", GetUsername(c))
}

func TestAuth_MissingHeader(t *testing.T) {
	cfg := testConfig()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	AuthMiddleware(cfg)(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp utils.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp.Error, "missing authorization header")
}

func TestAuth_WrongFormat(t *testing.T) {
	cfg := testConfig()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer")

	AuthMiddleware(cfg)(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp utils.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp.Error, "invalid authorization format")
}

func TestAuth_NotBearer(t *testing.T) {
	cfg := testConfig()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Basic xyz")

	AuthMiddleware(cfg)(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	cfg := testConfig()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token.here")

	AuthMiddleware(cfg)(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp utils.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp.Error, "invalid or expired token")
}

func TestAuth_ExpiredToken(t *testing.T) {
	expiredCfg := &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: -1 * time.Minute,
	}
	token, err := utils.GenerateToken(expiredCfg, uuid.New(), "testuser")
	assert.NoError(t, err)

	cfg := testConfig()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	AuthMiddleware(cfg)(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_GetUserID_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	uid := GetUserID(c)
	assert.Equal(t, uuid.Nil, uid)
}

func TestAuth_GetUsername_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	username := GetUsername(c)
	assert.Equal(t, "", username)
}
