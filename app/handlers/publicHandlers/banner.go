package publicHandlers

import (
	"net/http"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// GetActiveBanners retrieves all active banners, ordered by order_index
func GetActiveBanners(c *gin.Context) {
	isActive := true
	banners, err := db.FetchAllBanners(100, 0, "", &isActive, "order_index", "asc") // Set a reasonable limit, e.g., 100
	if err != nil {
		handlers.SendError(c, http.StatusInternalServerError, "Failed to fetch active banners: "+err.Error(), nil)
		return
	}
	handlers.SendSuccess(c, http.StatusOK, gin.H{"banners": banners})
}
