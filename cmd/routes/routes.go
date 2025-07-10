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

	// Set a higher limit for multipart forms (e.g., 8MB)
	router.MaxMultipartMemory = 8 << 20  // 8MB

	// Serve static files from the 'uploads' directory
	router.Static("/uploads", "uploads")

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
		 * Master Grup routes
		 * These routes require authentication
		 */
		grups := api.Group("/grups")
		grups.Use(AuthMiddleware())
		{
			grups.GET("/", handlers.GetAllGrups)
			grups.POST("/", handlers.CreateGrup)
			grups.GET("/deleted", handlers.GetDeletedGrups)
			grups.GET("/:id", handlers.GetGrupByID)
			grups.PUT("/:id", handlers.UpdateGrup)
			grups.DELETE("/:id", handlers.DeleteGrup)
			grups.POST("/restore/:id", handlers.RestoreGrup)
		}

		/**
		 * Master Unit routes
		 * These routes require authentication
		 */
		units := api.Group("/units")
		units.Use(AuthMiddleware())
		{
			units.GET("/", handlers.GetAllUnits)
			units.POST("/", handlers.CreateUnit)
			units.GET("/deleted", handlers.GetDeletedUnits)
			units.GET("/:id", handlers.GetUnitByID)
			units.PUT("/:id", handlers.UpdateUnit)
			units.DELETE("/:id", handlers.DeleteUnit)
			units.POST("/restore/:id", handlers.RestoreUnit)
		}

		/**
		 * Master Kat routes
		 * These routes require authentication
		 */
		kats := api.Group("/kats")
		kats.Use(AuthMiddleware())
		{
			kats.GET("/", handlers.GetAllKats)
			kats.POST("/", handlers.CreateKat)
			kats.GET("/deleted", handlers.GetDeletedKats)
			kats.GET("/:id", handlers.GetKatByID)
			kats.PUT("/:id", handlers.UpdateKat)
			kats.DELETE("/:id", handlers.DeleteKat)
			kats.POST("/restore/:id", handlers.RestoreKat)
		}

		/**
		 * Master Gender routes
		 * These routes require authentication
		 */
		genders := api.Group("/genders")
		genders.Use(AuthMiddleware())
		{
			genders.GET("/", handlers.GetAllGenders)
			genders.POST("/", handlers.CreateGender)
			genders.GET("/deleted", handlers.GetDeletedGenders)
			genders.GET("/:id", handlers.GetGenderByID)
			genders.PUT("/:id", handlers.UpdateGender)
			genders.DELETE("/:id", handlers.DeleteGender)
			genders.POST("/restore/:id", handlers.RestoreGender)
		}

		/**
		 * Master Tipe routes
		 * These routes require authentication
		 */
		tipes := api.Group("/tipes")
		tipes.Use(AuthMiddleware())
		{
			tipes.GET("/", handlers.GetAllTipes)
			tipes.POST("/", handlers.CreateTipe)
			tipes.GET("/deleted", handlers.GetDeletedTipes)
			tipes.GET("/:id", handlers.GetTipeByID)
			tipes.PUT("/:id", handlers.UpdateTipe)
			tipes.DELETE("/:id", handlers.DeleteTipe)
			tipes.POST("/restore/:id", handlers.RestoreTipe)
		}

		/**
		 * Accessible without authentication
		 */
		api.GET("/filters", handlers.GetFilterOptions)
		api.GET("/category-colors", handlers.GetAllCategoryColorLabels)
		api.GET("/category-colors/:column", handlers.GetCategoryColorLabelsByColumn)
		api.GET("/category-colors/:column/:value", handlers.GetCategoryColorLabelByColumnAndValue)

		/**
		 * Banners routes
		 * These routes require authentication for management, but public for fetching active banners
		 */
		banners := api.Group("/banners")
		{
			banners.GET("/active", handlers.GetActiveBanners) // Public endpoint for active banners
		}

		// Protected banner routes
		bannersProtected := api.Group("/banners")
		bannersProtected.Use(AuthMiddleware())
		{
			bannersProtected.GET("/", handlers.GetAllBanners)
			bannersProtected.POST("/", handlers.CreateBanner)
			bannersProtected.GET("/deleted", handlers.GetDeletedBanners)
			bannersProtected.GET("/:id", handlers.GetBannerByID)
			bannersProtected.PUT("/:id", handlers.UpdateBanner)
			bannersProtected.DELETE("/:id", handlers.DeleteBanner)
			bannersProtected.POST("/restore/:id", handlers.RestoreBanner)
		}
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
