package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/everysoft/inventary-be/app/models"
)

func CountAllProducts(queryStr string, filters map[string]string) (int, error) {
	baseQuery := "SELECT COUNT(no) FROM master_products WHERE tanggal_hapus IS NULL"
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
			OR CASE 
				WHEN tanggal_terima IS NOT NULL THEN 
					CASE
						WHEN (CURRENT_DATE - tanggal_terima) < 365 THEN 'Fresh'
						WHEN (CURRENT_DATE - tanggal_terima) < 730 THEN 'Normal'
						ELSE 'Aging'
					END
				ELSE 'Unknown' 
			   END ILIKE $` + fmt.Sprintf("%d", paramCount) + `
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
			status TEXT,
			supplier TEXT,
			diupdate_oleh TEXT,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			tanggal_hapus TIMESTAMPTZ
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

func FetchAllProducts(limit, offset int, queryStr string, filters map[string]string, sortColumn string, sortDirection string) ([]models.Product, error) {
	products := []models.Product{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		no, artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, 
		CASE 
			WHEN tanggal_terima IS NOT NULL THEN 
				CASE
					WHEN (CURRENT_DATE - tanggal_terima) < 365 THEN 'Fresh'
					WHEN (CURRENT_DATE - tanggal_terima) < 730 THEN 'Normal'
					ELSE 'Aging'
				END
			ELSE 'Unknown' 
		END AS usia,
		status, supplier, diupdate_oleh, tanggal_update, tanggal_hapus
	FROM master_products
	WHERE tanggal_hapus IS NULL`

	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			artikel ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR size ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR grup ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR unit ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR kat ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR model ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR gender ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR tipe ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR CAST(harga AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR CAST(no AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR CASE 
				WHEN tanggal_terima IS NOT NULL THEN 
					CASE
						WHEN (CURRENT_DATE - tanggal_terima) < 365 THEN 'Fresh'
						WHEN (CURRENT_DATE - tanggal_terima) < 730 THEN 'Normal'
						ELSE 'Aging'
					END
				ELSE 'Unknown' 
			   END ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR status ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR supplier ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR diupdate_oleh ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR EXISTS (
				SELECT 1 FROM master_colors mc
				WHERE (mc.nama ILIKE $` + fmt.Sprintf("%d", paramCount) + `)
				AND CAST(mc.id AS TEXT) IN (
					SELECT unnest(string_to_array(master_products.warna, ','))
				)
			)
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add filter conditions for each field
	validFilterFields := []string{"grup", "unit", "kat", "model", "gender", "tipe", "status", "supplier"}
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
		var p models.Product
		var usia string
		if err := rows.Scan(
			&p.No, &p.Artikel, &p.Warna, &p.Size, &p.Grup, &p.Unit, &p.Kat,
			&p.Model, &p.Gender, &p.Tipe, &p.Harga, &p.TanggalProduk,
			&p.TanggalTerima, &usia, &p.Status, &p.Supplier,
			&p.DiupdateOleh, &p.TanggalUpdate, &p.TanggalHapus,
		); err != nil {
			return nil, err
		}

		p.Usia = usia

		// Fetch color information
		colors, err := FetchColorsByIDs(p.Warna)
		if err != nil {
			log.Printf("Error fetching colors for product %s: %v", p.Artikel, err)
		} else {
			p.Colors = colors
		}

		products = append(products, p)
	}

	return products, nil
}

func FetchProductByArtikel(artikel string) (models.Product, error) {
	var p models.Product
	var usia string
	err := DB.QueryRow(`
		SELECT 
			no, artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, 
			tanggal_produk, tanggal_terima, 
			CASE 
				WHEN tanggal_terima IS NOT NULL THEN 
					CASE
						WHEN (CURRENT_DATE - tanggal_terima) < 365 THEN 'Fresh'
						WHEN (CURRENT_DATE - tanggal_terima) < 730 THEN 'Normal'
						ELSE 'Aging'
					END
				ELSE 'Unknown' 
			END AS usia,
			status, supplier, diupdate_oleh, tanggal_update, tanggal_hapus 
		FROM master_products 
		WHERE artikel = $1 AND tanggal_hapus IS NULL
	`, artikel).
		Scan(&p.No, &p.Artikel, &p.Warna, &p.Size, &p.Grup, &p.Unit, &p.Kat, &p.Model, &p.Gender, &p.Tipe, &p.Harga, &p.TanggalProduk, &p.TanggalTerima, &usia, &p.Status, &p.Supplier, &p.DiupdateOleh, &p.TanggalUpdate, &p.TanggalHapus)

	if err == sql.ErrNoRows {
		return p, errors.New("not_found")
	}

	if err != nil {
		return p, err
	}

	p.Usia = usia

	// Fetch color information
	colors, err := FetchColorsByIDs(p.Warna)
	if err != nil {
		log.Printf("Error fetching colors for product %s: %v", p.Artikel, err)
	} else {
		p.Colors = colors
	}

	return p, nil
}

func FetchProductByArtikelIncludeDeleted(artikel string) (models.Product, error) {
	var p models.Product
	var usia string
	err := DB.QueryRow(`
		SELECT 
			no, artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, 
			tanggal_produk, tanggal_terima, 
			CASE 
				WHEN tanggal_terima IS NOT NULL THEN 
					CASE
						WHEN (CURRENT_DATE - tanggal_terima) < 365 THEN 'Fresh'
						WHEN (CURRENT_DATE - tanggal_terima) < 730 THEN 'Normal'
						ELSE 'Aging'
					END
				ELSE 'Unknown' 
			END AS usia,
			status, supplier, diupdate_oleh, tanggal_update, tanggal_hapus 
		FROM master_products 
		WHERE artikel = $1
	`, artikel).
		Scan(&p.No, &p.Artikel, &p.Warna, &p.Size, &p.Grup, &p.Unit, &p.Kat, &p.Model, &p.Gender, &p.Tipe, &p.Harga, &p.TanggalProduk, &p.TanggalTerima, &usia, &p.Status, &p.Supplier, &p.DiupdateOleh, &p.TanggalUpdate, &p.TanggalHapus)

	if err == sql.ErrNoRows {
		return p, errors.New("not_found")
	}

	if err != nil {
		return p, err
	}

	p.Usia = usia

	// Fetch color information
	colors, err := FetchColorsByIDs(p.Warna)
	if err != nil {
		log.Printf("Error fetching colors for product %s: %v", p.Artikel, err)
	} else {
		p.Colors = colors
	}

	return p, nil
}

func InsertProduct(p *models.Product) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_products 
		(artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, status, supplier, diupdate_oleh, tanggal_update) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
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
		p.Status,
		p.Supplier,
		p.DiupdateOleh,
		p.TanggalUpdate,
	).Scan(&p.No)
}

func UpdateProduct(artikel string, p *models.Product) (models.Product, error) {
	// First fetch the existing product to get current values
	currentProduct, err := FetchProductByArtikel(artikel)
	if err != nil {
		return *p, err
	}

	// Build dynamic query with only fields that need to be updated
	query := "UPDATE master_products SET"
	args := []interface{}{}
	paramCount := 1
	fieldsToUpdate := 0

	// Process string fields
	stringFields := map[string]string{
		"warna":         p.Warna,
		"size":          p.Size,
		"grup":          p.Grup,
		"unit":          p.Unit,
		"kat":           p.Kat,
		"model":         p.Model,
		"gender":        p.Gender,
		"tipe":          p.Tipe,
		"status":        p.Status,
		"supplier":      p.Supplier,
		"diupdate_oleh": p.DiupdateOleh,
	}

	for field, value := range stringFields {
		if value != "" {
			fieldsToUpdate++
			query += fmt.Sprintf(" %s = $%d,", field, paramCount)
			args = append(args, value)
			paramCount++
		}
	}

	// Process numeric field
	if p.Harga != 0 {
		fieldsToUpdate++
		query += fmt.Sprintf(" harga = $%d,", paramCount)
		args = append(args, p.Harga)
		paramCount++
	}

	// Process date fields
	zeroTime := time.Time{}
	dateFields := map[string]time.Time{
		"tanggal_produk": p.TanggalProduk,
		"tanggal_terima": p.TanggalTerima,
	}

	for field, value := range dateFields {
		if value != zeroTime {
			fieldsToUpdate++
			query += fmt.Sprintf(" %s = $%d,", field, paramCount)
			args = append(args, value)
			paramCount++
		}
	}

	// Always update tanggal_update
	fieldsToUpdate++
	query += fmt.Sprintf(" tanggal_update = $%d,", paramCount)
	args = append(args, p.TanggalUpdate)
	paramCount++

	// Remove the trailing comma and complete the query
	query = query[:len(query)-1] + " WHERE artikel = $" + fmt.Sprintf("%d", paramCount)
	args = append(args, artikel)

	// If no fields to update, return the current product
	if fieldsToUpdate == 1 { // Only tanggal_update was added
		return currentProduct, nil
	}

	// Execute the query
	result, err := DB.Exec(query, args...)
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
	// Soft delete by setting tanggal_hapus to the current time
	currentTime := time.Now()
	result, err := DB.Exec("UPDATE master_products SET tanggal_hapus = $1 WHERE artikel = $2 AND tanggal_hapus IS NULL", currentTime, artikel)
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

// RestoreProduct restores a soft-deleted product by setting tanggal_hapus to NULL
func RestoreProduct(artikel string) error {
	result, err := DB.Exec("UPDATE master_products SET tanggal_hapus = NULL WHERE artikel = $1 AND tanggal_hapus IS NOT NULL", artikel)
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

		// Get unique values for this field, excluding NULL/empty values and soft-deleted records
		query := fmt.Sprintf("SELECT DISTINCT %s FROM master_products WHERE %s IS NOT NULL AND %s != '' AND tanggal_hapus IS NULL ORDER BY %s",
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

// FetchDeletedProducts retrieves all soft-deleted products with pagination
func FetchDeletedProducts(limit, offset int, queryStr string, filters map[string]string, sortColumn string, sortDirection string) ([]models.Product, error) {
	products := []models.Product{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		no, artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, 
		CASE 
			WHEN tanggal_terima IS NOT NULL THEN 
				CASE
					WHEN (CURRENT_DATE - tanggal_terima) < 365 THEN 'Fresh'
					WHEN (CURRENT_DATE - tanggal_terima) < 730 THEN 'Normal'
					ELSE 'Aging'
				END
			ELSE 'Unknown' 
		END AS usia,
		status, supplier, diupdate_oleh, tanggal_update, tanggal_hapus
	FROM master_products
	WHERE tanggal_hapus IS NOT NULL`

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
			OR CASE 
				WHEN tanggal_terima IS NOT NULL THEN 
					CASE
						WHEN (CURRENT_DATE - tanggal_terima) < 365 THEN 'Fresh'
						WHEN (CURRENT_DATE - tanggal_terima) < 730 THEN 'Normal'
						ELSE 'Aging'
					END
				ELSE 'Unknown' 
			   END ILIKE $` + fmt.Sprintf("%d", paramCount) + `
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
		"tanggal_hapus": true,
	}

	// Default sort
	if sortColumn == "" || !validColumns[sortColumn] {
		sortColumn = "tanggal_hapus"
	}

	// Default direction
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
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
		var p models.Product
		var usia string
		if err := rows.Scan(
			&p.No, &p.Artikel, &p.Warna, &p.Size, &p.Grup, &p.Unit, &p.Kat,
			&p.Model, &p.Gender, &p.Tipe, &p.Harga, &p.TanggalProduk,
			&p.TanggalTerima, &usia, &p.Status, &p.Supplier,
			&p.DiupdateOleh, &p.TanggalUpdate, &p.TanggalHapus,
		); err != nil {
			return nil, err
		}

		p.Usia = usia

		// Fetch color information
		colors, err := FetchColorsByIDs(p.Warna)
		if err != nil {
			log.Printf("Error fetching colors for product %s: %v", p.Artikel, err)
		} else {
			p.Colors = colors
		}

		products = append(products, p)
	}

	return products, nil
}

