package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard response structure for API responses
type APIResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	ErrorField string      `json:"error_field,omitempty"`
}

// PaginatedData is a standard response for paginated data
type PaginatedData struct {
	Items      interface{} `json:"items"`
	Page       int         `json:"page"`
	TotalPages int         `json:"total_pages"`
	TotalItems int         `json:"total_items"`
	Filters    interface{} `json:"filters,omitempty"`
	Sort       string      `json:"sort,omitempty"`
	Order      string      `json:"order,omitempty"`
}

// sendError sends a standardized error response
func sendError(c *gin.Context, statusCode int, message string, errorField *string) {
	c.JSON(statusCode, APIResponse{
		Success:    false,
		Error:      message,
		ErrorField: *errorField,
	})
}

// sendSuccess sends a standardized success response
func sendSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Helper function to respond with error
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// Helper function to respond with JSON
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
