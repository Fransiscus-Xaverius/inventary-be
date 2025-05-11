package handlers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/everysoft/inventary-be/app/master_size"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetAllSizes handles fetching all sizes with pagination and search
func GetAllSizes(c *gin.Context) {
	// Read query params
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "id")
	sortDirection := c.DefaultQuery("order", "asc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountAllSizes(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count sizes"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated sizes with search term applied
	sizes, err := db.GetAllSizes(page, limit, queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sizes"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"sizes":      sizes,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// GetSizeByID handles fetching a single size by ID
func GetSizeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	size, err := db.GetSizeByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Size not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch size"})
		}
		return
	}

	c.JSON(http.StatusOK, size)
}

// CreateSize handles creating a new size
func CreateSize(c *gin.Context) {
	var size master_size.Size
	if err := c.ShouldBindJSON(&size); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if size.Value == "" || size.Unit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Value and unit are required"})
		return
	}

	// Set update timestamp
	size.TanggalUpdate = time.Now()

	// Insert to database
	err := db.CreateSize(&size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create size: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, size)
}

// UpdateSize handles updating an existing size
func UpdateSize(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Fetch the existing size first to avoid overwriting with zero values
	existingSize, err := db.GetSizeByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Size not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing size"})
		}
		return
	}

	// Create a map to hold the raw JSON request body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a size struct with the existing data
	sizeToUpdate := existingSize

	// Only update fields that were provided in the request
	if value, ok := requestBody["value"].(string); ok {
		sizeToUpdate.Value = value
	}

	if unit, ok := requestBody["unit"].(string); ok {
		sizeToUpdate.Unit = unit
	}

	// Update the tanggal_update field to now
	sizeToUpdate.TanggalUpdate = time.Now()

	// Perform the update operation
	if err := db.UpdateSize(sizeToUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update size: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, sizeToUpdate)
}

// DeleteSize handles soft-deleting a size
func DeleteSize(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = db.DeleteSize(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Size not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete size"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Size deleted successfully"})
}

// GetDeletedSizes retrieves all soft-deleted sizes with pagination
func GetDeletedSizes(c *gin.Context) {
	// Read query params
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "tanggal_hapus")
	sortDirection := c.DefaultQuery("order", "desc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountDeletedSizes(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count deleted sizes"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted sizes with search term applied
	sizes, err := db.GetDeletedSizes(page, limit, queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deleted sizes"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"sizes":      sizes,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// RestoreSize restores a soft-deleted size
func RestoreSize(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// First check if the size exists and is deleted
	size, err := db.GetSizeByIDIncludeDeleted(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Size not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch size"})
		}
		return
	}

	// Check if size is already active (not deleted)
	if size.TanggalHapus == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Size is already active and not deleted"})
		return
	}

	// Restore the size
	err = db.RestoreSize(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore size: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Size restored successfully"})
}
