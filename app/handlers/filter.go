package handlers

import (
	"net/http"

	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetFilterOptions returns all unique values for filterable fields
func GetFilterOptions(c *gin.Context) {
	// Fetch all filter options from the database
	filterOptions, err := db.FetchFilterOptions()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch filter options: "+err.Error(), nil)
		return
	}

	// Return the filter options
	sendSuccess(c, http.StatusOK, filterOptions)
}
