package adminHandlers

import (
	"encoding/json"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/app/helpers"
	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/app/validation/master_product"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetAllProducts handles retrieving all products with pagination and filtering
func GetAllProducts(c *gin.Context) {
	// Parse pagination parameters
	limit, offset, page, err := helpers.ParsePaginationParams(c)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Get query params
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "no")
	sortDirection := c.DefaultQuery("order", "asc")

	// Extract filter parameters
	filters := helpers.ExtractFilters(c)

	// Fetch total count with filters applied
	totalCount, err := db.CountAllProducts(queryStr, filters)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count products", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated products with filters applied
	products, err := db.FetchAllProducts(limit, offset, queryStr, filters, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch products", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, handlers.PaginatedData{
		Items:      products,
		Page:       page,
		TotalPages: totalPages,
		TotalItems: totalCount,
		Filters:    filters,
		Sort:       sortColumn,
		Order:      sortDirection,
	})
}

// GetProductByArtikel handles retrieving a single product by its artikel
func GetProductByArtikel(c *gin.Context) {
	artikel := c.Param("artikel")
	product, err := db.FetchProductByArtikel(artikel)

	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Product not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch product", nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, product)
}

// CreateProduct handles creating a new product
func CreateProduct(c *gin.Context) {
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate marketplace keys
	if marketplace, ok := requestBody["marketplace"].(map[string]interface{}); ok {
		validKeys := map[string]bool{
			"tokopedia": true,
			"shopee":    true,
			"lazada":    true,
			"tiktok":    true,
			"bukalapak": true,
		}
		for key := range marketplace {
			if !validKeys[key] {
				handlers.SendError(c, http.StatusBadRequest, "Invalid key in marketplace object: "+key, nil)
				return
			}
		}
	} else {
		handlers.SendError(c, http.StatusBadRequest, "Marketplace object is required", nil)
		return
	}

	// Manually create product from request body
	var product models.Product
	jsonBody, _ := json.Marshal(requestBody)
	if err := json.Unmarshal(jsonBody, &product); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Perform validation using the validation package
	if validationErr := master_product.ValidateCreate(&product); validationErr != nil {
		handlers.SendError(c, http.StatusBadRequest, validationErr.Error, &validationErr.ErrorField)
		return
	}

	// Define fields that need to be converted from IDs to values
	fieldsToConvert := []string{"Grup", "Unit", "Kat", "Gender", "Tipe"}

	// Convert all IDs to values in one call
	helpers.ConvertProductFields(&product, fieldsToConvert)

	product.TanggalUpdate = time.Now()
	err := db.InsertProduct(&product)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			handlers.SendError(c, http.StatusBadRequest, "Product with this artikel already exists", nil)
			return
		}
		handlers.SendError(c, http.StatusInternalServerError, "Failed to create product: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusCreated, product)
}

