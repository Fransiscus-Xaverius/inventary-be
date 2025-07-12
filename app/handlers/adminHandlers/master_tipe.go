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
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountAllTipes(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count tipe values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated tipe values with search term applied
	tipes, err := db.FetchAllTipes(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch tipe values", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"tipes":      tipes,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// GetTipeByID handles fetching a single tipe value by ID
func GetTipeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	tipe, err := db.FetchTipeByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Tipe value not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch tipe value", nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, tipe)
}

// CreateTipe handles creating a new tipe value
func CreateTipe(c *gin.Context) {
	var tipe models.Tipe
	if err := c.ShouldBindJSON(&tipe); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate required fields
	if tipe.Value == "" {
		handlers.SendError(c, http.StatusBadRequest, "Tipe value is required", nil)
		return
	}

	// Set update timestamp
	tipe.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertTipe(&tipe)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to create tipe value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusCreated, tipe)
}

// UpdateTipe handles updating an existing tipe value
func UpdateTipe(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	// Fetch the existing tipe value first to verify it exists
	_, err = db.FetchTipeByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Tipe value not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch existing tipe value", nil)
		}
		return
	}

	// Parse request body
	var tipeToUpdate models.Tipe
	if err := c.ShouldBindJSON(&tipeToUpdate); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Update the record
	updatedTipe, err := db.UpdateTipe(id, &tipeToUpdate)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to update tipe value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, updatedTipe)
}

// DeleteTipe handles soft-deleting a tipe value
func DeleteTipe(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.DeleteTipe(id)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to delete tipe value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Tipe value deleted successfully"})
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
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountDeletedTipes(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count deleted tipe values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted tipe values with search term applied
	tipes, err := db.FetchDeletedTipes(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch deleted tipe values", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"tipes":      tipes,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// RestoreTipe handles restoring a soft-deleted tipe value
func RestoreTipe(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.RestoreTipe(id)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to restore tipe value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Tipe value restored successfully"})
}
