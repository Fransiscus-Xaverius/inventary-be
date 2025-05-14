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

// GetAllKats handles fetching all category values with pagination and search
func GetAllKats(c *gin.Context) {
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
	totalCount, err := db.CountAllKats(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count category values"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated category values with search term applied
	kats, err := db.FetchAllKats(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category values"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"kat_values": kats,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// GetKatByID handles fetching a single category value by ID
func GetKatByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	kat, err := db.FetchKatByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category value not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category value"})
		}
		return
	}

	c.JSON(http.StatusOK, kat)
}

// CreateKat handles creating a new category value
func CreateKat(c *gin.Context) {
	var kat models.Kat
	if err := c.ShouldBindJSON(&kat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if kat.Value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category value is required"})
		return
	}

	// Set update timestamp
	kat.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertKat(&kat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category value: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, kat)
}

// UpdateKat handles updating an existing category value
func UpdateKat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Fetch the existing category value first to verify it exists
	_, err = db.FetchKatByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category value not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing category value"})
		}
		return
	}

	// Parse request body
	var katToUpdate models.Kat
	if err := c.ShouldBindJSON(&katToUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the record
	updatedKat, err := db.UpdateKat(id, &katToUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedKat)
}

// DeleteKat handles soft-deleting a category value
func DeleteKat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = db.DeleteKat(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category value deleted successfully"})
}

// GetDeletedKats retrieves all soft-deleted category values with pagination
func GetDeletedKats(c *gin.Context) {
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
	totalCount, err := db.CountDeletedKats(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count deleted category values"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted category values with search term applied
	kats, err := db.FetchDeletedKats(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deleted category values"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"kat_values": kats,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// RestoreKat handles restoring a soft-deleted category value
func RestoreKat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = db.RestoreKat(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore category value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category value restored successfully"})
}
