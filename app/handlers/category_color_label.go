package handlers

import (
	"database/sql"
	"net/http"

	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetAllCategoryColorLabels retrieves all category color labels
func GetAllCategoryColorLabels(c *gin.Context) {
	// Fetch all category color labels from the database
	categoryColorLabels, err := db.FetchAllCategoryColorLabels()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch category category color labels: "+err.Error(), nil)
		return
	}

	// Return the category color labels
	sendSuccess(c, http.StatusOK, categoryColorLabels)
}

// GetCategoryColorLabelsByColumn retrieves category color labels for a specific column
func GetCategoryColorLabelsByColumn(c *gin.Context) {
	columnName := c.Param("column")

	// Fetch category color labels for the specified column from the database
	categoryColorLabels, err := db.FetchCategoryColorLabelsByColumn(columnName)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch category category color labels: "+err.Error(), nil)
		return
	}

	// Return the category color labels
	sendSuccess(c, http.StatusOK, gin.H{
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
			sendError(c, http.StatusNotFound, "Category color label not found", nil)
			return
		}
		sendError(c, http.StatusInternalServerError, "Failed to fetch category color label: "+err.Error(), nil)
		return
	}

	// Return the category color label
	sendSuccess(c, http.StatusOK, categoryColorLabel)
}
