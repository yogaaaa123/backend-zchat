package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/config"
	"github.com/satria/obrolan-api/internal/models"
	"github.com/satria/obrolan-api/internal/repositories"
	"github.com/satria/obrolan-api/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired refresh token")
	ErrTokenReuse         = errors.New("token reuse detected, all sessions revoked")
)

type AuthService struct {
	userRepo        *repositories.UserRepository
	refreshTokenRepo *repositories.RefreshTokenRepository
	cfg             *config.Config
}

func NewAuthService(userRepo *repositories.UserRepository, refreshTokenRepo *repositories.RefreshTokenRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, refreshTokenRepo: refreshTokenRepo, cfg: cfg}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Bio       string    `json:"bio,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt string    `json:"created_at"`
}

func modelToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Bio:       user.Bio,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(utils.TimeFormat),
	}
}

func (s *AuthService) createRefreshToken(userID uuid.UUID) (string, error) {
	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	rawToken := hex.EncodeToString(bytes)

	// Hash for storage
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Save to DB
	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenExpiresIn),
	}

	if err := s.refreshTokenRepo.Create(refreshToken); err != nil {
		return "", err
	}

	return rawToken, nil
}

func (s *AuthService) issueTokens(user *models.User) (*AuthResponse, error) {
	// Generate JWT
	token, err := utils.GenerateToken(s.cfg, user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// Delete old refresh tokens for this user
	s.refreshTokenRepo.DeleteByUserID(user.ID)

	// Generate new refresh token
	refreshToken, err := s.createRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         modelToUserResponse(user),
	}, nil
}

func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check email uniqueness
	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	// Check username uniqueness
	existing, err = s.userRepo.FindByUsername(req.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return s.issueTokens(user)
}

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(user)
}

func (s *AuthService) Me(userID uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	resp := modelToUserResponse(user)
	return &resp, nil
}

func (s *AuthService) Logout(userID uuid.UUID) error {
	return s.refreshTokenRepo.DeleteByUserID(userID)
}

func (s *AuthService) Refresh(req RefreshRequest) (*AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	hash := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Atomic: find and delete in one transaction with row lock
	storedToken, err := s.refreshTokenRepo.FindByTokenHashAndDelete(tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			expired, err2 := s.refreshTokenRepo.FindExpiredByHash(tokenHash)
			if err2 == nil && expired != nil {
				s.refreshTokenRepo.DeleteByUserID(expired.UserID)
				return nil, ErrTokenReuse
			}
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	user, err := s.userRepo.FindByID(storedToken.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return s.issueTokens(user)
}
