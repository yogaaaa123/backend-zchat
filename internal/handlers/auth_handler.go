package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/middleware"
	"github.com/satria/obrolan-api/internal/services"
	"github.com/satria/obrolan-api/internal/utils"
)

// AuthServiceInterface defines the methods needed by AuthHandler
type AuthServiceInterface interface {
	Register(req services.RegisterRequest) (*services.AuthResponse, error)
	Login(req services.LoginRequest) (*services.AuthResponse, error)
	Me(userID uuid.UUID) (*services.UserResponse, error)
	Logout(userID uuid.UUID) error
	Refresh(req services.RefreshRequest) (*services.AuthResponse, error)
}

type AuthHandler struct {
	authService AuthServiceInterface
}

func NewAuthHandler(authService AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary      Register new user
// @Description  Create account with username, email, and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body services.RegisterRequest true "Registration data"
// @Success      201 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      409 {object} utils.APIResponse
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Register(req)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, services.ErrEmailTaken) || errors.Is(err, services.ErrUsernameTaken) {
			status = http.StatusConflict
		}

		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.CreatedJSON(c, result)
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate with email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body services.LoginRequest true "Login credentials"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Login(req)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, services.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}

		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.SuccessJSON(c, result)
}

// Me godoc
// @Summary      Get current user
// @Description  Get profile of the currently authenticated user
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.authService.Me(userID)
	if err != nil {
		utils.ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessJSON(c, user)
}

// Logout godoc
// @Summary      Logout user
// @Description  Revoke all refresh tokens for current user
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if err := h.authService.Logout(userID); err != nil {
		utils.ErrorJSON(c, http.StatusInternalServerError, "failed to logout")
		return
	}

	utils.MessageJSON(c, http.StatusOK, "Logged out successfully")
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Get new access token using refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body services.RefreshRequest true "Refresh token"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req services.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Refresh(req)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, services.ErrInvalidToken) || errors.Is(err, services.ErrTokenReuse) {
			status = http.StatusUnauthorized
		}

		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.SuccessJSON(c, result)
}
