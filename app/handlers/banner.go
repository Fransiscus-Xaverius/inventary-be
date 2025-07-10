package handlers

import (
	"fmt"
	"math"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SaveUploadedFile saves the uploaded file to the specified directory
func SaveUploadedFile(c *gin.Context, file *multipart.FileHeader, destination string) (string, error) {
	// Generate a unique filename
	extension := filepath.Ext(file.Filename)
	filename := uuid.New().String() + extension
	filePath := filepath.Join(destination, filename)

	// Save the file
	err := c.SaveUploadedFile(file, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to save uploaded file: %w", err)
	}

	return "/uploads/" + filename, nil // Return the URL path
}

// GetAllBanners handles fetching all banners with pagination and search
func GetAllBanners(c *gin.Context) {
	// Read query params
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	isActiveStr := c.Query("is_active") // Optional filter for is_active
	sortColumn := c.DefaultQuery("sort", "id")
	sortDirection := c.DefaultQuery("order", "asc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		sendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Parse is_active filter
	var isActive *bool
	if isActiveStr != "" {
		b, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			sendError(c, http.StatusBadRequest, "Invalid is_active parameter", nil)
			return
		}
		isActive = &b
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term and is_active filter applied
	totalCount, err := db.CountAllBanners(queryStr, isActive)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to count banners", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated banners with search term and is_active filter applied
	banners, err := db.FetchAllBanners(limit, offset, queryStr, isActive, sortColumn, sortDirection)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch banners", nil)
		return
	}

	// Respond with pagination metadata
	sendSuccess(c, http.StatusOK, gin.H{
		"banners":    banners,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// GetBannerByID handles fetching a single banner by ID
func GetBannerByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	banner, err := db.FetchBannerByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Banner not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to fetch banner", nil)
		}
		return
	}

	sendSuccess(c, http.StatusOK, banner)
}

// CreateBanner handles creating a new banner with image upload
func CreateBanner(c *gin.Context) {
	var banner models.Banner

	// Bind non-file fields from form data
	if err := c.ShouldBind(&banner); err != nil {
		sendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Handle image upload
	file, err := c.FormFile("image")
	if err != nil {
		sendError(c, http.StatusBadRequest, "Image file is required", nil)
		return
	}

	// Save the file to the uploads directory
	filePath, err := SaveUploadedFile(c, file, "uploads/banners")
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to save image: "+err.Error(), nil)
		return
	}
	banner.ImageUrl = filePath

	// Validate required fields
	if banner.Title == "" {
		sendError(c, http.StatusBadRequest, "Banner title is required", nil)
		return
	}

	// Set creation and update timestamps
	banner.CreatedAt = time.Now()
	banner.UpdatedAt = time.Now()

	// Insert to database
	err = db.InsertBanner(&banner)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create banner: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusCreated, banner)
}

// UpdateBanner handles updating an existing banner with optional image upload
func UpdateBanner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	// Fetch the existing banner first to verify it exists
	existingBanner, err := db.FetchBannerByID(id)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Banner not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to fetch existing banner", nil)
		}
		return
	}

	var bannerToUpdate models.Banner
	// Bind non-file fields from form data
	if err := c.ShouldBind(&bannerToUpdate); err != nil {
		sendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Handle optional image upload
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		// New image uploaded, save it
		filePath, err := SaveUploadedFile(c, file, "uploads")
		if err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to save new image: "+err.Error(), nil)
			return
		}
		bannerToUpdate.ImageUrl = filePath
	} else {
		// No new image, retain existing one
		bannerToUpdate.ImageUrl = existingBanner.ImageUrl
	}

	// Update the record
	updatedBanner, err := db.UpdateBanner(id, &bannerToUpdate)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to update banner: "+err.Error(), nil)
		return
	}

	sendSuccess(c, http.StatusOK, updatedBanner)
}

// DeleteBanner handles soft-deleting a banner
func DeleteBanner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.DeleteBanner(id)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Banner not found", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to delete banner", nil)
		}
		return
	}

	sendSuccess(c, http.StatusOK, gin.H{"message": "Banner deleted successfully"})
}

// GetDeletedBanners retrieves all soft-deleted banners with pagination
func GetDeletedBanners(c *gin.Context) {
	// Read query params
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	queryStr := c.DefaultQuery("q", "")
	sortColumn := c.DefaultQuery("sort", "deleted_at")
	sortDirection := c.DefaultQuery("order", "desc")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		sendError(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	// Get current page from offset
	page := (offset / limit) + 1

	// Fetch total count with search term applied
	totalCount, err := db.CountDeletedBanners(queryStr)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to count deleted banners", nil)
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Fetch paginated deleted banners with search term applied
	banners, err := db.FetchDeletedBanners(limit, offset, queryStr, sortColumn, sortDirection)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch deleted banners", nil)
		return
	}

	// Respond with pagination metadata
	sendSuccess(c, http.StatusOK, gin.H{
		"banners":    banners,
		"page":       page,
		"total_page": totalPages,
		"total":      totalCount,
		"sort":       sortColumn,
		"order":      sortDirection,
	})
}

// RestoreBanner handles restoring a soft-deleted banner
func RestoreBanner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	err = db.RestoreBanner(id)
	if err != nil {
		if err.Error() == "not_found" {
			sendError(c, http.StatusNotFound, "Banner not found or already active", nil)
		} else {
			sendError(c, http.StatusInternalServerError, "Failed to restore banner: "+err.Error(), nil)
		}
		return
	}

	sendSuccess(c, http.StatusOK, gin.H{"message": "Banner restored successfully"})
}

// GetActiveBanners retrieves all active banners, ordered by order_index
func GetActiveBanners(c *gin.Context) {
	isActive := true
	banners, err := db.FetchAllBanners(100, 0, "", &isActive, "order_index", "asc") // Set a reasonable limit, e.g., 100
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch active banners: "+err.Error(), nil)
		return
	}
	sendSuccess(c, http.StatusOK, gin.H{"banners": banners})
}