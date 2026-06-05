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

type LikeHandler struct {
	likeService *services.LikeService
}

func NewLikeHandler(likeService *services.LikeService) *LikeHandler {
	return &LikeHandler{likeService: likeService}
}

// Toggle godoc
// @Summary      Toggle like
// @Description  Like or unlike a thread (toggle)
// @Tags         Likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Thread ID (UUID)"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Failure      404 {object} utils.APIResponse
// @Router       /api/v1/threads/{id}/like [post]
func (h *LikeHandler) Toggle(c *gin.Context) {
	userID := middleware.GetUserID(c)

	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid thread id")
		return
	}

	result, err := h.likeService.Toggle(userID, threadID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, services.ErrThreadNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.SuccessJSON(c, result)
}
