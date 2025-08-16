package adminHandlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetAllNewsletters handles fetching all newsletter entries with pagination and search
func GetAllNewsletters(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "created_at")
	sortDirection := c.DefaultQuery("order", "desc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	page := (offset / limit) + 1

	totalCount, err := db.CountAllNewsletters(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count newsletter entries", nil)
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	newsletters, err := db.FetchAllNewsletters(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch newsletter entries", nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"newsletters": newsletters,
		"page":        page,
		"total_page":  totalPages,
		"total":       totalCount,
		"sort":        sortColumn,
		"order":       sortDirection,
	})
}

// GetNewsletterByID handles fetching a single newsletter entry by ID
func GetNewsletterByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	newsletter, err := db.FetchNewsletterByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Newsletter entry not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch newsletter entry", nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, newsletter)
}

// UpdateNewsletter handles updating an existing newsletter entry
func UpdateNewsletter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var newsletter models.Newsletter
	if err := c.ShouldBindJSON(&newsletter); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedNewsletter, err := db.UpdateNewsletter(id, &newsletter)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to update newsletter entry: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, updatedNewsletter)
}

// DeleteNewsletter handles soft-deleting a newsletter entry
func DeleteNewsletter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.DeleteNewsletter(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Newsletter entry not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to delete newsletter entry", nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Newsletter entry deleted successfully"})
}

// GetDeletedNewsletters retrieves all soft-deleted newsletter entries
func GetDeletedNewsletters(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "deleted_at")
	sortDirection := c.DefaultQuery("order", "desc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		handlers.SendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	page := (offset / limit) + 1

	totalCount, err := db.CountDeletedNewsletters(queryStr)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count deleted newsletter entries", nil)
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	newsletters, err := db.FetchDeletedNewsletters(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch deleted newsletter entries", nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"newsletters": newsletters,
		"page":        page,
		"total_page":  totalPages,
		"total":       totalCount,
		"sort":        sortColumn,
		"order":       sortDirection,
	})
}

// RestoreNewsletter handles restoring a soft-deleted newsletter entry
func RestoreNewsletter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.RestoreNewsletter(id)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Newsletter entry not found or already active", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to restore newsletter entry: "+err.Error(), nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Newsletter entry restored successfully"})
}
