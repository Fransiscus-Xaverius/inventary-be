package publicHandlers

import (
	"database/sql"
	"net/http"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetAllCategoryColorLabels retrieves all category color labels
func GetAllCategoryColorLabels(c *gin.Context) {
	// Fetch all category color labels from the database
	categoryColorLabels, err := db.FetchAllCategoryColorLabels()
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch category category color labels: "+err.Error(), nil)
		return
	}

	// Return the category color labels
	handlers.SendSuccess(c, http.StatusOK, categoryColorLabels)
}

// GetCategoryColorLabelsByColumn retrieves category color labels for a specific column
func GetCategoryColorLabelsByColumn(c *gin.Context) {
	columnName := c.Param("column")

	// Fetch category color labels for the specified column from the database
	categoryColorLabels, err := db.FetchCategoryColorLabelsByColumn(columnName)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch category category color labels: "+err.Error(), nil)
		return
	}

	// Return the category color labels
	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"column":                columnName,
		"category_color_labels": categoryColorLabels,
	})
}

// GetCategoryColorLabelByColumnAndValue retrieves a specific category color label
func GetCategoryColorLabelByColumnAndValue(c *gin.Context) {
	columnName := c.Param("column")
	categoryValue := c.Param("value")

	// Fetch the specific category color label from the database
	categoryColorLabel, err := db.FetchCategoryColorLabelByColumnAndValue(columnName, categoryValue)
	if err != nil {
		if err == sql.ErrNoRows {
			handlers.SendError(c, http.StatusNotFound, "Category color label not found", nil)
			return
		}
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch category color label: "+err.Error(), nil)
		return
	}

	// Return the category color label
	handlers.SendSuccess(c, http.StatusOK, categoryColorLabel)
}
