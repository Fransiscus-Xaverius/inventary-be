package db

import (
	"database/sql"
	"fmt"
	"log"
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
	if err := CreateMasterProductsTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create master_products table: %w", err)
	}

	if err := CreateCategoryColorLabelsTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create category_color_labels table: %w", err)
	}

	if err := CreateMasterColorsTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to create master_colors table: %w", err)
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
