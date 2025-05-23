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
		sendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountAllGenders(queryStr)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to count gender values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated gender values with search term applied
	genders, err := db.FetchAllGenders(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch gender values", nil)
		return
	}

	// Respond with pagination metadata
	sendSuccess(c, http.StatusOK, gin.H{
		"genders":    genders,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// GetGenderByID handles fetching a single gender value by ID
func GetGenderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	gender, err := db.FetchGenderByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Gender value not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to fetch gender value", nil)
		}
		return
	}

	sendSuccess(c, http.StatusOK, gender)
}

// CreateGender handles creating a new gender value
func CreateGender(c *gin.Context) {
	var gender models.Gender
	if err := c.ShouldBindJSON(&gender); err != nil {
		sendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate required fields
	if gender.Value == "" {
		sendError(c, http.StatusBadRequest, "Gender value is required", nil)
		return
	}

	// Set update timestamp
	gender.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertGender(&gender)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create gender value: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusCreated, gender)
}

// UpdateGender handles updating an existing gender value
func UpdateGender(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	// Fetch the existing gender value first to verify it exists
	_, err = db.FetchGenderByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Gender value not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to fetch existing gender value", nil)
		}
		return
	}

	// Parse request body
	var genderToUpdate models.Gender
	if err := c.ShouldBindJSON(&genderToUpdate); err != nil {
		sendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Update the record
	updatedGender, err := db.UpdateGender(id, &genderToUpdate)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to update gender value: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusOK, updatedGender)
}

// DeleteGender handles soft-deleting a gender value
func DeleteGender(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.DeleteGender(id)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to delete gender value: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusOK, gin.H{"message": "Gender value deleted successfully"})
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
		sendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountDeletedGenders(queryStr)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to count deleted gender values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted gender values with search term applied
	genders, err := db.FetchDeletedGenders(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch deleted gender values", nil)
		return
	}

	// Respond with pagination metadata
	sendSuccess(c, http.StatusOK, gin.H{
		"gender":     genders,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// RestoreGender handles restoring a soft-deleted gender value
func RestoreGender(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.RestoreGender(id)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to restore gender value: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusOK, gin.H{"message": "Gender value restored successfully"})
}
