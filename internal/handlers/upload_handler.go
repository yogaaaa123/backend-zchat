package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/satria/obrolan-api/internal/services"
	"github.com/satria/obrolan-api/internal/utils"
)

type UploadHandler struct {
	uploadService services.Uploader
}

func NewUploadHandler(uploadService services.Uploader) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// Upload godoc
// @Summary      Upload image
// @Description  Upload image to Cloudinary, returns URL
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Image file (jpeg, png, gif, webp, max 5MB)"
// @Success      200 {object} utils.APIResponse
// @Failure      400 {object} utils.APIResponse
// @Failure      401 {object} utils.APIResponse
// @Router       /api/v1/upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorJSON(c, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	url, err := h.uploadService.UploadImage(file, header)
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()

		switch msg {
		case "file too large: max 5MB", "invalid file type: allowed jpeg, png, gif, webp", "invalid file extension":
			status = http.StatusBadRequest
		}

		utils.ErrorJSON(c, status, msg)
		return
	}

	utils.SuccessJSON(c, gin.H{"url": url})
}
