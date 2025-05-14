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

// GetCategoryColorLabelByColumnAndValue retrieves a specific category color label
func GetCategoryColorLabelByColumnAndValue(c *gin.Context) {
	columnName := c.Param("column")
	categoryValue := c.Param("value")

	// Fetch the specific category color label from the database
	categoryColorLabel, err := db.FetchCategoryColorLabelByColumnAndValue(columnName, categoryValue)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category color label not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category color label: " + err.Error()})
		return
	}

	// Return the category color label
	c.JSON(http.StatusOK, gin.H{
		"nama_kolom":     categoryColorLabel.NamaKolom,
		"keterangan":     categoryColorLabel.Keterangan,
		"kode_warna":     categoryColorLabel.KodeWarna,
		"nama_warna":     categoryColorLabel.NamaWarna,
		"tanggal_update": categoryColorLabel.TanggalUpdate,
	})
}
