package handlers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetAllTipes handles fetching all tipe values with pagination and search
func GetAllTipes(c *gin.Context) {
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
	totalCount, err := db.CountAllTipes(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count tipe values"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated tipe values with search term applied
	tipes, err := db.FetchAllTipes(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tipe values"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"tipe_values": tipes,
		"page":        page,
		"total_page":  totalPages,
		"total":       totalCount,
		"sort":        sortColumn,
		"order":       sortDirection,
	})
}

// GetTipeByID handles fetching a single tipe value by ID
func GetTipeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	tipe, err := db.FetchTipeByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tipe value not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tipe value"})
		}
		return
	}

	c.JSON(http.StatusOK, tipe)
}

// CreateTipe handles creating a new tipe value
func CreateTipe(c *gin.Context) {
	var tipe models.Tipe
	if err := c.ShouldBindJSON(&tipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if tipe.Value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipe value is required"})
		return
	}

	// Set update timestamp
	tipe.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertTipe(&tipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tipe value: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tipe)
}

// UpdateTipe handles updating an existing tipe value
func UpdateTipe(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Fetch the existing tipe value first to verify it exists
	_, err = db.FetchTipeByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tipe value not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing tipe value"})
		}
		return
	}

	// Parse request body
	var tipeToUpdate models.Tipe
	if err := c.ShouldBindJSON(&tipeToUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the record
	updatedTipe, err := db.UpdateTipe(id, &tipeToUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tipe value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedTipe)
}

// DeleteTipe handles soft-deleting a tipe value
func DeleteTipe(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = db.DeleteTipe(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tipe value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipe value deleted successfully"})
}

// GetDeletedTipes retrieves all soft-deleted tipe values with pagination
func GetDeletedTipes(c *gin.Context) {
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
	totalCount, err := db.CountDeletedTipes(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count deleted tipe values"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted tipe values with search term applied
	tipes, err := db.FetchDeletedTipes(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deleted tipe values"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"tipe_values": tipes,
		"page":        page,
		"total_page":  totalPages,
		"total":       totalCount,
		"sort":        sortColumn,
		"order":       sortDirection,
	})
}

// RestoreTipe handles restoring a soft-deleted tipe value
func RestoreTipe(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = db.RestoreTipe(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore tipe value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipe value restored successfully"})
}