// UpdateProduct handles updating an existing product
func UpdateProduct(c *gin.Context) {
	artikel := c.Param("artikel")

	// Fetch the existing product first to avoid overwriting with zero values
	existingProduct, err := db.FetchProductByArtikel(artikel)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Product not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch existing product", nil)
		}
		return
	}

	// Create a map to hold the raw JSON request body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Create a product struct with the existing data
	productToUpdate := existingProduct

	// Define string fields mapping and update in one loop
	stringFields := map[string]*string{
		"nama":          &productToUpdate.Nama,
		"deskripsi":     &productToUpdate.Deskripsi,
		"warna":         &productToUpdate.Warna,
		"size":          &productToUpdate.Size,
		"grup":          &productToUpdate.Grup,
		"unit":          &productToUpdate.Unit,
		"kat":           &productToUpdate.Kat,
		"model":         &productToUpdate.Model,
		"gender":        &productToUpdate.Gender,
		"tipe":          &productToUpdate.Tipe,
		"status":        &productToUpdate.Status,
		"supplier":      &productToUpdate.Supplier,
		"diupdate_oleh": &productToUpdate.DiupdateOleh,
	}

	// Process all string fields
	for field, target := range stringFields {
		if value, ok := requestBody[field].(string); ok {
			*target = value
		}
	}

	// Handle numeric field
	if harga, ok := requestBody["harga"].(float64); ok {
		productToUpdate.Harga = harga
	}

	// Handle numeric field
	if hargaDiskon, ok := requestBody["harga_diskon"].(float64); ok {
		productToUpdate.HargaDiskon = hargaDiskon
	}

	// Handle rating field
	if rating, ok := requestBody["rating"].(float64); ok {
		productToUpdate.Rating = rating
	}

	// Handle marketplace field
	if marketplace, ok := requestBody["marketplace"].(map[string]interface{}); ok {
		// Validate keys
		validKeys := map[string]bool{
			"tokopedia": true,
			"shopee":    true,
			"lazada":    true,
			"tiktok":    true,
			"bukalapak": true,
		}
		for key := range marketplace {
			if !validKeys[key] {
				handlers.SendError(c, http.StatusBadRequest, "Invalid key in marketplace object: "+key, nil)
				return
			}
		}

		// Unmarshal to existing struct
		marketplaceJSON, _ := json.Marshal(marketplace)
		if err := json.Unmarshal(marketplaceJSON, &productToUpdate.Marketplace); err != nil {
			handlers.SendError(c, http.StatusBadRequest, "Failed to unmarshal marketplace data: "+err.Error(), nil)
			return
		}
	}

	// Handle gambar field
	if gambar, ok := requestBody["gambar"].([]string); ok {
		productToUpdate.Gambar = gambar
	}

	// Define date fields mapping and update in one loop
	dateFields := map[string]*time.Time{
		"tanggal_produk": &productToUpdate.TanggalProduk,
		"tanggal_terima": &productToUpdate.TanggalTerima,
	}

	// Process all date fields
	for field, target := range dateFields {
		if dateStr, ok := requestBody[field].(string); ok && dateStr != "" {
			date, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				*target = date
			}
		}
	}

	// Update the tanggal_update field to now
	productToUpdate.TanggalUpdate = time.Now()

	// Validate the updated product
	if validationErr := master_product.ValidateUpdate(&productToUpdate); validationErr != nil {
		handlers.SendError(c, http.StatusBadRequest, validationErr.Error, &validationErr.ErrorField)
		return
	}

	// Define fields that need to be converted from IDs to values
	fieldsToConvert := []string{"Grup", "Unit", "Kat", "Gender", "Tipe"}

	// Convert all IDs to values in one call
	helpers.ConvertProductFields(&productToUpdate, fieldsToConvert)

	// Perform the update operation
	updatedProduct, err := db.UpdateProduct(artikel, &productToUpdate)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to update product: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, updatedProduct)
}

// DeleteProduct handles soft-deleting a product
func DeleteProduct(c *gin.Context) {
	artikel := c.Param("artikel")
	err := db.DeleteProduct(artikel)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Product not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to delete product", nil)
		}
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// GetDeletedProducts retrieves all soft-deleted products with pagination
func GetDeletedProducts(c *gin.Context) {
	// Parse pagination parameters
	limit, offset, page, err := helpers.ParsePaginationParams(c)
	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Get query params
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "tanggal_hapus")
	sortDirection := c.DefaultQuery("order", "desc")

	// Extract filter parameters
	filters := helpers.ExtractFilters(c)

	// Fetch total count with filters applied
	totalCount, err := db.CountDeletedProducts(queryStr, filters)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count deleted products", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted products with filters applied
	products, err := db.FetchDeletedProducts(limit, offset, queryStr, filters, sortColumn, sortDirection)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch deleted products", nil)
		return
	}

	// Respond with pagination metadata
	handlers.SendSuccess(c, http.StatusOK, handlers.PaginatedData{
		Items:      products,
		Page:       page,
		TotalPages: totalPages,
		TotalItems: totalCount,
		Filters:    filters,
		Sort:       sortColumn,
		Order:      sortDirection,
	})
}

// RestoreProduct restores a soft-deleted product
func RestoreProduct(c *gin.Context) {
	artikel := c.Param("artikel")

	// First check if the product exists and is deleted
	product, err := db.FetchProductByArtikelIncludeDeleted(artikel)
	if err != nil {
		if err.Error() == "not_found" {
			handlers.SendError(c, http.StatusNotFound, "Product not found", nil)
		} else {
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch product", nil)
		}
		return
	}

	// Check if product is already active (not deleted)
	if product.TanggalHapus == nil {
		handlers.SendError(c, http.StatusBadRequest, "Product is already active and not deleted", nil)
		return
	}

	// Restore the product
	err = db.RestoreProduct(artikel)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to restore product: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusOK, gin.H{"message": "Product restored successfully"})
}
