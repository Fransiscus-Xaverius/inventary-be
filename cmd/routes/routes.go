package routes

import (
	"net/http"
	"time"

	"github.com/everysoft/inventary-be/app/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes() *gin.Engine {
	router := gin.Default()

	// Better CORS middleware configuration
	router.Use(CORSMiddleware())

	// Public routes group
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterHandler) // You'll need to update these handler functions
			auth.POST("/login", handlers.LoginHandler)       // to use gin.Context instead of http.HandlerFunc
		}

		/**
		 * Protected routes
		 * These routes require authentication
		 */
		products := api.Group("/products")
		products.Use(AuthMiddleware())
		{
			products.GET("/", handlers.GetAllProducts)
			products.POST("/", handlers.CreateProduct)
			products.GET("/deleted", handlers.GetDeletedProducts) // Route for fetching deleted products
			products.GET("/:artikel", handlers.GetProductByArtikel)

			products.PUT("/:artikel", handlers.UpdateProduct)
			products.DELETE("/:artikel", handlers.DeleteProduct)
			products.POST("/restore/:artikel", handlers.RestoreProduct) // Route for restoring deleted products
		}

		/**
		 * Master Colors routes
		 * These routes require authentication
		 */
		colors := api.Group("/colors")
		colors.Use(AuthMiddleware())
		{
			colors.GET("/", handlers.GetAllColors)
			colors.POST("/", handlers.CreateColor)
			colors.GET("/deleted", handlers.GetDeletedColors)
			colors.GET("/:id", handlers.GetColorByID)
			colors.PUT("/:id", handlers.UpdateColor)
			colors.DELETE("/:id", handlers.DeleteColor)
			colors.POST("/restore/:id", handlers.RestoreColor)
		}

		/**
		 * Master Sizes routes
		 * These routes require authentication
		 */
		sizes := api.Group("/sizes")
		sizes.Use(AuthMiddleware())
		{
			sizes.GET("/", handlers.GetAllSizes)
			sizes.POST("/", handlers.CreateSize)
			sizes.GET("/deleted", handlers.GetDeletedSizes)
			sizes.GET("/:id", handlers.GetSizeByID)
			sizes.PUT("/:id", handlers.UpdateSize)
			sizes.DELETE("/:id", handlers.DeleteSize)
			sizes.POST("/restore/:id", handlers.RestoreSize)
		}

		/**
		 * Accessible without authentication
		 */
		api.GET("/filters", handlers.GetFilterOptions)

		api.GET("/category-colors", handlers.GetAllCategoryColorLabels)
		api.GET("/category-colors/:column", handlers.GetCategoryColorLabelsByColumn)
	}

	return router
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// HealthCheck handler using Gin native syntax
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
		"time":   time.Now(),
	})
}

// GetProfile handler using Gin native syntax
func GetProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "You have access to protected content",
		"user_id": c.GetHeader("X-User-ID"),
		"role":    c.GetHeader("X-User-Role"),
	})
}

// AuthMiddleware using Gin native syntax
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		// Add your token validation logic here
		// You might want to set some values in the context
		// c.Set("user_id", userID)
		// c.Set("user_role", userRole)

		c.Next()
	}
}

// CreateServer creates a configured HTTP server
func CreateServer(port string, handler *gin.Engine) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}
