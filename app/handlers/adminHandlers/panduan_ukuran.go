package adminHandlers

import (
	"net/http"
	"os"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/app/helpers"
	"github.com/gin-gonic/gin"
)

// UploadPanduanUkuran handles uploading the sizing guide image
func UploadPanduanUkuran(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Image file is required", nil)
		return
	}

	// Save the file to the uploads/panduan directory with a static name
	filePath, err := helpers.SaveUploadedFileWithStaticName(c, file, "uploads/panduan", "1.png", nil)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to save image: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"image_url": filePath})
}

// DeletePanduanUkuran handles deleting the sizing guide image
func DeletePanduanUkuran(c *gin.Context) {
	filePath := "uploads/panduan/1.png"

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		handlers.SendError(c, http.StatusNotFound, "Sizing guide image not found", nil)
		return
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to delete image: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Sizing guide deleted successfully"})
}
