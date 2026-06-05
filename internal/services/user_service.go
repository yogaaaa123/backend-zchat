package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/repositories"
	"github.com/satria/obrolan-api/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrWrongPassword = errors.New("current password is incorrect")
	ErrUserNotFound  = errors.New("user not found")
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type UpdateProfileRequest struct {
	Bio       string `json:"bio" binding:"max=500"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,max=500,url"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=100"`
}

type PublicUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Bio       string    `json:"bio,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt string    `json:"created_at"`
}

func (s *UserService) GetProfile(userID uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	resp := modelToUserResponse(user)
	return &resp, nil
}

func (s *UserService) UpdateProfile(userID uuid.UUID, req UpdateProfileRequest) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	user.Bio = req.Bio
	user.AvatarURL = req.AvatarURL

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	resp := modelToUserResponse(user)
	return &resp, nil
}

func (s *UserService) ChangePassword(userID uuid.UUID, req ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return ErrWrongPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}

func (s *UserService) GetPublicProfile(userID uuid.UUID) (*PublicUserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &PublicUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Bio:       user.Bio,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(utils.TimeFormat),
	}, nil
}
