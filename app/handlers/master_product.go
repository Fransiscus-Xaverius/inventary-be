package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/app/validation/master_product"
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
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Perform validation using the new validation package
	if validationErr := master_product.ValidateCreate(&product); validationErr != nil {
		c.JSON(http.StatusBadRequest, validationErr)
		return
	}

	product.TanggalUpdate = time.Now()
	err := db.InsertProduct(&product)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.JSON(http.StatusBadRequest, gin.H{
				"column":  "artikel",
				"message": "Product with this artikel already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func UpdateProduct(c *gin.Context) {
	artikel := c.Param("artikel")

	// Fetch the existing product first to avoid overwriting with zero values
	existingProduct, err := db.FetchProductByArtikel(artikel)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing product"})
		}
		return
	}

	// Create a map to hold the raw JSON request body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a product struct with the existing data
	productToUpdate := existingProduct

	// Only update fields that were provided in the request
	if warna, ok := requestBody["warna"].(string); ok {
		productToUpdate.Warna = warna
	}

	if size, ok := requestBody["size"].(string); ok {
		productToUpdate.Size = size
	}

	if grup, ok := requestBody["grup"].(string); ok {
		productToUpdate.Grup = grup
	}

	if unit, ok := requestBody["unit"].(string); ok {
		productToUpdate.Unit = unit
	}

	if kat, ok := requestBody["kat"].(string); ok {
		productToUpdate.Kat = kat
	}

	if model, ok := requestBody["model"].(string); ok {
		productToUpdate.Model = model
	}

	if gender, ok := requestBody["gender"].(string); ok {
		productToUpdate.Gender = gender
	}

	if tipe, ok := requestBody["tipe"].(string); ok {
		productToUpdate.Tipe = tipe
	}

	if status, ok := requestBody["status"].(string); ok {
		productToUpdate.Status = status
	}

	if supplier, ok := requestBody["supplier"].(string); ok {
		productToUpdate.Supplier = supplier
	}

	if diupdateOleh, ok := requestBody["diupdate_oleh"].(string); ok {
		productToUpdate.DiupdateOleh = diupdateOleh
	}

	// Handle numeric fields
	if harga, ok := requestBody["harga"].(float64); ok {
		productToUpdate.Harga = harga
	}

	// Handle date fields if they're provided
	if tanggalProduk, ok := requestBody["tanggal_produk"].(string); ok && tanggalProduk != "" {
		date, err := time.Parse("2006-01-02", tanggalProduk)
		if err == nil {
			productToUpdate.TanggalProduk = date
		}
	}

	if tanggalTerima, ok := requestBody["tanggal_terima"].(string); ok && tanggalTerima != "" {
		date, err := time.Parse("2006-01-02", tanggalTerima)
		if err == nil {
			productToUpdate.TanggalTerima = date
		}
	}

	// Update the tanggal_update field to now
	productToUpdate.TanggalUpdate = time.Now()

	// Validate the updated product
	if validationErr := master_product.ValidateUpdate(&productToUpdate); validationErr != nil {
		c.JSON(http.StatusBadRequest, validationErr)
		return
	}

	// Perform the update operation
	updatedProduct, err := db.UpdateProduct(artikel, &productToUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product: " + err.Error()})
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
