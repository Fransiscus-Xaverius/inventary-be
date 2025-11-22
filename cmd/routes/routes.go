package routes

import (
	"net/http"
	"time"

	"github.com/everysoft/inventary-be/app/handlers/adminHandlers"
	"github.com/everysoft/inventary-be/app/handlers/publicHandlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes() *gin.Engine {
	router := gin.Default()

	// Set a higher limit for multipart forms (e.g., 8MB)
	router.MaxMultipartMemory = 100 << 20 // 100MB

	// Serve static files from the 'uploads' directory
	router.Static("/uploads", "uploads")

	// Better CORS middleware configuration
	router.Use(CORSMiddleware())

	// Public routes group
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			// auth.POST("/register", publicHandlers.RegisterHandler) // You'll need to update these handler functions
			auth.POST("/login", publicHandlers.LoginHandler) // to use gin.Context instead of http.HandlerFunc
		}

		/**
		 * Public routes
		 */
		api.GET("/filters", publicHandlers.GetFilterOptions)
		api.GET("/category-colors", publicHandlers.GetAllCategoryColorLabels)
		api.GET("/category-colors/:column", publicHandlers.GetCategoryColorLabelsByColumn)
		api.GET("/category-colors/:column/:value", publicHandlers.GetCategoryColorLabelByColumnAndValue)

		/**
		 * Products routes
		 */
		products := api.Group("/products")
		{
			products.GET("", publicHandlers.GetAllProducts)
			products.GET("/:artikel", publicHandlers.GetProductByArtikel)
		}

		/**
		 * Banners routes
		 * These routes require authentication for management, but public for fetching active banners
		 */
		banners := api.Group("/banners")
		{
			banners.GET("/active", publicHandlers.GetActiveBanners) // Public endpoint for active banners
		}

		/**
		 * Newsletter routes
		 */
		api.POST("/newsletter", publicHandlers.SubscribeToNewsletter)

		/**
		 * Admin routes
		 * These routes require authentication
		 */
		admin := api.Group("/admin")
		admin.Use(AuthMiddleware())
		{
			/**
			 * Master Products routes
			 * These routes require authentication
			 */
			productsProtected := admin.Group("/products")
			{
				productsProtected.GET("", adminHandlers.GetAllProducts)
				productsProtected.POST("", adminHandlers.CreateProduct)
				productsProtected.GET("/deleted", adminHandlers.GetDeletedProducts) // Route for fetching deleted products
				productsProtected.GET("/:artikel", adminHandlers.GetProductByArtikel)
				productsProtected.PUT("/:artikel", adminHandlers.UpdateProduct)
				productsProtected.DELETE("/:artikel", adminHandlers.DeleteProduct)
				productsProtected.POST("/restore/:artikel", adminHandlers.RestoreProduct) // Route for restoring deleted products
			}

			/**
			 * Master Colors routes
			 * These routes require authentication
			 */
			colorsProtected := admin.Group("/colors")
			{
				colorsProtected.GET("", adminHandlers.GetAllColors)
				colorsProtected.POST("", adminHandlers.CreateColor)
				colorsProtected.GET("/deleted", adminHandlers.GetDeletedColors)
				colorsProtected.GET("/:id", adminHandlers.GetColorByID)
				colorsProtected.PUT("/:id", adminHandlers.UpdateColor)
				colorsProtected.DELETE("/:id", adminHandlers.DeleteColor)
				colorsProtected.POST("/restore/:id", adminHandlers.RestoreColor)
			}

			/**
			 * Master Grup routes
			 * These routes require authentication
			 */
			grupsProtected := admin.Group("/grups")
			{
				grupsProtected.GET("", adminHandlers.GetAllGrups)
				grupsProtected.POST("", adminHandlers.CreateGrup)
				grupsProtected.GET("/deleted", adminHandlers.GetDeletedGrups)
				grupsProtected.GET("/:id", adminHandlers.GetGrupByID)
				grupsProtected.PUT("/:id", adminHandlers.UpdateGrup)
				grupsProtected.DELETE("/:id", adminHandlers.DeleteGrup)
				grupsProtected.POST("/restore/:id", adminHandlers.RestoreGrup)
			}

			/**
			 * Master Unit routes
			 * These routes require authentication
			 */
			unitsProtected := admin.Group("/units")
			{
				unitsProtected.GET("", adminHandlers.GetAllUnits)
				unitsProtected.POST("", adminHandlers.CreateUnit)
				unitsProtected.GET("/deleted", adminHandlers.GetDeletedUnits)
				unitsProtected.GET("/:id", adminHandlers.GetUnitByID)
				unitsProtected.PUT("/:id", adminHandlers.UpdateUnit)
				unitsProtected.DELETE("/:id", adminHandlers.DeleteUnit)
				unitsProtected.POST("/restore/:id", adminHandlers.RestoreUnit)
			}

			/**
			 * Master Kat routes
			 * These routes require authentication
			 */
			katsProtected := admin.Group("/kats")
			{
				katsProtected.GET("", adminHandlers.GetAllKats)
				katsProtected.POST("", adminHandlers.CreateKat)
				katsProtected.GET("/deleted", adminHandlers.GetDeletedKats)
				katsProtected.GET("/:id", adminHandlers.GetKatByID)
				katsProtected.PUT("/:id", adminHandlers.UpdateKat)
				katsProtected.DELETE("/:id", adminHandlers.DeleteKat)
				katsProtected.POST("/restore/:id", adminHandlers.RestoreKat)
			}

			/**
			 * Master Gender routes
			 * These routes require authentication
			 */
			gendersProtected := admin.Group("/genders")
			{
				gendersProtected.GET("", adminHandlers.GetAllGenders)
				gendersProtected.POST("", adminHandlers.CreateGender)
				gendersProtected.GET("/deleted", adminHandlers.GetDeletedGenders)
				gendersProtected.GET("/:id", adminHandlers.GetGenderByID)
				gendersProtected.PUT("/:id", adminHandlers.UpdateGender)
				gendersProtected.DELETE("/:id", adminHandlers.DeleteGender)
				gendersProtected.POST("/restore/:id", adminHandlers.RestoreGender)
			}

			/**
			 * Master Tipe routes
			 * These routes require authentication
			 */
			tipesProtected := admin.Group("/tipes")
			{
				tipesProtected.GET("", adminHandlers.GetAllTipes)
				tipesProtected.POST("", adminHandlers.CreateTipe)
				tipesProtected.GET("/deleted", adminHandlers.GetDeletedTipes)
				tipesProtected.GET("/:id", adminHandlers.GetTipeByID)
				tipesProtected.PUT("/:id", adminHandlers.UpdateTipe)
				tipesProtected.DELETE("/:id", adminHandlers.DeleteTipe)
				tipesProtected.POST("/restore/:id", adminHandlers.RestoreTipe)
			}

			/**
			 * Master Banners routes
			 * These routes require authentication
			 */
			bannersProtected := admin.Group("/banners")
			{
				bannersProtected.GET("", adminHandlers.GetAllBanners)
				bannersProtected.POST("", adminHandlers.CreateBanner)
				bannersProtected.GET("/deleted", adminHandlers.GetDeletedBanners)
				bannersProtected.GET("/:id", adminHandlers.GetBannerByID)
				bannersProtected.PUT("/:id", adminHandlers.UpdateBanner)
				bannersProtected.DELETE("/:id", adminHandlers.DeleteBanner)
				bannersProtected.POST("/restore/:id", adminHandlers.RestoreBanner)
			}

			/**
			 * Panduan Ukuran routes
			 * These routes require authentication
			 */
			panduanUkuranProtected := admin.Group("/panduan-ukuran")
			{
				panduanUkuranProtected.POST("", adminHandlers.UploadPanduanUkuran)
				panduanUkuranProtected.DELETE("", adminHandlers.DeletePanduanUkuran)
			}

			/**
			 * Master Newsletter routes
			 * These routes require authentication
			 */
			newslettersProtected := admin.Group("/newsletters")
			{
				newslettersProtected.GET("", adminHandlers.GetAllNewsletters)
				newslettersProtected.GET("/deleted", adminHandlers.GetDeletedNewsletters)
				newslettersProtected.GET("/:id", adminHandlers.GetNewsletterByID)
				newslettersProtected.PUT("/:id", adminHandlers.UpdateNewsletter)
				newslettersProtected.DELETE("/:id", adminHandlers.DeleteNewsletter)
				newslettersProtected.POST("/restore/:id", adminHandlers.RestoreNewsletter)
			}
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
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Multipart", "true")

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
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}
