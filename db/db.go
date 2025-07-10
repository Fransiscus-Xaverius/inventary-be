package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	settings "github.com/everysoft/inventary-be/settings" // Adjust based on your module name
	_ "github.com/lib/pq"                                 // PostgreSQL driver
)

var DB *sql.DB

func SetupDB(config *settings.Config) (*sql.DB, error) {
	// Create PostgreSQL connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host, config.Database.Port, config.Database.User,
		config.Database.Password, config.Database.DBName)

	// Open database connection
	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection is working
	err = DB.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return DB, nil
}

// InitDB initializes all database tables
func InitDB() error {
	// Create tables if they don't exist
	if err := CreateUsersTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	if err := CreateMasterProductsTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create master_products table: %w", err)
	}

	if err := CreateCategoryColorLabelsTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create category_color_labels table: %w", err)
	}

	if err := CreateMasterColorsTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create master_colors table: %w", err)
	}

	if err := CreateMasterGrupsTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create master_grups table: %w", err)
	}

	if err := CreateMasterUnitsTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create master_units table: %w", err)
	}

	if err := CreateMasterKatsTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create master_kats table: %w", err)
	}

	if err := CreateMasterGendersTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create master_genders table: %w", err)
	}

	if err := CreateMasterTipesTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create master_tipes table: %w", err)
	}

	if err := CreateBannersTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create banners table: %w", err)
	}

	log.Println("Database initialization completed successfully")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}

// RunSeeders executes all the seeder SQL files in the correct order
func RunSeeders() error {
	log.Println("Running seeders...")

	// Define all seeder files in execution order
	orderedSeeders := []struct {
		name string
		path string
	}{
		{name: "colors", path: "db/sql/seeder_master_colors.sql"},                        // First load base colors
		{name: "grups", path: "db/sql/seeder_master_grups.sql"},                          // Master grups
		{name: "units", path: "db/sql/seeder_master_units.sql"},                          // Master units
		{name: "kats", path: "db/sql/seeder_master_kats.sql"},                            // Master categories
		{name: "genders", path: "db/sql/seeder_master_genders.sql"},                      // Master genders
		{name: "tipes", path: "db/sql/seeder_master_tipes.sql"},                          // Master tipes
		{name: "banners", path: "db/sql/seeder_banners.sql"},                              // Master banners
		{name: "category_color_labels", path: "db/sql/seeder_category_color_labels.sql"}, // Then category colors
		{name: "products", path: "db/sql/seeder_master_products.sql"},                    // Finally load products with references
	}

	// Process each seeder in order
	for _, seeder := range orderedSeeders {
		log.Printf("Running %s seeder...", seeder.name)

		// Read the SQL file content
		content, err := os.ReadFile(seeder.path)
		if err != nil {
			return fmt.Errorf("failed to read seeder file %s: %w", seeder.path, err)
		}

		// Execute the SQL content
		_, err = DB.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute seeder %s: %w", seeder.path, err)
		}

		log.Printf("Successfully executed %s seeder", seeder.name)
	}

	log.Println("All seeders executed successfully")
	return nil
}

// RunSpecificSeeders executes only the specified seeder SQL files
func RunSpecificSeeders(seeders []string) error {
	log.Println("Running specific seeders...")

	// Define all seeder files
	allSeeders := map[string]string{
		"colors":                "db/sql/seeder_master_colors.sql",
		"category_color_labels": "db/sql/seeder_category_color_labels.sql",
		"grups":                 "db/sql/seeder_master_grups.sql",
		"units":                 "db/sql/seeder_master_units.sql",
		"kats":                  "db/sql/seeder_master_kats.sql",
		"genders":               "db/sql/seeder_master_genders.sql",
		"tipes":                 "db/sql/seeder_master_tipes.sql",
		"products":              "db/sql/seeder_master_products.sql",
	}

	// Validate requested seeders
	for _, name := range seeders {
		if _, exists := allSeeders[name]; !exists {
			return fmt.Errorf("unknown seeder: %s", name)
		}
	}

	// Run only the specified seeders
	for _, name := range seeders {
		file := allSeeders[name]
		log.Printf("Running %s seeder...", name)

		// Read the SQL file content
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read seeder file %s: %w", file, err)
		}

		// Execute the SQL content
		_, err = DB.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute seeder %s: %w", file, err)
		}

		log.Printf("Successfully executed %s seeder", name)
	}

	log.Println("Specified seeders executed successfully")
	return nil
}
