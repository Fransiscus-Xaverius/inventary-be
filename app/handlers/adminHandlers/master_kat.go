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
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountAllKats(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count category values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated category values with search term applied
	kats, err := db.FetchAllKats(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch category values", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"kats":       kats,
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
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	kat, err := db.FetchKatByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Category value not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch category value", nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, kat)
}

// CreateKat handles creating a new category value
func CreateKat(c *gin.Context) {
	var kat models.Kat
	if err := c.ShouldBindJSON(&kat); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate required fields
	if kat.Value == "" {
		handlers.SendError(c, http.StatusBadRequest, "Category value is required", nil)
		return
	}

	// Set update timestamp
	kat.TanggalUpdate = time.Now()

	// Insert to database
	err := db.InsertKat(&kat)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to create category value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusCreated, kat)
}

// UpdateKat handles updating an existing category value
func UpdateKat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	// Fetch the existing category value first to verify it exists
	_, err = db.FetchKatByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Category value not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch existing category value", nil)
		}
		return
	}

	// Parse request body
	var katToUpdate models.Kat
	if err := c.ShouldBindJSON(&katToUpdate); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Update the record
	updatedKat, err := db.UpdateKat(id, &katToUpdate)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to update category value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, updatedKat)
}

// DeleteKat handles soft-deleting a category value
func DeleteKat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.DeleteKat(id)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to delete category value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Category value deleted successfully"})
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
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountDeletedKats(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count deleted category values", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted category values with search term applied
	kats, err := db.FetchDeletedKats(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch deleted category values", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"kats":       kats,
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
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.RestoreKat(id)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to restore category value: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Category value restored successfully"})
}
