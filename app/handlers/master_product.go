package handlers

import (
	"math"
	"net/http"
	"strings"
	"time"

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
		sendError(c, http.StatusBadRequest, err.Error(), nil)
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
		sendError(c, http.StatusInternalServerError, "Failed to count products", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated products with filters applied
	products, err := db.FetchAllProducts(limit, offset, queryStr, filters, sortColumn, sortDirection)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch products", nil)
		return
	}

	// Respond with pagination metadata
	sendSuccess(c, http.StatusOK, PaginatedData{
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
			sendError(c, http.StatusNotFound, "Product not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to fetch product", nil)
		}
		return
	}

	sendSuccess(c, http.StatusOK, product)
}

// CreateProduct handles creating a new product
func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		sendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Perform validation using the validation package
	if validationErr := master_product.ValidateCreate(&product); validationErr != nil {
		sendError(c, http.StatusBadRequest, validationErr.Error, &validationErr.ErrorField)
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
			sendError(c, http.StatusBadRequest, "Product with this artikel already exists", nil)
			return
		}
		sendError(c, http.StatusInternalServerError, "Failed to create product: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusCreated, product)
}

// UpdateProduct handles updating an existing product
func UpdateProduct(c *gin.Context) {
	artikel := c.Param("artikel")

	// Fetch the existing product first to avoid overwriting with zero values
	existingProduct, err := db.FetchProductByArtikel(artikel)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Product not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to fetch existing product", nil)
		}
		return
	}

	// Create a map to hold the raw JSON request body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		sendError(c, http.StatusBadRequest, err.Error(), nil)
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
		sendError(c, http.StatusBadRequest, validationErr.Error, &validationErr.ErrorField)
		return
	}

	// Define fields that need to be converted from IDs to values
	fieldsToConvert := []string{"Grup", "Unit", "Kat", "Gender", "Tipe"}

	// Convert all IDs to values in one call
	helpers.ConvertProductFields(&productToUpdate, fieldsToConvert)

	// Perform the update operation
	updatedProduct, err := db.UpdateProduct(artikel, &productToUpdate)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to update product: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusOK, updatedProduct)
}

// DeleteProduct handles soft-deleting a product
func DeleteProduct(c *gin.Context) {
	artikel := c.Param("artikel")
	err := db.DeleteProduct(artikel)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Product not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to delete product", nil)
		}
		return
	}

	sendSuccess(c, http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// GetDeletedProducts retrieves all soft-deleted products with pagination
func GetDeletedProducts(c *gin.Context) {
	// Parse pagination parameters
	limit, offset, page, err := helpers.ParsePaginationParams(c)
	if err != nil {
		sendError(c, http.StatusBadRequest, err.Error(), nil)
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
		sendError(c, http.StatusInternalServerError, "Failed to count deleted products", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted products with filters applied
	products, err := db.FetchDeletedProducts(limit, offset, queryStr, filters, sortColumn, sortDirection)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch deleted products", nil)
		return
	}

	// Respond with pagination metadata
	sendSuccess(c, http.StatusOK, PaginatedData{
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
			sendError(c, http.StatusNotFound, "Product not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to fetch product", nil)
		}
		return
	}

	// Check if product is already active (not deleted)
	if product.TanggalHapus == nil {
		sendError(c, http.StatusBadRequest, "Product is already active and not deleted", nil)
		return
	}

	// Restore the product
	err = db.RestoreProduct(artikel)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to restore product: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusOK, gin.H{"message": "Product restored successfully"})
}
