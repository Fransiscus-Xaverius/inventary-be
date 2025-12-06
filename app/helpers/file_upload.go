package helpers

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	MaxFileSize          = 20 * 1024 * 1024 // 20MB
	AspectRatioTolerance = 0.1
)

// FileUploadOptions defines validation options for file uploads.
type FileUploadOptions struct {
	ValidateAspectRatio  bool
	AllowedAspectRatios  []float64
	AspectRatioTolerance float64
	MinWidth             int
	MinHeight            int
}

// SaveUploadedFile saves an uploaded file with optional validation.
func SaveUploadedFile(c *gin.Context, file *multipart.FileHeader, destination string, opts *FileUploadOptions) (string, error) {
	if file.Size > MaxFileSize {
		return "", fmt.Errorf("file size exceeds the limit of 20MB")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// If validation options are provided, perform validation.
	if opts != nil {
		config, _, err := image.DecodeConfig(src)
		if err != nil {
			return "", fmt.Errorf("invalid image format: %w", err)
		}

		if opts.MinWidth > 0 && opts.MinHeight > 0 {
			if config.Width < opts.MinWidth || config.Height < opts.MinHeight {
				return "", fmt.Errorf("image resolution must be at least %dpx by %dpx", opts.MinWidth, opts.MinHeight)
			}
		}

		if opts.ValidateAspectRatio && len(opts.AllowedAspectRatios) > 0 {
			imageAspectRatio := float64(config.Width) / float64(config.Height)
			valid := false
			tolerance := AspectRatioTolerance
			if opts.AspectRatioTolerance > 0 {
				tolerance = opts.AspectRatioTolerance
			}
			for _, allowedRatio := range opts.AllowedAspectRatios {
				if math.Abs(imageAspectRatio-allowedRatio) <= tolerance {
					valid = true
					break
				}
			}
			if !valid {
				return "", fmt.Errorf("invalid image aspect ratio")
			}
		}

		// Seek back to the beginning of the file after reading for validation
		if _, err := src.Seek(0, 0); err != nil {
			return "", fmt.Errorf("failed to reset file reader: %w", err)
		}
	}

	// Generate a unique filename
	extension := filepath.Ext(file.Filename)
	filename := uuid.New().String() + extension
	filePath := filepath.Join(destination, filename)

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destination, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create the destination file
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Copy the file content
	if _, err = io.Copy(out, src); err != nil {
		return "", fmt.Errorf("failed to save file content: %w", err)
	}

	return "/" + filePath, nil
}

// SaveUploadedFileWithStaticName saves an uploaded file with a static name and optional validation.
func SaveUploadedFileWithStaticName(c *gin.Context, file *multipart.FileHeader, destination string, staticFilename string, opts *FileUploadOptions) (string, error) {
	if file.Size > MaxFileSize {
		return "", fmt.Errorf("file size exceeds the limit of 20MB")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// If validation options are provided, perform validation.
	if opts != nil {
		config, _, err := image.DecodeConfig(src)
		if err != nil {
			return "", fmt.Errorf("invalid image format: %w", err)
		}

		if opts.MinWidth > 0 && opts.MinHeight > 0 {
			if config.Width < opts.MinWidth || config.Height < opts.MinHeight {
				return "", fmt.Errorf("image resolution must be at least %dpx by %dpx", opts.MinWidth, opts.MinHeight)
			}
		}

		if opts.ValidateAspectRatio && len(opts.AllowedAspectRatios) > 0 {
			imageAspectRatio := float64(config.Width) / float64(config.Height)
			valid := false
			tolerance := AspectRatioTolerance
			if opts.AspectRatioTolerance > 0 {
				tolerance = opts.AspectRatioTolerance
			}
			for _, allowedRatio := range opts.AllowedAspectRatios {
				if math.Abs(imageAspectRatio-allowedRatio) <= tolerance {
					valid = true
					break
				}
			}
			if !valid {
				return "", fmt.Errorf("invalid image aspect ratio")
			}
		}

		// Seek back to the beginning of the file after reading for validation
		if _, err := src.Seek(0, 0); err != nil {
			return "", fmt.Errorf("failed to reset file reader: %w", err)
		}
	}

	filePath := filepath.Join(destination, staticFilename)

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destination, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create the destination file
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Copy the file content
	if _, err = io.Copy(out, src); err != nil {
		return "", fmt.Errorf("failed to save file content: %w", err)
	}

	return "/" + filePath, nil
}
