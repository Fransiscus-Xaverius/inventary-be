package adminHandlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
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

// deleteImageFiles removes image files from the filesystem
func deleteImageFiles(imageUrls []string) {
	for _, imageUrl := range imageUrls {
		if imageUrl != "" {
			// Remove the leading "/" from the URL to get the file path
			filePath := strings.TrimPrefix(imageUrl, "/")
			if err := os.Remove(filePath); err != nil {
				log.Printf("Warning: Failed to delete image file %s: %v", filePath, err)
			} else {
				log.Printf("Successfully deleted old image file: %s", filePath)
			}
		}
	}
}

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

	// Extract marketplace and offline filters
	isMarketplaceFilter := strings.Contains(c.Request.URL.RawQuery, "online")
	isOfflineFilter := strings.Contains(c.Request.URL.RawQuery, "offline")

	// Fetch total count with filters applied
	totalCount, err := db.CountAllProducts(queryStr, filters, isMarketplaceFilter, isOfflineFilter)
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to count products", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated products with filters applied
	products, err := db.FetchAllProducts(limit, offset, queryStr, filters, sortColumn, sortDirection, isMarketplaceFilter, isOfflineFilter)
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

	// Unmarshal offline JSON string
	var offlineInfo models.OfflineStores
	if req.Offline != "" {
		if err := json.Unmarshal([]byte(req.Offline), &offlineInfo); err != nil {
			log.Printf("CreateProduct: Error unmarshaling offline JSON: %v", err)
			handlers.SendError(c, http.StatusBadRequest, "Invalid offline data format", nil)
			return
		}
		log.Printf("CreateProduct: Successfully unmarshaled offlineInfo: %+v", offlineInfo)

		// Set default is_active to true for any stores that don't specify it
		for i := range offlineInfo {
			// If is_active is not set (false), default it to true
			if !offlineInfo[i].IsActive {
				offlineInfo[i].IsActive = true
			}
		}
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
		HargaDiskon:   req.HargaDiskon, // This is already *float64
		Marketplace:   marketplaceInfo,
		Offline:       offlineInfo,
		Gambar:        imageUrls,
		TanggalProduk: tanggalProduk,
		TanggalTerima: tanggalTerima,
		Status:        req.Status,
		Supplier:      req.Supplier,
		DiupdateOleh:  req.DiupdateOleh,
	}

	log.Println("Product struct created:", product)

	// Handle rating JSON
	if req.Rating != "" {
		var rating models.ProductRating
		if err := json.Unmarshal([]byte(req.Rating), &rating); err != nil {
			log.Printf("CreateProduct: Failed to unmarshal rating JSON: %v", err)
			handlers.SendError(c, http.StatusBadRequest, "Invalid rating JSON format: "+err.Error(), nil)
			return
		}
		product.Rating = rating
	} else {
		// Set default rating
		product.Rating = models.ProductRating{
			Comfort: 0,
			Style:   0,
			Support: 0,
			Purpose: []string{""},
		}
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
	log.Println("--------------------------------")
	log.Println("UpdateProduct: Starting product update process.")

	artikel := c.Param("artikel")
	log.Printf("UpdateProduct: Artikel parameter: %s", artikel)

	// Fetch the existing product first to avoid overwriting with zero values
	existingProduct, err := db.FetchProductByArtikel(artikel)
	if err != nil {
		if err.Error() == "not_found" {
			log.Printf("UpdateProduct: Product with artikel %s not found", artikel)
			handlers.SendError(c, http.StatusNotFound, "Product not found", nil)
		} else {
			log.Printf("UpdateProduct: Error fetching existing product: %v", err)
			handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch existing product", nil)
		}
		return
	}
	log.Printf("UpdateProduct: Successfully fetched existing product: %+v", existingProduct)
	log.Println("")

	// Check content type to determine request format
	contentType := c.GetHeader("Content-Type")
	log.Printf("UpdateProduct: Content-Type: %s", contentType)
	isMultipart := strings.Contains(contentType, "multipart/form-data")
	log.Printf("UpdateProduct: Is multipart request: %t", isMultipart)
	log.Println("")

	// Variables for request processing
	var requestBody map[string]interface{}
	var imageUrls []string
	var imageProcessed bool

	if isMultipart {
		log.Println("UpdateProduct: Processing multipart/form-data request")

		// Get multipart form to access files and form data
		form, err := c.MultipartForm()
		if err != nil {
			log.Printf("UpdateProduct: Error getting multipart form: %v", err)
			handlers.SendError(c, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error(), nil)
			return
		}
		log.Printf("UpdateProduct: Form files: %+v", form.File)
		log.Printf("UpdateProduct: Form values: %+v", form.Value)
		log.Println("")

		// Convert form values to requestBody map
		requestBody = make(map[string]interface{})
		for key, values := range form.Value {
			if len(values) > 0 {
				requestBody[key] = values[0] // Take first value
			}
		}
		log.Printf("UpdateProduct: Converted form values to requestBody: %+v", requestBody)
		log.Println("")

		// Handle marketplace JSON parsing for multipart
		if marketplaceStr, ok := requestBody["marketplace"].(string); ok && marketplaceStr != "" {
			var marketplaceInfo map[string]interface{}
			if err := json.Unmarshal([]byte(marketplaceStr), &marketplaceInfo); err != nil {
				log.Printf("UpdateProduct: Error unmarshaling marketplace JSON: %v", err)
				handlers.SendError(c, http.StatusBadRequest, "Invalid marketplace data format", nil)
				return
			}
			requestBody["marketplace"] = marketplaceInfo
			log.Printf("UpdateProduct: Successfully unmarshaled marketplace: %+v", marketplaceInfo)
		} else if marketplaceStr, ok := requestBody["marketplace"].(string); ok && marketplaceStr == "" {
			log.Printf("UpdateProduct: Empty marketplace string received - will clear marketplace")
		}
		log.Printf("UpdateProduct: Final requestBody keys: %v", getMapKeys(requestBody))
		log.Println("")

		// Handle image uploads
		log.Println("UpdateProduct: Looking for image files (gambar[0], gambar[1], etc.)")

		// Collect all gambar files (both indexed and array format)
		var allFiles []*multipart.FileHeader

		// Check for array format (gambar)
		if files, ok := form.File["gambar"]; ok {
			log.Printf("UpdateProduct: Found %d files with key 'gambar'", len(files))
			allFiles = append(allFiles, files...)
		}

		// Check for indexed format (gambar[0], gambar[1], etc.)
		for imageIndex := range maxImages { // Check up to maxImages files
			key := fmt.Sprintf("gambar[%d]", imageIndex)
			if files, ok := form.File[key]; ok {
				log.Printf("UpdateProduct: Found %d files with key '%s'", len(files), key)
				allFiles = append(allFiles, files...)
			}
		}
		log.Println("")

		if len(allFiles) == 0 {
			log.Println("UpdateProduct: No image files found - keeping existing images")
		} else {
			log.Printf("UpdateProduct: Found total %d files", len(allFiles))

			if len(allFiles) > maxImages {
				log.Println("UpdateProduct: Maximum of " + strconv.Itoa(maxImages) + " images allowed")
				handlers.SendError(c, http.StatusBadRequest, "Maximum of "+strconv.Itoa(maxImages)+" images allowed", nil)
				return
			}

			for i, file := range allFiles {
				log.Printf("UpdateProduct: Processing file %d: %s, Size: %d", i, file.Filename, file.Size)

				// Check if file has content
				if file.Size == 0 {
					log.Printf("UpdateProduct: File %d has zero size, skipping", i)
					continue
				}

				// Handle image upload
				filePath, err := helpers.SaveUploadedFile(c, file, "uploads/products/", nil)
				if err != nil {
					log.Printf("UpdateProduct: Failed to save image %d: %v", i, err)
					handlers.SendError(c, http.StatusInternalServerError, "Failed to save image: "+err.Error(), nil)
					return
				}
				imageUrls = append(imageUrls, filePath)
				log.Printf("UpdateProduct: Successfully saved image %d to: %s", i, filePath)
			}

			if len(imageUrls) > 0 {
				imageProcessed = true
				log.Printf("UpdateProduct: Successfully processed %d images: %v", len(imageUrls), imageUrls)
			}
		}
		log.Println("")

	} else {
		log.Println("UpdateProduct: Processing JSON request")

		// Handle JSON request (existing logic)
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			log.Printf("UpdateProduct: Error binding JSON: %v", err)
			handlers.SendError(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		log.Printf("UpdateProduct: Successfully bound JSON request: %+v", requestBody)
		log.Printf("UpdateProduct: JSON requestBody keys: %v", getMapKeys(requestBody))
		log.Println("")
	}

	// Create a product struct with the existing data
	productToUpdate := existingProduct
	log.Println("UpdateProduct: Created productToUpdate with existing data")

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
			log.Printf("UpdateProduct: Updating field %s from '%s' to '%s'", field, *target, value)
			*target = value
		}
	}

	// Handle numeric field
	if harga, ok := requestBody["harga"].(float64); ok {
		log.Printf("UpdateProduct: Updating harga from %f to %f", productToUpdate.Harga, harga)
		productToUpdate.Harga = harga
	}

	// Handle harga_diskon field properly as *float64
	if hargaDiskon, ok := requestBody["harga_diskon"].(float64); ok {
		oldValue := "nil"
		if productToUpdate.HargaDiskon != nil {
			oldValue = fmt.Sprintf("%f", *productToUpdate.HargaDiskon)
		}
		log.Printf("UpdateProduct: Updating harga_diskon from %s to %f", oldValue, hargaDiskon)
		productToUpdate.HargaDiskon = &hargaDiskon
	}

	// Handle rating field (can come as either JSON object or JSON string)
	if ratingRaw, exists := requestBody["rating"]; exists {
		var rating map[string]interface{}
		// var err error

		// Try to handle as map[string]interface{} first (parsed JSON object)
		if ratingMap, ok := ratingRaw.(map[string]interface{}); ok {
			rating = ratingMap
			log.Printf("UpdateProduct: Rating received as parsed JSON object: %+v", rating)
		} else if ratingStr, ok := ratingRaw.(string); ok {
			// Handle as JSON string (needs parsing)
			log.Printf("UpdateProduct: Rating received as JSON string: %s", ratingStr)
			if err := json.Unmarshal([]byte(ratingStr), &rating); err != nil {
				log.Printf("UpdateProduct: Failed to parse rating JSON string: %v", err)
				handlers.SendError(c, http.StatusBadRequest, "Invalid rating JSON format: "+err.Error(), nil)
				return
			}
			log.Printf("UpdateProduct: Successfully parsed rating JSON string: %+v", rating)
		} else {
			log.Printf("UpdateProduct: Rating field has unsupported type: %T, value: %+v", ratingRaw, ratingRaw)
			handlers.SendError(c, http.StatusBadRequest, "Rating field must be a JSON object or JSON string", nil)
			return
		}

		// Validate required keys
		requiredKeys := []string{"comfort", "style", "support", "purpose"}
		for _, key := range requiredKeys {
			if _, exists := rating[key]; !exists {
				log.Printf("UpdateProduct: Missing required rating key: %s", key)
				handlers.SendError(c, http.StatusBadRequest, "Missing required rating key: "+key, nil)
				return
			}
		}

		// Unmarshal to ProductRating struct
		ratingJSON, _ := json.Marshal(rating)
		var newRating models.ProductRating
		if err := json.Unmarshal(ratingJSON, &newRating); err != nil {
			log.Printf("UpdateProduct: Failed to unmarshal rating data: %v", err)
			handlers.SendError(c, http.StatusBadRequest, "Failed to unmarshal rating data: "+err.Error(), nil)
			return
		}
		log.Printf("UpdateProduct: Successfully updated rating: %+v", newRating)
		productToUpdate.Rating = newRating
	}

	// Handle marketplace field
	if marketplace, ok := requestBody["marketplace"].(map[string]interface{}); ok {
		log.Printf("UpdateProduct: Processing marketplace update: %+v", marketplace)
		log.Printf("UpdateProduct: Marketplace length: %d", len(marketplace))
		log.Printf("UpdateProduct: Marketplace keys: %v", getMapKeys(marketplace))

		// Check if marketplace is empty (clearing the field)
		if len(marketplace) == 0 {
			log.Printf("UpdateProduct: Clearing marketplace field")
			productToUpdate.Marketplace = models.MarketplaceInfo{}
		} else {
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
					log.Printf("UpdateProduct: Invalid marketplace key: %s", key)
					handlers.SendError(c, http.StatusBadRequest, "Invalid key in marketplace object: "+key, nil)
					return
				}
			}

			// Clear existing marketplace struct to ensure old values are removed
			productToUpdate.Marketplace = models.MarketplaceInfo{}

			// Unmarshal to the cleared struct
			marketplaceJSON, _ := json.Marshal(marketplace)
			if err := json.Unmarshal(marketplaceJSON, &productToUpdate.Marketplace); err != nil {
				log.Printf("UpdateProduct: Failed to unmarshal marketplace data: %v", err)
				handlers.SendError(c, http.StatusBadRequest, "Failed to unmarshal marketplace data: "+err.Error(), nil)
				return
			}
			v := reflect.ValueOf(productToUpdate.Marketplace)
			typeOfS := v.Type()

			for i := 0; i < v.NumField(); i++ {
				log.Printf("UpdateProduct: Marketplace key: %s, value: %+v", typeOfS.Field(i).Name, v.Field(i).Interface())
			}
		}
		log.Printf("UpdateProduct: Successfully updated marketplace: %+v", productToUpdate.Marketplace)
		log.Printf("UpdateProduct: Marketplace after update - Tokopedia: %v, Shopee: %v, Lazada: %v, Tiktok: %v, Bukalapak: %v",
			productToUpdate.Marketplace.Tokopedia, productToUpdate.Marketplace.Shopee, productToUpdate.Marketplace.Lazada,
			productToUpdate.Marketplace.Tiktok, productToUpdate.Marketplace.Bukalapak)
	} else {
		log.Printf("UpdateProduct: No marketplace field in request body")
	}

	// Handle offline field - check for both string and array formats
	if offlineStr, ok := requestBody["offline"].(string); ok {
		if offlineStr == "" {
			log.Printf("UpdateProduct: Clearing offline field (empty string)")
			productToUpdate.Offline = models.OfflineStores{}
		} else {
			log.Printf("UpdateProduct: Processing offline update (string format): %s", offlineStr)

			var offlineStores models.OfflineStores
			if err := json.Unmarshal([]byte(offlineStr), &offlineStores); err != nil {
				log.Printf("UpdateProduct: Failed to unmarshal offline data: %v", err)
				handlers.SendError(c, http.StatusBadRequest, "Failed to unmarshal offline data: "+err.Error(), nil)
				return
			}

			// Set default is_active to true for any stores that don't specify it
			for i := range offlineStores {
				if !offlineStores[i].IsActive {
					offlineStores[i].IsActive = true
				}
			}

			productToUpdate.Offline = offlineStores
		}
		log.Printf("UpdateProduct: Successfully updated offline: %+v", productToUpdate.Offline)
	} else if offline, ok := requestBody["offline"].([]interface{}); ok {
		log.Printf("UpdateProduct: Processing offline update (array format): %+v", offline)

		// Check if offline array is empty (clearing the field)
		if len(offline) == 0 {
			log.Printf("UpdateProduct: Clearing offline field (empty array)")
			productToUpdate.Offline = models.OfflineStores{}
		} else {
			var offlineStores models.OfflineStores
			offlineJSON, _ := json.Marshal(offline)
			if err := json.Unmarshal(offlineJSON, &offlineStores); err != nil {
				log.Printf("UpdateProduct: Failed to unmarshal offline data: %v", err)
				handlers.SendError(c, http.StatusBadRequest, "Failed to unmarshal offline data: "+err.Error(), nil)
				return
			}

			// Set default is_active to true for any stores that don't specify it
			for i := range offlineStores {
				if !offlineStores[i].IsActive {
					offlineStores[i].IsActive = true
				}
			}

			productToUpdate.Offline = offlineStores
		}
		log.Printf("UpdateProduct: Successfully updated offline: %+v", productToUpdate.Offline)
	}

	// Handle image updates
	if imageProcessed {
		log.Printf("UpdateProduct: Replacing images with newly uploaded ones")
		log.Printf("UpdateProduct: Old images: %v", existingProduct.Gambar)
		log.Printf("UpdateProduct: New images: %v", imageUrls)

		// Delete old image files
		if len(existingProduct.Gambar) > 0 {
			log.Println("UpdateProduct: Deleting old image files")
			deleteImageFiles(existingProduct.Gambar)
		}

		// Set new images
		productToUpdate.Gambar = imageUrls
		log.Println("UpdateProduct: Successfully replaced images with uploaded files")
	} else if gambar, ok := requestBody["gambar"].([]interface{}); ok {
		// Handle gambar field as array of strings from JSON
		var gambarStrings []string
		for _, img := range gambar {
			if imgStr, ok := img.(string); ok {
				gambarStrings = append(gambarStrings, imgStr)
			}
		}
		if len(gambarStrings) > 0 {
			log.Printf("UpdateProduct: Updating gambar URLs from %v to %v", productToUpdate.Gambar, gambarStrings)

			// Delete old image files
			if len(existingProduct.Gambar) > 0 {
				log.Println("UpdateProduct: Deleting old image files")
				deleteImageFiles(existingProduct.Gambar)
			}

			productToUpdate.Gambar = gambarStrings
		}
	} else if gambarSlice, ok := requestBody["gambar"].([]string); ok {
		// Handle gambar field as slice of strings from JSON
		log.Printf("UpdateProduct: Updating gambar URLs from %v to %v", productToUpdate.Gambar, gambarSlice)

		// Delete old image files
		if len(existingProduct.Gambar) > 0 {
			log.Println("UpdateProduct: Deleting old image files")
			deleteImageFiles(existingProduct.Gambar)
		}

		productToUpdate.Gambar = gambarSlice
	} else {
		log.Println("UpdateProduct: No image updates - keeping existing images")
	}
	log.Println("")

	// Define date fields mapping and update in one loop
	dateFields := map[string]*time.Time{
		"tanggal_produk": &productToUpdate.TanggalProduk,
		"tanggal_terima": &productToUpdate.TanggalTerima,
	}

	// Process all date fields
	for field, target := range dateFields {
		if dateStr, ok := requestBody[field].(string); ok && dateStr != "" {
			// Try parsing different date formats
			var date time.Time
			var err error

			// Try ISO format first (from multipart forms)
			if date, err = time.Parse("2006-01-02T15:04:05Z", dateStr); err != nil {
				// Try simple date format (from JSON)
				if date, err = time.Parse("2006-01-02", dateStr); err != nil {
					log.Printf("UpdateProduct: Failed to parse date for %s: %s, error: %v", field, dateStr, err)
					continue
				}
			}

			log.Printf("UpdateProduct: Updating %s from %v to %v", field, *target, date)
			*target = date
		}
	}

	// Update the tanggal_update field to now
	productToUpdate.TanggalUpdate = time.Now()
	log.Printf("UpdateProduct: Updated tanggal_update to: %v", productToUpdate.TanggalUpdate)
	log.Println("")

	// Validate the updated product
	log.Println("UpdateProduct: Validating updated product")
	if validationErr := master_product.ValidateUpdate(&productToUpdate); validationErr != nil {
		log.Printf("UpdateProduct: Validation error: %s", validationErr.Error)
		handlers.SendError(c, http.StatusBadRequest, validationErr.Error, &validationErr.ErrorField)
		return
	}
	log.Println("UpdateProduct: Product validation successful")

	// Define fields that need to be converted from IDs to values
	fieldsToConvert := []string{"Grup", "Unit", "Kat", "Gender", "Tipe"}

	// Convert all IDs to values in one call
	log.Println("UpdateProduct: Converting product fields from IDs to values")
	helpers.ConvertProductFields(&productToUpdate, fieldsToConvert)
	log.Printf("UpdateProduct: Product after field conversion: %+v", productToUpdate)

	// Perform the update operation
	log.Println("UpdateProduct: Attempting to update product in database")
	updatedProduct, err := db.UpdateProduct(artikel, &productToUpdate)
	if err != nil {
		log.Printf("UpdateProduct: Failed to update product in database: %v", err)
		handlers.SendError(c, http.StatusInternalServerError, "Failed to update product: "+err.Error(), nil)
		return
	}

	log.Printf("UpdateProduct: Successfully updated product: %+v", updatedProduct)
	handlers.SendSuccess(c, http.StatusOK, updatedProduct)
	log.Println("--------------------------------")
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

// TestFileUpload is a debug endpoint to test file uploads
func TestFileUpload(c *gin.Context) {
	log.Println("TestFileUpload: Starting file upload test")

	// Check content type
	contentType := c.GetHeader("Content-Type")
	log.Printf("TestFileUpload: Content-Type: %s", contentType)

	// Try to get multipart form
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("TestFileUpload: Error getting multipart form: %v", err)
		handlers.SendError(c, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error(), nil)
		return
	}

	log.Printf("TestFileUpload: Form files: %+v", form.File)

	// Check for gambar files
	if files, ok := form.File["gambar"]; ok {
		log.Printf("TestFileUpload: Found %d files with key 'gambar'", len(files))
		for i, file := range files {
			log.Printf("TestFileUpload: File %d - Filename: %s, Size: %d, Header: %+v",
				i, file.Filename, file.Size, file.Header)
		}
	} else {
		log.Println("TestFileUpload: No files found with key 'gambar'")
	}

	// Check all form fields
	log.Printf("TestFileUpload: All form fields: %+v", form.Value)

	handlers.SendSuccess(c, http.StatusOK, gin.H{
		"message":      "File upload test completed",
		"files_found":  len(form.File),
		"content_type": contentType,
	})
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
