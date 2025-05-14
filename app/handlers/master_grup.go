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

// GetAllGrups handles fetching all grup values with pagination and search
func GetAllGrups(c *gin.Context) {
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
	totalCount, err := db.CountAllGrups(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count grup values"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated grup values with search term applied
	grups, err := db.FetchAllGrups(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch grup values"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"grup_values": grups,
		"page":        page,
		"total_page":  totalPages,
		"total":       totalCount,
		"sort":        sortColumn,
		"order":       sortDirection,
	})
}

// GetGrupByID handles fetching a single grup value by ID
func GetGrupByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	grup, err := db.FetchGrupByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Grup value not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch grup value"})
		}
		return
	}

	c.JSON(http.StatusOK, grup)
}

// CreateGrup handles creating a new grup value
func CreateGrup(c *gin.Context) {
	var grup models.Grup
	if err := c.ShouldBindJSON(&grup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if grup.Value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Grup value is required"})
		return
	}

	// Set update timestamp
	grup.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertGrup(&grup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create grup value: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, grup)
}

// UpdateGrup handles updating an existing grup value
func UpdateGrup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Fetch the existing grup value first to verify it exists
	_, err = db.FetchGrupByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Grup value not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing grup value"})
		}
		return
	}

	// Parse request body
	var grupToUpdate models.Grup
	if err := c.ShouldBindJSON(&grupToUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the record
	updatedGrup, err := db.UpdateGrup(id, &grupToUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update grup value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedGrup)
}

// DeleteGrup handles soft-deleting a grup value
func DeleteGrup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = db.DeleteGrup(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete grup value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grup value deleted successfully"})
}

// GetDeletedGrups retrieves all soft-deleted grup values with pagination
func GetDeletedGrups(c *gin.Context) {
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
	totalCount, err := db.CountDeletedGrups(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count deleted grup values"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted grup values with search term applied
	grups, err := db.FetchDeletedGrups(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deleted grup values"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"grup_values": grups,
		"page":        page,
		"total_page":  totalPages,
		"total":       totalCount,
		"sort":        sortColumn,
		"order":       sortDirection,
	})
}

// RestoreGrup handles restoring a soft-deleted grup value
func RestoreGrup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = db.RestoreGrup(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore grup value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grup value restored successfully"})
}
