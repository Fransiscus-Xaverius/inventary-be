// handlers/auth_handlers.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	
	"github.com/everysoft/inventary-be/app/auth"
	"github.com/everysoft/inventary-be/db"
	"github.com/google/uuid"
)

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Basic validation
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Email) == "" {
		RespondWithError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Check if user already exists
	exists, err := db.UserExists(req.Username, req.Email)
	if err != nil {
		log.Printf("Error checking if user exists: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Registration failed")
		return
	}
	
	if exists {
		RespondWithError(w, http.StatusConflict, "Username or email already in use")
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Registration failed")
		return
	}

	// Create user
	user := auth.User{
		ID:       uuid.New().String(),
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user", // Default role
	}

	// Save user to database
	if err := db.CreateUser(user); err != nil {
		log.Printf("Error creating user: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Registration failed")
		return
	}

	// Generate token
	token, expiresAt, err := auth.GenerateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Registration successful but token generation failed")
		return
	}

	// Prepare response
	response := auth.TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: auth.User{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	// Return response
	RespondWithJSON(w, http.StatusCreated, response)
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Retrieve user from database
	user, err := db.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("Login attempt failed for %s: %v", req.Username, err)
		RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verify password
	if !auth.VerifyPassword(user.Password, req.Password) {
		log.Printf("Invalid password for user %s", req.Username)
		RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate token
	token, expiresAt, err := auth.GenerateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Authentication successful but token generation failed")
		return
	}

	// Set cookie for web clients (optional)
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil, // Only send over HTTPS if available
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	// Prepare response
	response := auth.TokenResponse{
		Token:     token, 
		ExpiresAt: expiresAt,
		User: auth.User{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	// Return response
	RespondWithJSON(w, http.StatusOK, response)
}

// AuthMiddleware is a middleware to authenticate requests
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from request
		tokenString := auth.ExtractTokenFromRequest(r)
		if tokenString == "" {
			RespondWithError(w, http.StatusUnauthorized, "Authorization token required")
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Attach user info to request context
		// You can implement context.WithValue here to add user info
		// For simplicity, we'll just add it to request headers
		r.Header.Set("X-User-ID", claims.UserID)
		r.Header.Set("X-User-Role", claims.Role)

		// Proceed to next handler
		next.ServeHTTP(w, r)
	})
}