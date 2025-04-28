package routes

import (
	"net/http"
	"time"

	"github.com/everysoft/inventary-be/app/handlers"
)

// SetupRoutes configures all API routes
func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Public routes
	mux.HandleFunc("/api/auth/register", handlers.RegisterHandler)
	mux.HandleFunc("/api/auth/login", handlers.LoginHandler)
	
	// Public health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Protected routes - API endpoints that require authentication
	protectedMux := http.NewServeMux()
	
	// Example protected endpoint
	protectedMux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		handlers.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "You have access to protected content",
			"user_id": r.Header.Get("X-User-ID"),
			"role":    r.Header.Get("X-User-Role"),
		})
	})
	
	// Apply auth middleware to protected routes
	mux.Handle("/api/protected/", handlers.AuthMiddleware(http.StripPrefix("/api/protected", protectedMux)))
	
	return mux
}

// CreateServer creates a configured HTTP server
func CreateServer(port string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}