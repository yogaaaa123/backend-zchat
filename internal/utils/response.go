package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginationMeta struct {
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

type PaginatedResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    interface{}     `json:"data,omitempty"`
	Meta    PaginationMeta  `json:"meta"`
}

func SuccessJSON(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func CreatedJSON(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

func MessageJSON(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success: status >= 200 && status < 300,
		Message: message,
	})
}

func ErrorJSON(c *gin.Context, status int, err string) {
	c.JSON(status, APIResponse{
		Success: false,
		Error:   err,
	})
}

func PaginatedJSON(c *gin.Context, data interface{}, page, limit int, total int64) {
	totalPage := int(total) / limit
	if int(total)%limit != 0 {
		totalPage++
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    data,
		Meta: PaginationMeta{
			Page:      page,
			Limit:     limit,
			Total:     total,
			TotalPage: totalPage,
		},
	})
}
