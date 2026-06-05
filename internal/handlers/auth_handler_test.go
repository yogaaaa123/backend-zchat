package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/services"
	"github.com/stretchr/testify/assert"
)

type mockAuthService struct {
	registerFn func(services.RegisterRequest) (*services.AuthResponse, error)
	loginFn    func(services.LoginRequest) (*services.AuthResponse, error)
	meFn       func(uuid.UUID) (*services.UserResponse, error)
	logoutFn   func(uuid.UUID) error
	refreshFn  func(services.RefreshRequest) (*services.AuthResponse, error)
}

func (m *mockAuthService) Register(req services.RegisterRequest) (*services.AuthResponse, error) {
	return m.registerFn(req)
}
func (m *mockAuthService) Login(req services.LoginRequest) (*services.AuthResponse, error) {
	return m.loginFn(req)
}
func (m *mockAuthService) Me(userID uuid.UUID) (*services.UserResponse, error) {
	return m.meFn(userID)
}
func (m *mockAuthService) Logout(userID uuid.UUID) error {
	return m.logoutFn(userID)
}
func (m *mockAuthService) Refresh(req services.RefreshRequest) (*services.AuthResponse, error) {
	return m.refreshFn(req)
}

func TestRegister_Success_201(t *testing.T) {
	mock := &mockAuthService{
		registerFn: func(req services.RegisterRequest) (*services.AuthResponse, error) {
			return &services.AuthResponse{
				Token:        "jwt-token",
				RefreshToken: "refresh-token",
				User: services.UserResponse{
					ID:       uuid.New(),
					Username: "newuser",
					Email:    "new@test.com",
				},
			}, nil
		},
	}
	h := &AuthHandler{authService: mock}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"username":"newuser","email":"new@test.com","password":"password123"}`
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp struct {
		Success bool                    `json:"success"`
		Data    *services.AuthResponse  `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, "newuser", resp.Data.User.Username)
}

func TestRegister_EmailTaken_409(t *testing.T) {
	mock := &mockAuthService{
		registerFn: func(req services.RegisterRequest) (*services.AuthResponse, error) {
			return nil, services.ErrEmailTaken
		},
	}
	h := &AuthHandler{authService: mock}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"username":"newuser","email":"taken@test.com","password":"password123"}`
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegister_InvalidBody_400(t *testing.T) {
	h := &AuthHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"username":""}`
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Success_200(t *testing.T) {
	mock := &mockAuthService{
		loginFn: func(req services.LoginRequest) (*services.AuthResponse, error) {
			return &services.AuthResponse{
				Token:        "jwt-token",
				RefreshToken: "refresh-token",
				User: services.UserResponse{
					ID:       uuid.New(),
					Username: "existing",
					Email:    "existing@test.com",
				},
			}, nil
		},
	}
	h := &AuthHandler{authService: mock}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"email":"existing@test.com","password":"password123"}`
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogin_WrongPassword_401(t *testing.T) {
	mock := &mockAuthService{
		loginFn: func(req services.LoginRequest) (*services.AuthResponse, error) {
			return nil, services.ErrInvalidCredentials
		},
	}
	h := &AuthHandler{authService: mock}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"email":"existing@test.com","password":"wrongpass"}`
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMe_Authenticated_200(t *testing.T) {
	mock := &mockAuthService{
		meFn: func(userID uuid.UUID) (*services.UserResponse, error) {
			return &services.UserResponse{
				ID:       userID,
				Username: "testuser",
				Email:    "test@test.com",
			}, nil
		},
	}
	h := &AuthHandler{authService: mock}

	c, w := authRequest(t, "GET", "/api/v1/auth/me")
	h.Me(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogout_Success_200(t *testing.T) {
	mock := &mockAuthService{
		logoutFn: func(userID uuid.UUID) error {
			return nil
		},
	}
	h := &AuthHandler{authService: mock}

	c, w := authRequest(t, "POST", "/api/v1/auth/logout")
	h.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRefresh_Valid_200(t *testing.T) {
	mock := &mockAuthService{
		refreshFn: func(req services.RefreshRequest) (*services.AuthResponse, error) {
			return &services.AuthResponse{
				Token:        "new-jwt",
				RefreshToken: "new-refresh",
				User: services.UserResponse{ID: uuid.New(), Username: "user", Email: "u@t.com"},
			}, nil
		},
	}
	h := &AuthHandler{authService: mock}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"refresh_token":"valid-refresh-token"}`
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Refresh(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRefresh_InvalidToken_401(t *testing.T) {
	mock := &mockAuthService{
		refreshFn: func(req services.RefreshRequest) (*services.AuthResponse, error) {
			return nil, services.ErrInvalidToken
		},
	}
	h := &AuthHandler{authService: mock}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"refresh_token":"invalid"}`
	c.Request, _ = http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Refresh(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
