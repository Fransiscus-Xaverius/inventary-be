package handlers

import (
	"log"
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/everysoft/inventary-be/app/auth"
	"github.com/everysoft/inventary-be/db"
	"github.com/google/uuid"
)

// RegisterHandler converts to Gin handler
func RegisterHandler(c *gin.Context) {
	var registerRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Validate required fields
	if strings.TrimSpace(registerRequest.Username) == "" ||
		strings.TrimSpace(registerRequest.Password) == "" ||
		strings.TrimSpace(registerRequest.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username, password and email are required",
		})
		return
	}

	// Check if user exists
	exists, err := db.UserExists(registerRequest.Username, registerRequest.Email)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already exists",
		})
		return
	}

	// Create user
	hashedPassword, err := auth.HashPassword(registerRequest.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	user := auth.User{
		ID:       uuid.New().String(),
		Username: registerRequest.Username,
		Password: hashedPassword,
		Email:    registerRequest.Email,
		Role:     "user", // Default role
	}

	if err := db.CreateUser(user); err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	// Generate token
	token, expiresAt, err := auth.GenerateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"expires_at": expiresAt,
		"user": user,
	})
}

// LoginHandler converts to Gin handler
func LoginHandler(c *gin.Context) {
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Validate required fields
	if strings.TrimSpace(loginRequest.Username) == "" ||
		strings.TrimSpace(loginRequest.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username and password are required",
		})
		return
	}

	// Get user from database
	user, err := db.GetUserByUsername(loginRequest.Username)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Verify password
	if !auth.VerifyPassword(user.Password, loginRequest.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Generate token
	token, expiresAt, err := auth.GenerateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"expires_at": expiresAt,
		"user": user,
	})
}

// Optional: Helper middleware for authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "No token provided",
			})
			return
		}

		// Remove 'Bearer ' prefix if present
		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := auth.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