// CountDeletedProducts counts all soft-deleted products
func CountDeletedProducts(queryStr string, filters map[string]string) (int, error) {
	baseQuery := "SELECT COUNT(no) FROM master_products WHERE tanggal_hapus IS NOT NULL"
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
			OR CASE 
				WHEN tanggal_terima IS NOT NULL THEN 
					CASE
						WHEN (CURRENT_DATE - tanggal_terima) < 365 THEN 'Fresh'
						WHEN (CURRENT_DATE - tanggal_terima) < 730 THEN 'Normal'
						ELSE 'Aging'
					END
				ELSE 'Unknown' 
			   END ILIKE $` + fmt.Sprintf("%d", paramCount) + `
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

// fetchColorInfosForProduct gets color details for a given set of color IDs
func fetchColorInfosForProduct(colorIDs []int) ([]models.ColorInfo, error) {
	if len(colorIDs) == 0 {
		return []models.ColorInfo{}, nil
	}

	// Build a query with placeholders for all IDs
	query := "SELECT id, nama, hex FROM master_colors WHERE id IN ("
	args := make([]interface{}, 0, len(colorIDs))

	for i, id := range colorIDs {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+1)
		args = append(args, id)
	}
	query += ")"

	// Execute the query
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	colors := []models.ColorInfo{}
	for rows.Next() {
		var c models.ColorInfo
		if err := rows.Scan(&c.ID, &c.Name, &c.Hex); err != nil {
			return nil, err
		}
		colors = append(colors, c)
	}

	return colors, nil
}
