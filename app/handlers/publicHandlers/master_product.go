package publicHandlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/app/helpers"
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

	// Extract marketplace and offline filters using helper that understands multiple truthy values
	isMarketplaceFilter := helpers.QueryBool(c, "online")
	isOfflineFilter := helpers.QueryBool(c, "offline")

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

// GetProductByID handles retrieving a single product by its ID
func GetProductByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)

	if err != nil {
		handlers.SendError(c, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	product, err := db.FetchProductByID(id)

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
