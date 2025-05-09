package handlers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	Product "github.com/everysoft/inventary-be/app/master_product"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

func GetAllProducts(c *gin.Context) {
	// Read query params
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "no")
	sortDirection := c.DefaultQuery("order", "asc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Extract filter parameters
	filters := make(map[string]string)
	validFilterFields := []string{"warna", "size", "grup", "unit", "kat", "model", "gender", "tipe", "status", "supplier"}

	for _, field := range validFilterFields {
		if value := c.Query(field); value != "" {
			filters[field] = value
		}
	}

	// Fetch total count with filters applied
	totalCount, err := db.CountAllProducts(queryStr, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated products with filters applied
	products, err := db.FetchAllProducts(limit, offset, queryStr, filters, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"products":   products,
		"page":       page,
		"total_page": totalPages,
		"filters":    filters,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

func GetProductByArtikel(c *gin.Context) {
	artikel := c.Param("artikel")
	product, err := db.FetchProductByArtikel(artikel)

	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		}
		return
	}

	c.JSON(http.StatusOK, product)
}

func CreateProduct(c *gin.Context) {
	var product Product.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.TanggalUpdate = time.Now()
	err := db.InsertProduct(&product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func UpdateProduct(c *gin.Context) {
	artikel := c.Param("artikel")
	var product Product.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.TanggalUpdate = time.Now()
	updatedProduct, err := db.UpdateProduct(artikel, &product)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedProduct)
}

func DeleteProduct(c *gin.Context) {
	artikel := c.Param("artikel")
	err := db.DeleteProduct(artikel)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// GetFilterOptions returns all unique values for filterable fields
func GetFilterOptions(c *gin.Context) {
	// Fetch all filter options from the database
	filterOptions, err := db.FetchFilterOptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch filter options: " + err.Error()})
		return
	}

	// Return the filter options
	c.JSON(http.StatusOK, gin.H{
		"fields": filterOptions,
	})
}

// GetDeletedProducts retrieves all soft-deleted products with pagination
func GetDeletedProducts(c *gin.Context) {
	// Read query params
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "tanggal_hapus")
	sortDirection := c.DefaultQuery("order", "desc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Extract filter parameters
	filters := make(map[string]string)
	validFilterFields := []string{"warna", "size", "grup", "unit", "kat", "model", "gender", "tipe", "status", "supplier"}

	for _, field := range validFilterFields {
		if value := c.Query(field); value != "" {
			filters[field] = value
		}
	}

	// Fetch total count with filters applied
	totalCount, err := db.CountDeletedProducts(queryStr, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count deleted products"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted products with filters applied
	products, err := db.FetchDeletedProducts(limit, offset, queryStr, filters, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deleted products"})
		return
	}

	// Respond with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"products":   products,
		"page":       page,
		"total_page": totalPages,
		"filters":    filters,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// RestoreProduct restores a soft-deleted product
func RestoreProduct(c *gin.Context) {
	artikel := c.Param("artikel")

	// First check if the product exists and is deleted
	product, err := db.FetchProductByArtikelIncludeDeleted(artikel)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		}
		return
	}

	// Check if product is already active (not deleted)
	if product.TanggalHapus == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product is already active and not deleted"})
		return
	}

	// Restore the product
	err = db.RestoreProduct(artikel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore product: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product restored successfully"})
}
