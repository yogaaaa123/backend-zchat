package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_GetProfile_Success(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)

	uid, _ := uuid.Parse(userID)
	resp, err := testUserSvc.GetProfile(uid)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, userID, resp.ID.String())
	assert.NotEmpty(t, resp.Username)
	assert.NotEmpty(t, resp.Email)
}

func TestUser_GetProfile_NotFound(t *testing.T) {
	cleanDB(t)

	_, err := testUserSvc.GetProfile(uuid.New())

	assert.Error(t, err)
}

func TestUser_UpdateProfile_Success(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)

	uid, _ := uuid.Parse(userID)
	req := UpdateProfileRequest{
		Bio:       "Hello, this is my bio!",
		AvatarURL: "https://example.com/avatar.jpg",
	}

	resp, err := testUserSvc.UpdateProfile(uid, req)

	require.NoError(t, err)
	assert.Equal(t, "Hello, this is my bio!", resp.Bio)
	assert.Equal(t, "https://example.com/avatar.jpg", resp.AvatarURL)
}

func TestUser_ChangePassword_Success(t *testing.T) {
	cleanDB(t)

	regResp, err := testAuthSvc.Register(RegisterRequest{
		Username: "chgpw_" + uuid.New().String()[:8],
		Email:    "chgpw_" + uuid.New().String()[:8] + "@test.com",
		Password: "oldpassword123",
	})
	require.NoError(t, err)

	req := ChangePasswordRequest{
		OldPassword: "oldpassword123",
		NewPassword: "newpassword456",
	}

	err = testUserSvc.ChangePassword(regResp.User.ID, req)
	assert.NoError(t, err)

	// Verifikasi bisa login pake password baru
	_, err = testAuthSvc.Login(LoginRequest{
		Email:    regResp.User.Email,
		Password: "newpassword456",
	})
	assert.NoError(t, err)
}

func TestUser_ChangePassword_WrongOldPassword(t *testing.T) {
	cleanDB(t)

	regResp, err := testAuthSvc.Register(RegisterRequest{
		Username: "wrongpw_" + uuid.New().String()[:8],
		Email:    "wrongpw_" + uuid.New().String()[:8] + "@test.com",
		Password: "oldpassword123",
	})
	require.NoError(t, err)

	req := ChangePasswordRequest{
		OldPassword: "salahpassword",
		NewPassword: "newpassword456",
	}

	err = testUserSvc.ChangePassword(regResp.User.ID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "current password is incorrect")
}

func TestUser_GetPublicProfile_Success(t *testing.T) {
	cleanDB(t)
	userID := seedUser(t)

	uid, _ := uuid.Parse(userID)
	resp, err := testUserSvc.GetPublicProfile(uid)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, userID, resp.ID.String())
	assert.NotEmpty(t, resp.Username)
}

func TestUser_GetPublicProfile_NotFound(t *testing.T) {
	cleanDB(t)

	_, err := testUserSvc.GetPublicProfile(uuid.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}
