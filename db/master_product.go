package db

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func CreateMasterProductsTableIfNotExists() error {
	query := `
	CREATE TABLE IF NOT EXISTS master_products (
		no SERIAL PRIMARY KEY,
		artikel TEXT NOT NULL,
		warna TEXT,
		size TEXT,
		grup TEXT,
		unit TEXT,
		kat TEXT,
		model TEXT,
		gender TEXT,
		tipe TEXT,
		harga NUMERIC(15,2),
		tanggal_produk DATE,
		tanggal_terima DATE,
		usia INT,
		status TEXT,
		supplier TEXT,
		diupdate_oleh TEXT,
		tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_master_products_artikel ON master_products(artikel);
	CREATE INDEX IF NOT EXISTS idx_master_products_grup ON master_products(grup);
	ALTER TABLE master_products ADD CONSTRAINT IF NOT EXISTS uq_master_products_artikel UNIQUE (artikel);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create master_products table: %w", err)
	}

	log.Println("Ensured master_products table exists")
	return nil
}