package adminHandlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/app/helpers"
	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/app/validation/master_product"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

var maxImages = 10

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
	log.Println("--------------------------------")
	log.Println("CreateProduct: Starting product creation process.")

	// Check if the request is multipart form data
	contentType := c.GetHeader("Content-Type")
	log.Printf("CreateProduct: Content-Type: %s", contentType)

	if !strings.Contains(contentType, "multipart/form-data") {
		log.Println("CreateProduct: Warning - Request is not multipart/form-data")
		handlers.SendError(c, http.StatusBadRequest, "Request must be multipart/form-data for file uploads", nil)
		return
	}
	log.Println("")

	var req models.CreateProductRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("CreateProduct: Error binding form data: %v", err)
		handlers.SendError(c, http.StatusBadRequest, "Invalid form data: "+err.Error(), nil)
		return
	}
	log.Printf("CreateProduct: Successfully bound request: %+v", req)
	log.Println("")

	// Unmarshal marketplace JSON string
	var marketplaceInfo models.MarketplaceInfo
	if req.Marketplace != "" {
		if err := json.Unmarshal([]byte(req.Marketplace), &marketplaceInfo); err != nil {
			log.Printf("CreateProduct: Error unmarshaling marketplace JSON: %v", err)
			handlers.SendError(c, http.StatusBadRequest, "Invalid marketplace data format", nil)
			return
		}
		log.Printf("CreateProduct: Successfully unmarshaled marketplaceInfo: %+v", marketplaceInfo)
	}
	log.Println("")

	// Parse date fields
	var tanggalProduk time.Time
	if req.TanggalProduk != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z", req.TanggalProduk)
		if err != nil {
			log.Println("Invalid date format for tanggal_produk:", err)
			errorField := "tanggal_produk"
			handlers.SendError(c, http.StatusBadRequest, "Invalid date format for tanggal_produk, use YYYY-MM-DD", &errorField)
			return
		}
		tanggalProduk = t
	}
	log.Println("")

	var tanggalTerima time.Time
	if req.TanggalTerima != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z", req.TanggalTerima)
		if err != nil {
			log.Println("Invalid date format for tanggal_terima:", err)
			errorField := "tanggal_terima"
			handlers.SendError(c, http.StatusBadRequest, "Invalid date format for tanggal_terima, use YYYY-MM-DD", &errorField)
			return
		}
		tanggalTerima = t
	}
	log.Println("--------------------------------")

	// Handle image uploads
	var imageUrls []string
	log.Println("CreateProduct: Looking for indexed image files (gambar[0], gambar[1], etc.)")

	// Get multipart form to access files with indexed keys
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("CreateProduct: Error getting multipart form: %v", err)
		handlers.SendError(c, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error(), nil)
		return
	}
	log.Println("")

	log.Printf("CreateProduct: Form files: %+v", form.File)
	log.Println("")

	// Collect all gambar files (both indexed and array format)
	var allFiles []*multipart.FileHeader

	// Check for array format (gambar)
	if files, ok := form.File["gambar"]; ok {
		log.Printf("CreateProduct: Found %d files with key 'gambar'", len(files))
		allFiles = append(allFiles, files...)
	}
	log.Println("")

	// Check for indexed format (gambar[0], gambar[1], etc.)
	for imageIndex := range maxImages { // Check up to maxImages files
		key := fmt.Sprintf("gambar[%d]", imageIndex)
		if files, ok := form.File[key]; ok {
			log.Printf("CreateProduct: Found %d files with key '%s'", len(files), key)
			allFiles = append(allFiles, files...)
		}
	}
	log.Println("")

	if len(allFiles) == 0 {
		log.Println("CreateProduct: No files found with keys 'gambar' or 'gambar[x]'")
	} else {
		log.Printf("CreateProduct: Found total %d files", len(allFiles))

		if len(allFiles) > maxImages {
			log.Println("Maximum of " + strconv.Itoa(maxImages) + " images allowed")
			handlers.SendError(c, http.StatusBadRequest, "Maximum of "+strconv.Itoa(maxImages)+" images allowed", nil)
			return
		}
		log.Println("")

		for i, file := range allFiles {
			log.Printf("CreateProduct: Processing file %d: %s, Size: %d", i, file.Filename, file.Size)

			// Check if file has content
			if file.Size == 0 {
				log.Printf("CreateProduct: File %d has zero size, skipping", i)
				continue
			}

			// Handle image upload
			filePath, err := helpers.SaveUploadedFile(c, file, "uploads/products/", nil)
			if err != nil {
				log.Printf("CreateProduct: Failed to save image %d: %v", i, err)
				handlers.SendError(c, http.StatusInternalServerError, "Failed to save image: "+err.Error(), nil)
				return
			}
			imageUrls = append(imageUrls, filePath)
			log.Printf("CreateProduct: Successfully saved image %d to: %s", i, filePath)
		}
	}
	log.Println("")

	// Check if we have any valid images
	if len(imageUrls) == 0 {
		log.Println("CreateProduct: No valid images were uploaded")
		handlers.SendError(c, http.StatusBadRequest, "At least one valid image file is required", nil)
		return
	}
	log.Println("")

	log.Printf("CreateProduct: Successfully processed %d images: %v", len(imageUrls), imageUrls)
	log.Println("--------------------------------")

	product := models.Product{
		Artikel:       req.Artikel,
		Nama:          req.Nama,
		Deskripsi:     req.Deskripsi,
		Warna:         req.Warna,
		Size:          req.Size,
		Grup:          req.Grup,
		Unit:          req.Unit,
		Kat:           req.Kat,
		Model:         req.Model,
		Gender:        req.Gender,
		Tipe:          req.Tipe,
		Harga:         req.Harga,
		HargaDiskon:   req.HargaDiskon,
		Marketplace:   marketplaceInfo,
		Gambar:        imageUrls,
		TanggalProduk: tanggalProduk,
		TanggalTerima: tanggalTerima,
		Status:        req.Status,
		Supplier:      req.Supplier,
		DiupdateOleh:  req.DiupdateOleh,
	}

	log.Println("Product struct created:", product)

	if req.Rating != nil {
		product.Rating = *req.Rating
	} else {
		product.Rating = 0
	}

	log.Printf("CreateProduct: Product struct created: %+v", product)

	// Perform validation using the validation package
	if validationErr := master_product.ValidateCreate(&product); validationErr != nil {
		log.Println("Validation error:", validationErr)
		handlers.SendError(c, http.StatusBadRequest, validationErr.Error, &validationErr.ErrorField)
		return
	}

	// Define fields that need to be converted from IDs to values
	fieldsToConvert := []string{"Grup", "Unit", "Kat", "Gender", "Tipe"}

	// Convert all IDs to values in one call
	helpers.ConvertProductFields(&product, fieldsToConvert)

	product.TanggalUpdate = time.Now()
	log.Println("CreateProduct: Attempting to insert product into DB.")
	err = db.InsertProduct(&product)
	if err != nil {
		log.Printf("CreateProduct: Failed to insert product into DB: %v", err)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			handlers.SendError(c, http.StatusBadRequest, "Product with this artikel already exists", nil)
			return
		}
		handlers.SendError(c, http.StatusInternalServerError, "Failed to create product: "+err.Error(), nil)
		return
	}

	handlers.SendSuccess(c, http.StatusCreated, product)
	log.Println("--------------------------------")
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
