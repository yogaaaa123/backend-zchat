package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_Success(t *testing.T) {
	cleanDB(t)

	req := RegisterRequest{
		Username: "reg_" + uuid.New().String()[:8],
		Email:    "reg_" + uuid.New().String()[:8] + "@test.com",
		Password: "password123",
	}

	resp, err := testAuthSvc.Register(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, uuid.Nil, resp.User.ID)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	cleanDB(t)

	email := "dup_" + uuid.New().String()[:8] + "@test.com"

	_, err := testAuthSvc.Register(RegisterRequest{
		Username: "dup1_" + uuid.New().String()[:8],
		Email:    email,
		Password: "password123",
	})
	require.NoError(t, err)

	_, err = testAuthSvc.Register(RegisterRequest{
		Username: "dup2_" + uuid.New().String()[:8],
		Email:    email,
		Password: "password123",
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrEmailTaken)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	cleanDB(t)

	username := "dup_" + uuid.New().String()[:8]

	_, err := testAuthSvc.Register(RegisterRequest{
		Username: username,
		Email:    "dup_uname1_" + uuid.New().String()[:8] + "@test.com",
		Password: "password123",
	})
	require.NoError(t, err)

	_, err = testAuthSvc.Register(RegisterRequest{
		Username: username,
		Email:    "dup_uname2_" + uuid.New().String()[:8] + "@test.com",
		Password: "password123",
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUsernameTaken)

}

func TestLogin_WrongPassword(t *testing.T) {
	cleanDB(t)

	email := "wrongpw_" + uuid.New().String()[:8] + "@test.com"

	_, err := testAuthSvc.Register(RegisterRequest{
		Username: "wrongpw_" + uuid.New().String()[:8],
		Email:    email,
		Password: "password123",
	})
	require.NoError(t, err)

	_, err = testAuthSvc.Login(LoginRequest{
		Email:    email,
		Password: "wrongpassword",
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_NonExistentEmail(t *testing.T) {
	cleanDB(t)

	_, err := testAuthSvc.Login(LoginRequest{
		Email:    "nonexistent_" + uuid.New().String()[:8] + "@test.com",
		Password: "password123",
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestRefresh_Success(t *testing.T) {
	cleanDB(t)

	regResp, err := testAuthSvc.Register(RegisterRequest{
		Username: "ref_" + uuid.New().String()[:8],
		Email:    "ref_" + uuid.New().String()[:8] + "@test.com",
		Password: "password123",
	})
	require.NoError(t, err)

	resp, err := testAuthSvc.Refresh(RefreshRequest{
		RefreshToken: regResp.RefreshToken,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, regResp.RefreshToken, resp.RefreshToken)
}

func TestRefresh_InvalidToken(t *testing.T) {
	cleanDB(t)

	_, err := testAuthSvc.Refresh(RefreshRequest{
		RefreshToken: "invalid-refresh-token-does-not-exist-12345",
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestLogout_Success(t *testing.T) {
	cleanDB(t)

	regResp, err := testAuthSvc.Register(RegisterRequest{
		Username: "logout_" + uuid.New().String()[:8],
		Email:    "logout_" + uuid.New().String()[:8] + "@test.com",
		Password: "password123",
	})
	require.NoError(t, err)

	err = testAuthSvc.Logout(regResp.User.ID)
	assert.NoError(t, err)

	_, err = testAuthSvc.Refresh(RefreshRequest{
		RefreshToken: regResp.RefreshToken,
	})
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestRegister_EmptyFields(t *testing.T) {
	cleanDB(t)

	// Gin's ShouldBindJSON handles validation (binding tags).
	// Service level: empty fields still get hashed and stored.
	// This test verifies the service doesn't panic with empty data.
	_, err := testAuthSvc.Register(RegisterRequest{
		Username: "",
		Email:    "",
		Password: "",
	})
	// Bcrypt hashes empty string — service doesn't reject it
	// (Gin layer will reject via binding:"required" tags)
	if err != nil {
		assert.Contains(t, err.Error(), "already registered")
	}
}
