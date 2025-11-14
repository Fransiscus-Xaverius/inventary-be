package publicHandlers

import (
	"net/http"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetFilterOptions returns all unique values for filterable fields
func GetFilterOptions(c *gin.Context) {
	// Fetch all filter options from the database
	filterOptions, err := db.FetchFilterOptions()
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch filter options: "+err.Error(), nil)
		return
	}

	// Return the filter options
	handlers.SendSuccess(c, http.StatusOK, filterOptions)
}
