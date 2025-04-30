// cmd/server/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"net/http"

	"github.com/everysoft/inventary-be/db"
	server "github.com/everysoft/inventary-be/cmd/routes"
	settings "github.com/everysoft/inventary-be/settings"
)

func main() {
	log.Printf("Starting YK InventaryBE service...")
	
	// Load configuration
	config, err := settings.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Setup database
	_, err = db.SetupDB(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.CloseDB()
	
	// Ensure database tables exist
	if err := db.CreateUsersTableIfNotExists(); err != nil {
		log.Fatalf("Failed to create database tables: %v", err)
	}

	if err = db.CreateMasterProductsTableIfNotExists(); err != nil {
		log.Fatalf("Failed to create database tables: %v", err)
	}
	
	// Setup routes - this will return *gin.Engine instead of *http.ServeMux
	router := server.SetupRoutes()
	
	// Create server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := server.CreateServer(port, router)
	
	// Run server in a goroutine so it doesn't block
	go func() {
		log.Printf("Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	
	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Doesn't block if no connections, but will otherwise wait until the timeout
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited properly")
}