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

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile godoc
// @Summary      Get own profile
// @Description  Get current user's profile
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		utils.ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessJSON(c, profile)
}

// UpdateProfile godoc
// @Summary      Update profile
// @Description  Update bio and avatar URL
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body services.UpdateProfileRequest true "Profile data"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	profile, err := h.userService.UpdateProfile(userID, req)
	if err != nil {
		utils.ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessJSON(c, profile)
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Change current user's password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body services.ChangePasswordRequest true "Password data"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.userService.ChangePassword(userID, req); err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, services.ErrWrongPassword) {
			status = http.StatusUnauthorized
		}

		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.MessageJSON(c, http.StatusOK, "Password updated")
}

// GetPublicProfile godoc
// @Summary      Get user by ID
// @Description  Get public profile of any user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID (UUID)"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      404 {object} utils.APIResponse
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetPublicProfile(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid user id")
		return
	}

	profile, err := h.userService.GetPublicProfile(userID)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, services.ErrUserNotFound) {
			status = http.StatusNotFound
		}

		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.SuccessJSON(c, profile)
}
