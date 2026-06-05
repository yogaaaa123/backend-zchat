package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/middleware"
	"github.com/satria/obrolan-api/internal/services"
	"github.com/satria/obrolan-api/internal/utils"
)

type CommentHandler struct {
	commentService *services.CommentService
}

func NewCommentHandler(commentService *services.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// Create godoc
// @Summary      Create comment
// @Description  Create comment or nested reply on a thread
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Thread ID (UUID)"
// @Param        request body services.CreateCommentRequest true "Comment data"
// @Success      201 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Failure      404 {object} utils.APIResponse
// @Router       /api/v1/threads/{id}/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid thread id")
		return
	}

	var req services.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.commentService.Create(userID, threadID, req)
	if err != nil {
		status := http.StatusInternalServerError

		switch {
		case errors.Is(err, services.ErrThreadNotFound):
			status = http.StatusNotFound
		case errors.Is(err, services.ErrParentNotFound), errors.Is(err, services.ErrInvalidParent), errors.Is(err, services.ErrInvalidParentID):
			status = http.StatusBadRequest
		}

		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.CreatedJSON(c, result)
}

// GetByThread godoc
// @Summary      Get comments
// @Description  Get all comments for a thread (with nested replies)
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Thread ID (UUID)"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Failure      404 {object} utils.APIResponse
// @Router       /api/v1/threads/{id}/comments [get]
func (h *CommentHandler) GetByThread(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid thread id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	comments, total, err := h.commentService.GetByThread(threadID, page, limit)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, services.ErrThreadNotFound) {
			status = http.StatusNotFound
		}
		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.PaginatedJSON(c, comments, page, limit, total)
}

// Delete godoc
// @Summary      Delete comment
// @Description  Delete comment (owner only)
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Comment ID (UUID)"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Failure      403 {object} utils.APIResponse
// @Failure      404 {object} utils.APIResponse
// @Router       /api/v1/comments/{id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)

	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid comment id")
		return
	}

	if err := h.commentService.Delete(userID, commentID); err != nil {
		status := http.StatusInternalServerError

		switch {
		case errors.Is(err, services.ErrCommentNotFound):
			status = http.StatusNotFound
		case errors.Is(err, services.ErrForbidden):
			status = http.StatusForbidden
		}

		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.MessageJSON(c, http.StatusOK, "Comment deleted")
}
