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

// GetAllGenders handles fetching all gender values with pagination and search
func GetAllGenders(c *gin.Context) {
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
	totalCount, err := db.CountAllGenders(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count gender values"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated gender values with search term applied
	genders, err := db.FetchAllGenders(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gender values"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"gender_values": genders,
		"page":          page,
		"total_page":    totalPages,
		"total":         totalCount,
		"sort":          sortColumn,
		"order":         sortDirection,
	})
}

// GetGenderByID handles fetching a single gender value by ID
func GetGenderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	gender, err := db.FetchGenderByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Gender value not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gender value"})
		}
		return
	}

	c.JSON(http.StatusOK, gender)
}

// CreateGender handles creating a new gender value
func CreateGender(c *gin.Context) {
	var gender models.Gender
	if err := c.ShouldBindJSON(&gender); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if gender.Value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gender value is required"})
		return
	}

	// Set update timestamp
	gender.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertGender(&gender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gender value: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gender)
}

// UpdateGender handles updating an existing gender value
func UpdateGender(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Fetch the existing gender value first to verify it exists
	_, err = db.FetchGenderByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Gender value not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing gender value"})
		}
		return
	}

	// Parse request body
	var genderToUpdate models.Gender
	if err := c.ShouldBindJSON(&genderToUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the record
	updatedGender, err := db.UpdateGender(id, &genderToUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update gender value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedGender)
}

// DeleteGender handles soft-deleting a gender value
func DeleteGender(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = db.DeleteGender(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete gender value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gender value deleted successfully"})
}

// GetDeletedGenders retrieves all soft-deleted gender values with pagination
func GetDeletedGenders(c *gin.Context) {
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
	totalCount, err := db.CountDeletedGenders(queryStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count deleted gender values"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted gender values with search term applied
	genders, err := db.FetchDeletedGenders(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deleted gender values"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"gender_values": genders,
		"page":          page,
		"total_page":    totalPages,
		"total":         totalCount,
		"sort":          sortColumn,
		"order":         sortDirection,
	})
}

// RestoreGender handles restoring a soft-deleted gender value
func RestoreGender(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = db.RestoreGender(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore gender value: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gender value restored successfully"})
}
