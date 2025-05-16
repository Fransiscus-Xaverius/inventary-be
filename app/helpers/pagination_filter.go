package helpers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParsePaginationParams extracts and validates pagination parameters from request
func ParsePaginationParams(c *gin.Context) (limit int, offset int, page int, err error) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		err = errors.New("invalid pagination parameters")
		return
	}

	// Get current page from offset
	page = (offset / limit) + 1
	return
}

// ExtractFilters gets filter parameters from the request
func ExtractFilters(c *gin.Context) map[string]string {
	filters := make(map[string]string)
	validFilterFields := []string{"warna", "size", "grup", "unit", "kat", "model", "gender", "tipe", "status", "supplier"}

	for _, field := range validFilterFields {
		if value := c.Query(field); value != "" {
			filters[field] = value
		}
	}
	return filters
}
