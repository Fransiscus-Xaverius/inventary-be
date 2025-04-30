package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/everysoft/inventary-be/app/handlers"
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
			auth.POST("/login", handlers.LoginHandler)      // to use gin.Context instead of http.HandlerFunc
		}

		
		products := api.Group("/products")
		products.Use(AuthMiddleware())
		{
			products.GET("/", handlers.GetAllProducts)
			products.POST("/", handlers.CreateProduct)
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