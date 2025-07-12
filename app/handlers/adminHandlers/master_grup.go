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
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountAllGrups(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count grup values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated grup values with search term applied
	grups, err := db.FetchAllGrups(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch grup values", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"grups":      grups,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// GetGrupByID handles fetching a single grup value by ID
func GetGrupByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	grup, err := db.FetchGrupByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Grup value not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch grup value", nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, grup)
}

// CreateGrup handles creating a new grup value
func CreateGrup(c *gin.Context) {
	var grup models.Grup
	if err := c.ShouldBindJSON(&grup); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate required fields
	if grup.Value == "" {
		handlers.SendError(c, http.StatusBadRequest, "Grup value is required", nil)
		return
	}

	// Set update timestamp
	grup.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertGrup(&grup)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to create grup value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusCreated, grup)
}

// UpdateGrup handles updating an existing grup value
func UpdateGrup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	// Fetch the existing grup value first to verify it exists
	_, err = db.FetchGrupByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Grup value not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch existing grup value", nil)
		}
		return
	}

	// Parse request body
	var grupToUpdate models.Grup
	if err := c.ShouldBindJSON(&grupToUpdate); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Update the record
	updatedGrup, err := db.UpdateGrup(id, &grupToUpdate)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to update grup value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, updatedGrup)
}

// DeleteGrup handles soft-deleting a grup value
func DeleteGrup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.DeleteGrup(id)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to delete grup value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Grup value deleted successfully"})
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
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountDeletedGrups(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count deleted grup values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted grup values with search term applied
	grups, err := db.FetchDeletedGrups(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch deleted grup values", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"grups":      grups,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// RestoreGrup handles restoring a soft-deleted grup value
func RestoreGrup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.RestoreGrup(id)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to restore grup value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Grup value restored successfully"})
}
