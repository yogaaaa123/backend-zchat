package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/config"
	"github.com/satria/obrolan-api/internal/utils"
)

var testCfg = &config.Config{
	JWTSecret:    "test-secret-for-e2e",
	JWTExpiresIn: 15 * time.Minute,
}

func testToken(t *testing.T) string {
	t.Helper()
	token, err := utils.GenerateToken(testCfg, uuid.New(), "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return token
}

func authRequest(t *testing.T, method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(method, path, nil)
	token := testToken(t)
	c.Request.Header.Set("Authorization", "Bearer "+token)
	c.Set("userID", uuid.New())
	c.Set("username", "testuser")
	return c, w
}

func publicRequest(t *testing.T, method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(method, path, nil)
	return c, w
}
