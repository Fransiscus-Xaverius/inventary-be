package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// ImageMetadataEntry represents a single entry in the images_metadata array.
// It describes whether an image at a specific index is an existing image (URL)
// or a new image (file upload).
type ImageMetadataEntry struct {
	Index int    `json:"index"` // Position in the final images array (0 = main image)
	Type  string `json:"type"`  // Either "existing" or "new"
	URL   string `json:"url"`   // The relative path of existing image (only for type="existing")
}

// ProductImageProcessor handles the unified image processing for products.
// It parses the images_metadata JSON, validates entries, processes file uploads,
// and returns the final ordered array of image URLs.
type ProductImageProcessor struct {
	Metadata      []ImageMetadataEntry
	FinalImages   []string
	UploadedFiles []string // Track uploaded files for cleanup on error
	MaxImages     int
	UploadDir     string
	UploadOptions *FileUploadOptions
}

// NewProductImageProcessor creates a new processor with default settings.
func NewProductImageProcessor(maxImages int, uploadDir string, opts *FileUploadOptions) *ProductImageProcessor {
	return &ProductImageProcessor{
		MaxImages:     maxImages,
		UploadDir:     uploadDir,
		UploadOptions: opts,
	}
}

// ParseMetadata parses and validates the images_metadata JSON string.
// It performs the following validations:
// - Metadata must be valid JSON
// - At least one image is required
// - Maximum 10 images allowed
// - Main image (index 0) is required
// - Indices must be contiguous starting from 0
func (p *ProductImageProcessor) ParseMetadata(metadataStr string) error {
	if metadataStr == "" {
		return fmt.Errorf("images_metadata is required")
	}

	// Parse JSON
	if err := json.Unmarshal([]byte(metadataStr), &p.Metadata); err != nil {
		log.Printf("ProductImageProcessor: Failed to parse images_metadata JSON: %v", err)
		return fmt.Errorf("images_metadata must be valid JSON: %v", err)
	}

	// Validate structure
	if len(p.Metadata) == 0 {
		return fmt.Errorf("at least one image is required")
	}

	if len(p.Metadata) > p.MaxImages {
		return fmt.Errorf("maximum %d images allowed", p.MaxImages)
	}

	// Check for main image (index 0)
	hasMainImage := false
	for _, entry := range p.Metadata {
		if entry.Index == 0 {
			hasMainImage = true
			break
		}
	}
	if !hasMainImage {
		return fmt.Errorf("main image (index 0) is required")
	}

	// Check for contiguous indices
	indices := make([]int, len(p.Metadata))
	for i, entry := range p.Metadata {
		indices[i] = entry.Index
	}
	sort.Ints(indices)
	for i, idx := range indices {
		if idx != i {
			return fmt.Errorf("image indices must be contiguous starting from 0")
		}
	}

	// Initialize final images array
	p.FinalImages = make([]string, len(p.Metadata))
	p.UploadedFiles = make([]string, 0)

	log.Printf("ProductImageProcessor: Successfully parsed metadata with %d entries", len(p.Metadata))
	return nil
}

// ProcessImages processes each metadata entry, handling both existing images
// and new file uploads. It validates existing image paths and saves new uploads.
func (p *ProductImageProcessor) ProcessImages(c *gin.Context, form *multipart.Form) error {
	for _, entry := range p.Metadata {
		log.Printf("ProductImageProcessor: Processing entry - Index: %d, Type: %s, URL: %s",
			entry.Index, entry.Type, entry.URL)

		switch entry.Type {
		case "existing":
			if err := p.processExistingImage(entry); err != nil {
				return err
			}
		case "new":
			if err := p.processNewImage(c, form, entry); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid type '%s' at index %d", entry.Type, entry.Index)
		}
	}

	log.Printf("ProductImageProcessor: Successfully processed all images: %v", p.FinalImages)
	return nil
}

// processExistingImage validates and processes an existing image entry.
func (p *ProductImageProcessor) processExistingImage(entry ImageMetadataEntry) error {
	// Validate URL is provided
	if entry.URL == "" {
		return fmt.Errorf("missing URL for existing image at index %d", entry.Index)
	}

	// Security: Validate the path
	if !IsValidImagePath(entry.URL) {
		return fmt.Errorf("invalid image path at index %d", entry.Index)
	}

	// Verify file exists on disk
	filePath := strings.TrimPrefix(entry.URL, "/")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("ProductImageProcessor: Image file not found at path: %s", filePath)
		return fmt.Errorf("image file not found at index %d", entry.Index)
	}

	p.FinalImages[entry.Index] = entry.URL
	log.Printf("ProductImageProcessor: Kept existing image at index %d: %s", entry.Index, entry.URL)
	return nil
}

