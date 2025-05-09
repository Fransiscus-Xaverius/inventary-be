package handlers

import (
	"net/http"

	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetAllCategoryColorLabels retrieves all category color labels
func GetAllCategoryColorLabels(c *gin.Context) {
	// Fetch all category color labels from the database
	categoryColorLabels, err := db.FetchAllCategoryColorLabels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category category color labels: " + err.Error()})
		return
	}

	// Return the category color labels
	c.JSON(http.StatusOK, gin.H{
		"category_color_labels": categoryColorLabels,
	})
}

// GetCategoryColorLabelsByColumn retrieves category color labels for a specific column
func GetCategoryColorLabelsByColumn(c *gin.Context) {
	columnName := c.Param("column")

	// Fetch category color labels for the specified column from the database
	categoryColorLabels, err := db.FetchCategoryColorLabelsByColumn(columnName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category category color labels: " + err.Error()})
		return
	}

	// Return the category color labels
	c.JSON(http.StatusOK, gin.H{
		"column":                columnName,
		"category_color_labels": categoryColorLabels,
	})
}
