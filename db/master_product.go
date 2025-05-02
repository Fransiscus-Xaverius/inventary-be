package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/everysoft/inventary-be/app/master_product"
)

func CountAllProducts(queryStr string, filters map[string]string) (int, error) {
	baseQuery := "SELECT COUNT(no) FROM master_products WHERE 1=1"
	args := []interface{}{}
	paramCount := 1
	
	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			artikel ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR warna ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR size ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR grup ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR unit ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR kat ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR model ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR gender ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR tipe ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR CAST(harga AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR CAST(no AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR CAST(usia AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR status ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR supplier ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR diupdate_oleh ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}
	
	// Add filter conditions for each field
	validFilterFields := []string{"warna", "size", "grup", "unit", "kat", "model", "gender", "tipe", "status", "supplier"}
	for _, field := range validFilterFields {
		if value, ok := filters[field]; ok && value != "" {
			filterQuery := ` AND ` + field + ` = $` + fmt.Sprintf("%d", paramCount)
			baseQuery += filterQuery
			args = append(args, value)
			paramCount++
		}
	}
	
	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

func CreateMasterProductsTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_products (
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
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_products_artikel ON master_products(artikel);`,
		`CREATE INDEX IF NOT EXISTS idx_master_products_grup ON master_products(grup);`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'uq_master_products_artikel'
			) THEN
				ALTER TABLE master_products ADD CONSTRAINT uq_master_products_artikel UNIQUE (artikel);
			END IF;
		END$$;`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_products table exists")
	return nil
}

func FetchAllProducts(limit, offset int, queryStr string, filters map[string]string, sortColumn string, sortDirection string) ([]master_product.Product, error) {
	products := []master_product.Product{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		no, artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, usia, status, supplier, diupdate_oleh, tanggal_update
	FROM master_products
	WHERE 1=1`
	
	args := []interface{}{}
	paramCount := 1
	
	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			artikel ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR warna ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR size ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR grup ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR unit ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR kat ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR model ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR gender ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR tipe ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR CAST(harga AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR CAST(no AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR CAST(usia AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR status ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR supplier ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR diupdate_oleh ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}
	
	// Add filter conditions for each field
	validFilterFields := []string{"warna", "size", "grup", "unit", "kat", "model", "gender", "tipe", "status", "supplier"}
	for _, field := range validFilterFields {
		if value, ok := filters[field]; ok && value != "" {
			filterQuery := ` AND ` + field + ` = $` + fmt.Sprintf("%d", paramCount)
			baseQuery += filterQuery
			args = append(args, value)
			paramCount++
		}
	}
	
	// Add sorting
	orderBy := " ORDER BY "
	// Map of valid column names to prevent SQL injection
	validColumns := map[string]bool{
		"no": true, "artikel": true, "warna": true, "size": true, "grup": true,
		"unit": true, "kat": true, "model": true, "gender": true, "tipe": true,
		"harga": true, "tanggal_produk": true, "tanggal_terima": true, "usia": true,
		"status": true, "supplier": true, "diupdate_oleh": true, "tanggal_update": true,
	}
	
	// Default sort
	if sortColumn == "" || !validColumns[sortColumn] {
		sortColumn = "no"
	}
	
	// Default direction
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "asc"
	}
	
	orderBy += sortColumn + " " + sortDirection
	
	// Add pagination
	paginationQuery := orderBy + ` LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	baseQuery += paginationQuery
	args = append(args, limit, offset)

	// Execute the query with all parameters
	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p master_product.Product
		if err := rows.Scan(
			&p.No, &p.Artikel, &p.Warna, &p.Size, &p.Grup, &p.Unit, &p.Kat,
			&p.Model, &p.Gender, &p.Tipe, &p.Harga, &p.TanggalProduk,
			&p.TanggalTerima, &p.Usia, &p.Status, &p.Supplier,
			&p.DiupdateOleh, &p.TanggalUpdate,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func FetchProductByArtikel(artikel string) (master_product.Product, error) {
	var p master_product.Product
	err := DB.QueryRow(`SELECT no, artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, usia, status, supplier, diupdate_oleh, tanggal_update FROM master_products WHERE artikel = $1`, artikel).
		Scan(&p.No, &p.Artikel, &p.Warna, &p.Size, &p.Grup, &p.Unit, &p.Kat, &p.Model, &p.Gender, &p.Tipe, &p.Harga, &p.TanggalProduk, &p.TanggalTerima, &p.Usia, &p.Status, &p.Supplier, &p.DiupdateOleh, &p.TanggalUpdate)

	if err == sql.ErrNoRows {
		return p, errors.New("not_found")
	}
	return p, err
}

func InsertProduct(p *master_product.Product) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_products 
		(artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, usia, status, supplier, diupdate_oleh, tanggal_update) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING no`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(
		p.Artikel,
		p.Warna,
		p.Size,
		p.Grup,
		p.Unit,
		p.Kat,
		p.Model,
		p.Gender,
		p.Tipe,
		p.Harga,
		p.TanggalProduk,
		p.TanggalTerima,
		p.Usia,
		p.Status,
		p.Supplier,
		p.DiupdateOleh,
		p.TanggalUpdate,
	).Scan(&p.No)
}

func UpdateProduct(artikel string, p *master_product.Product) (master_product.Product, error) {
	result, err := DB.Exec(`
		UPDATE master_products SET
			warna = $1,
			size = $2,
			grup = $3,
			unit = $4,
			kat = $5,
			model = $6,
			gender = $7,
			tipe = $8,
			harga = $9,
			tanggal_produk = $10,
			tanggal_terima = $11,
			usia = $12,
			status = $13,
			supplier = $14,
			diupdate_oleh = $15,
			tanggal_update = $16
		WHERE artikel = $17`,
		p.Warna, p.Size, p.Grup, p.Unit, p.Kat, p.Model, p.Gender, p.Tipe, p.Harga,
		p.TanggalProduk, p.TanggalTerima, p.Usia, p.Status, p.Supplier,
		p.DiupdateOleh, p.TanggalUpdate, artikel)

	if err != nil {
		return *p, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return *p, err
	}
	if rowsAffected == 0 {
		return *p, errors.New("not_found")
	}

	// Fetch updated product
	return FetchProductByArtikel(artikel)
}

func DeleteProduct(artikel string) error {
	result, err := DB.Exec("DELETE FROM master_products WHERE artikel = $1", artikel)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("not_found")
	}

	return nil
}

// FetchFilterOptions retrieves all unique values for specified fields
func FetchFilterOptions() (map[string]map[string]interface{}, error) {
	// Define the fields we want to get unique values for
	fields := []string{
		"warna", "size", "grup", "unit", "kat", "model", 
		"gender", "tipe", "status", "supplier", "diupdate_oleh",
	}
	
	result := make(map[string]map[string]interface{})
	
	// For each field, get its unique values
	for _, field := range fields {
		// Create an inner map to store the values array and total count
		fieldData := make(map[string]interface{})
		values := []string{}
		
		// Get unique values for this field, excluding NULL/empty values
		query := fmt.Sprintf("SELECT DISTINCT %s FROM master_products WHERE %s IS NOT NULL AND %s != '' ORDER BY %s", 
			field, field, field, field)
		
		rows, err := DB.Query(query)
		if err != nil {
			return nil, fmt.Errorf("error querying unique values for %s: %w", field, err)
		}
		defer rows.Close()
		
		// Collect all unique values
		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				return nil, fmt.Errorf("error scanning value for %s: %w", field, err)
			}
			values = append(values, value)
		}
		
		// Add the values and count to the field data
		fieldData["values"] = values
		fieldData["total"] = len(values)
		
		// Add this field to the result
		result[field] = fieldData
	}
	
	return result, nil
}
