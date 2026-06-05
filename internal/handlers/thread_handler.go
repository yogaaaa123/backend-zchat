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

type ThreadHandler struct {
	threadService *services.ThreadService
}

func NewThreadHandler(threadService *services.ThreadService) *ThreadHandler {
	return &ThreadHandler{threadService: threadService}
}

// Create godoc
// @Summary      Create thread
// @Description  Create a new forum thread
// @Tags         Threads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body services.CreateThreadRequest true "Thread data"
// @Success      201 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/threads [post]
func (h *ThreadHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req services.CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.threadService.Create(userID, req)
	if err != nil {
		utils.ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.CreatedJSON(c, result)
}

// GetAll godoc
// @Summary      List threads
// @Description  Get paginated list of threads
// @Tags         Threads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Success      200 {object} utils.PaginatedResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/threads [get]
func (h *ThreadHandler) GetAll(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	threads, total, err := h.threadService.GetAll(userID, page, limit)
	if err != nil {
		utils.ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.PaginatedJSON(c, threads, page, limit, total)
}

// GetByID godoc
// @Summary      Get thread by ID
// @Description  Get detailed thread with user info
// @Tags         Threads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Thread ID (UUID)"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      404 {object} utils.APIResponse
// @Router       /api/v1/threads/{id} [get]
func (h *ThreadHandler) GetByID(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid thread id")
		return
	}

	result, err := h.threadService.GetByID(userID, id)
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

// Update godoc
// @Summary      Update thread
// @Description  Update thread (owner only)
// @Tags         Threads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Thread ID (UUID)"
// @Param        request body services.UpdateThreadRequest true "Thread data"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Failure      403 {object} utils.APIResponse
// @Failure      404 {object} utils.APIResponse
// @Router       /api/v1/threads/{id} [put]
func (h *ThreadHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid thread id")
		return
	}

	var req services.UpdateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.threadService.Update(userID, id, req)
	if err != nil {
		status := http.StatusInternalServerError

		switch {
		case errors.Is(err, services.ErrThreadNotFound):
			status = http.StatusNotFound
		case errors.Is(err, services.ErrForbidden):
			status = http.StatusForbidden
		}

		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.SuccessJSON(c, result)
}

// Search godoc
// @Summary      Search threads
// @Description  Full-text search by title and content
// @Tags         Threads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        q query string true "Search keyword"
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Success      200 {object} utils.PaginatedResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/threads/search [get]
func (h *ThreadHandler) Search(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		utils.ErrorJSON(c, http.StatusBadRequest, "search query is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	userID := middleware.GetUserID(c)

	threads, total, err := h.threadService.Search(userID, keyword, page, limit)
	if err != nil {
		utils.ErrorJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.PaginatedJSON(c, threads, page, limit, total)
}

// Delete godoc
// @Summary      Delete thread
// @Description  Delete thread (owner only)
// @Tags         Threads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Thread ID (UUID)"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Failure      403 {object} utils.APIResponse
// @Failure      404 {object} utils.APIResponse
// @Router       /api/v1/threads/{id} [delete]
func (h *ThreadHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "invalid thread id")
		return
	}

	if err := h.threadService.Delete(userID, id); err != nil {
		status := http.StatusInternalServerError

		switch {
		case errors.Is(err, services.ErrThreadNotFound):
			status = http.StatusNotFound
		case errors.Is(err, services.ErrForbidden):
			status = http.StatusForbidden
		}

		utils.ErrorJSON(c, status, err.Error())
		return
	}

	utils.MessageJSON(c, http.StatusOK, "Thread deleted")
}