// processNewImage handles file upload for a new image entry.
func (p *ProductImageProcessor) processNewImage(c *gin.Context, form *multipart.Form, entry ImageMetadataEntry) error {
	fileKey := fmt.Sprintf("images_new_%d", entry.Index)

	files, ok := form.File[fileKey]
	if !ok || len(files) == 0 {
		return fmt.Errorf("missing file upload for new image at index %d", entry.Index)
	}

	file := files[0]
	log.Printf("ProductImageProcessor: Processing new file at index %d: %s (size: %d bytes)",
		entry.Index, file.Filename, file.Size)

	// Validate file has content
	if file.Size == 0 {
		return fmt.Errorf("empty file upload at index %d", entry.Index)
	}

	// Save the file
	savedPath, err := SaveUploadedFile(c, file, p.UploadDir, p.UploadOptions)
	if err != nil {
		log.Printf("ProductImageProcessor: Failed to save image at index %d: %v", entry.Index, err)
		return fmt.Errorf("failed to save image at index %d: %v", entry.Index, err)
	}

	// Track uploaded file for potential cleanup
	p.UploadedFiles = append(p.UploadedFiles, savedPath)
	p.FinalImages[entry.Index] = savedPath
	log.Printf("ProductImageProcessor: Saved new image at index %d: %s", entry.Index, savedPath)
	return nil
}

// Cleanup removes any files that were uploaded during processing.
// This should be called when an error occurs after some files have been saved.
func (p *ProductImageProcessor) Cleanup() {
	for _, filePath := range p.UploadedFiles {
		if filePath == "" {
			continue
		}
		path := strings.TrimPrefix(filePath, "/")
		if err := os.Remove(path); err != nil {
			log.Printf("ProductImageProcessor: Warning - Failed to cleanup file %s: %v", path, err)
		} else {
			log.Printf("ProductImageProcessor: Cleaned up file: %s", path)
		}
	}
}

// GetFinalImages returns the final ordered array of image URLs.
func (p *ProductImageProcessor) GetFinalImages() []string {
	return p.FinalImages
}

// IsValidImagePath validates that a path is within allowed directories
// and doesn't contain path traversal sequences.
// SECURITY: This prevents path traversal attacks.
func IsValidImagePath(path string) bool {
	// Must start with the uploads directory
	validPrefixes := []string{"/uploads/products/", "/uploads/product/"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(path, prefix) {
			hasValidPrefix = true
			break
		}
	}
	if !hasValidPrefix {
		log.Printf("IsValidImagePath: Path '%s' does not start with allowed prefix", path)
		return false
	}

	// Must not contain path traversal sequences
	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		log.Printf("IsValidImagePath: Path '%s' contains path traversal sequences", path)
		return false
	}

	// Must end with a valid image extension
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	ext := strings.ToLower(filepath.Ext(path))
	hasValidExtension := false
	for _, validExt := range validExtensions {
		if ext == validExt {
			hasValidExtension = true
			break
		}
	}
	if !hasValidExtension {
		log.Printf("IsValidImagePath: Path '%s' has invalid extension '%s'", path, ext)
		return false
	}

	return true
}

// CleanupOrphanedImages compares old and new image arrays and deletes
// images that are no longer referenced.
func CleanupOrphanedImages(oldImages []string, newImages []string) {
	// Create a set of new images for fast lookup
	newImageSet := make(map[string]bool)
	for _, img := range newImages {
		newImageSet[img] = true
	}

	// Delete images that are in old but not in new
	for _, oldImg := range oldImages {
		if oldImg == "" {
			continue
		}
		if !newImageSet[oldImg] {
			// This image is no longer referenced, delete it
			filePath := strings.TrimPrefix(oldImg, "/")
			if err := os.Remove(filePath); err != nil {
				log.Printf("CleanupOrphanedImages: Warning - Failed to delete orphaned image %s: %v",
					filePath, err)
			} else {
				log.Printf("CleanupOrphanedImages: Deleted orphaned image: %s", filePath)
			}
		}
	}
}
