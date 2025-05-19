// cmd/server/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	server "github.com/everysoft/inventary-be/cmd/routes"
	"github.com/everysoft/inventary-be/db"
	settings "github.com/everysoft/inventary-be/settings"
)

// printSeedHelp prints detailed help about available seeders
func printSeedHelp() {
	fmt.Println("\nAvailable seeders:")
	fmt.Println("  colors                - Populates master_colors table with common colors in Bahasa Indonesia")
	fmt.Println("  grups                 - Populates master_grups table with common groups in Bahasa Indonesia")
	fmt.Println("  units                 - Populates master_units table with common units in Bahasa Indonesia")
	fmt.Println("  kats                  - Populates master_kats table with common categories in Bahasa Indonesia")
	fmt.Println("  genders               - Populates master_genders table with common genders in Bahasa Indonesia")
	fmt.Println("  tipes                 - Populates master_tipes table with common types in Bahasa Indonesia")
	fmt.Println("  category_color_labels - Adds color labels for category attributes")
	fmt.Println("  products              - Adds/updates 1000 sample products in master_products table")
	fmt.Println("\nExamples:")
	fmt.Println("  ./main -seed                               # Run all seeders")
	fmt.Println("  ./main -seed-specific colors,products,sizes # Run only colors, products and sizes seeders")
	fmt.Println("")
}

func main() {
	// Parse command line flags
	seedFlag := flag.Bool("seed", false, "Run all database seeders and exit")
	seedSpecific := flag.String("seed-specific", "", "Run specific seeders (comma-separated: colors,category_color_labels,products) and exit")
	seedHelp := flag.Bool("seed-help", false, "Show information about available seeders")
	flag.Parse()

	// Show seeder help if requested
	if *seedHelp {
		printSeedHelp()
		return
	}

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

	// Initialize database tables
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Handle seeding
	if *seedFlag {
		log.Println("Seeding all tables...")
		if err := db.RunSeeders(); err != nil {
			log.Fatalf("Failed to run seeders: %v", err)
		}
		log.Println("Database seeding completed successfully")
		return
	}

	// Handle specific seeding
	if *seedSpecific != "" {
		specificSeeders := strings.Split(*seedSpecific, ",")
		for i := range specificSeeders {
			specificSeeders[i] = strings.TrimSpace(specificSeeders[i])
		}

		log.Printf("Seeding specific tables: %v", specificSeeders)
		if err := db.RunSpecificSeeders(specificSeeders); err != nil {
			log.Fatalf("Failed to run specific seeders: %v", err)
		}
		log.Println("Specific database seeding completed successfully")
		return
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
