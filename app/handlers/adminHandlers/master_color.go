package adminHandlers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetAllColors handles fetching all colors with pagination and search
func GetAllColors(c *gin.Context) {
	// Read query params
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "id")
	sortDirection := c.DefaultQuery("order", "asc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountAllColors(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count colors", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated colors with search term applied
	colors, err := db.FetchAllColors(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch colors", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"colors":     colors,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// GetColorByID handles fetching a single color by ID
func GetColorByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	color, err := db.FetchColorByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Color not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch color", nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, color)
}

// CreateColor handles creating a new color
func CreateColor(c *gin.Context) {
	var color models.Color
	if err := c.ShouldBindJSON(&color); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate required fields
	if color.Nama == "" {
		handlers.SendError(c, http.StatusBadRequest, "Color name is required", nil)
		return
	}

	// Set update timestamp
	color.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertColor(&color)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to create color: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusCreated, color)
}

// UpdateColor handles updating an existing color
func UpdateColor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	// Fetch the existing color first to avoid overwriting with zero values
	existingColor, err := db.FetchColorByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Color not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch existing color", nil)
		}
		return
	}

	// Create a map to hold the raw JSON request body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Create a color struct with the existing data
	colorToUpdate := existingColor

	// Only update fields that were provided in the request
	if nama, ok := requestBody["nama"].(string); ok {
		colorToUpdate.Nama = nama
	}

	if hex, ok := requestBody["hex"].(string); ok {
		colorToUpdate.Hex = hex
	}

	// Update the tanggal_update field to now
	colorToUpdate.TanggalUpdate = time.Now()

	// Perform the update operation
	updatedColor, err := db.UpdateColor(id, &colorToUpdate)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to update color: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, updatedColor)
}

// DeleteColor handles soft-deleting a color
func DeleteColor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.DeleteColor(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Color not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to delete color", nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Color deleted successfully"})
}

// GetDeletedColors retrieves all soft-deleted colors with pagination
func GetDeletedColors(c *gin.Context) {
	// Read query params
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "tanggal_hapus")
	sortDirection := c.DefaultQuery("order", "desc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with filters applied
	totalCount, err := db.CountDeletedColors(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count deleted colors", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted colors with filters applied
	colors, err := db.FetchDeletedColors(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch deleted colors", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"colors":     colors,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// RestoreColor restores a soft-deleted color
func RestoreColor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.RestoreColor(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Color not found or already active", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to restore color: "+err.Error(), nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Color restored successfully"})
}
