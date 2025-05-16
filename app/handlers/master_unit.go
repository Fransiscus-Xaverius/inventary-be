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

// GetAllUnits handles fetching all unit values with pagination and search
func GetAllUnits(c *gin.Context) {
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
	totalCount, err := db.CountAllUnits(queryStr)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to count unit values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated unit values with search term applied
	units, err := db.FetchAllUnits(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch unit values", nil)
		return
	}

	// Respond with pagination metadata
	sendSuccess(c, http.StatusOK, gin.H{
		"units":      units,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// GetUnitByID handles fetching a single unit value by ID
func GetUnitByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	unit, err := db.FetchUnitByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Unit value not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to fetch unit value", nil)
		}
		return
	}

	sendSuccess(c, http.StatusOK, unit)
}

// CreateUnit handles creating a new unit value
func CreateUnit(c *gin.Context) {
	var unit models.Unit
	if err := c.ShouldBindJSON(&unit); err != nil {
		sendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate required fields
	if unit.Value == "" {
		sendError(c, http.StatusBadRequest, "Unit value is required", nil)
		return
	}

	// Set update timestamp
	unit.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertUnit(&unit)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create unit value: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusCreated, unit)
}

// UpdateUnit handles updating an existing unit value
func UpdateUnit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	// Fetch the existing unit value first to verify it exists
	_, err = db.FetchUnitByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Unit value not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to fetch existing unit value", nil)
		}
		return
	}

	// Parse request body
	var unitToUpdate models.Unit
	if err := c.ShouldBindJSON(&unitToUpdate); err != nil {
		sendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Update the record
	updatedUnit, err := db.UpdateUnit(id, &unitToUpdate)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to update unit value: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusOK, updatedUnit)
}

// DeleteUnit handles soft-deleting a unit value
func DeleteUnit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.DeleteUnit(id)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to delete unit value: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusOK, gin.H{"message": "Unit value deleted successfully"})
}

// GetDeletedUnits retrieves all soft-deleted unit values with pagination
func GetDeletedUnits(c *gin.Context) {
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
	totalCount, err := db.CountDeletedUnits(queryStr)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to count deleted unit values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted unit values with search term applied
	units, err := db.FetchDeletedUnits(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch deleted unit values", nil)
		return
	}

	// Respond with pagination metadata
	sendSuccess(c, http.StatusOK, gin.H{
		"units":      units,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// RestoreUnit handles restoring a soft-deleted unit value
func RestoreUnit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.RestoreUnit(id)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to restore unit value: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusOK, gin.H{"message": "Unit value restored successfully"})
}
