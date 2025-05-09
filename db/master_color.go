package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/everysoft/inventary-be/app/master_color"
)

func CreateMasterColorsTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_colors (
			id SERIAL PRIMARY KEY,
			nama TEXT NOT NULL,
			hex TEXT,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			tanggal_hapus TIMESTAMPTZ
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_colors_nama ON master_colors(nama);`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'uq_master_colors_nama'
			) THEN
				ALTER TABLE master_colors ADD CONSTRAINT uq_master_colors_nama UNIQUE (nama);
			END IF;
		END$$;`,
		// Add hex column if it doesn't exist (for backward compatibility)
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'master_colors' AND column_name = 'hex'
			) THEN
				ALTER TABLE master_colors ADD COLUMN hex TEXT;
			END IF;
		END$$;`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_colors table exists")
	return nil
}

func CountAllColors(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_colors WHERE tanggal_hapus IS NULL"
	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR nama ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR hex ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

func FetchAllColors(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]master_color.Color, error) {
	colors := []master_color.Color{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, nama, hex, tanggal_update, tanggal_hapus
	FROM master_colors
	WHERE tanggal_hapus IS NULL`

	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR nama ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR hex ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add sorting
	orderBy := " ORDER BY "
	// Map of valid column names to prevent SQL injection
	validColumns := map[string]bool{
		"id": true, "nama": true, "hex": true, "tanggal_update": true,
	}

	// Default sort
	if sortColumn == "" || !validColumns[sortColumn] {
		sortColumn = "id"
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
		var c master_color.Color
		if err := rows.Scan(
			&c.ID, &c.Nama, &c.Hex, &c.TanggalUpdate, &c.TanggalHapus,
		); err != nil {
			return nil, err
		}
		colors = append(colors, c)
	}

	return colors, nil
}

func FetchColorByID(id int) (master_color.Color, error) {
	var c master_color.Color
	err := DB.QueryRow(`SELECT id, nama, hex, tanggal_update, tanggal_hapus FROM master_colors WHERE id = $1 AND tanggal_hapus IS NULL`, id).
		Scan(&c.ID, &c.Nama, &c.Hex, &c.TanggalUpdate, &c.TanggalHapus)

	if err == sql.ErrNoRows {
		return c, errors.New("not_found")
	}
	return c, err
}

func InsertColor(c *master_color.Color) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_colors 
		(nama, hex, tanggal_update) 
		VALUES ($1, $2, $3)
		RETURNING id`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(
		c.Nama,
		c.Hex,
		c.TanggalUpdate,
	).Scan(&c.ID)
}

func UpdateColor(id int, c *master_color.Color) (master_color.Color, error) {
	// First fetch the existing color to get current values
	currentColor, err := FetchColorByID(id)
	if err != nil {
		return *c, err
	}

	// Build dynamic query with only fields that need to be updated
	query := "UPDATE master_colors SET"
	args := []interface{}{}
	paramCount := 1
	fieldsToUpdate := 0

	// Check each field and only update if it's not empty in the request
	// For string fields
	if c.Nama != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" nama = $%d,", paramCount)
		args = append(args, c.Nama)
		paramCount++
	}

	if c.Hex != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" hex = $%d,", paramCount)
		args = append(args, c.Hex)
		paramCount++
	}

	// Always update tanggal_update
	fieldsToUpdate++
	query += fmt.Sprintf(" tanggal_update = $%d,", paramCount)
	args = append(args, c.TanggalUpdate)
	paramCount++

	// Remove the trailing comma and complete the query
	query = query[:len(query)-1] + " WHERE id = $" + fmt.Sprintf("%d", paramCount)
	args = append(args, id)

	// If no fields to update, return the current color
	if fieldsToUpdate == 1 { // Only tanggal_update was added
		return currentColor, nil
	}

	// Execute the query
	result, err := DB.Exec(query, args...)
	if err != nil {
		return *c, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return *c, err
	}
	if rowsAffected == 0 {
		return *c, errors.New("not_found")
	}

	// Fetch updated color
	return FetchColorByID(id)
}

func DeleteColor(id int) error {
	// Soft delete by setting tanggal_hapus to the current time
	currentTime := time.Now()
	result, err := DB.Exec("UPDATE master_colors SET tanggal_hapus = $1 WHERE id = $2 AND tanggal_hapus IS NULL", currentTime, id)
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

func RestoreColor(id int) error {
	result, err := DB.Exec("UPDATE master_colors SET tanggal_hapus = NULL WHERE id = $1 AND tanggal_hapus IS NOT NULL", id)
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

func CountDeletedColors(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_colors WHERE tanggal_hapus IS NOT NULL"
	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR nama ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

func FetchDeletedColors(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]master_color.Color, error) {
	colors := []master_color.Color{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, nama, hex, tanggal_update, tanggal_hapus
	FROM master_colors
	WHERE tanggal_hapus IS NOT NULL`

	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR nama ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR hex ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add sorting
	orderBy := " ORDER BY "
	// Map of valid column names to prevent SQL injection
	validColumns := map[string]bool{
		"id": true, "nama": true, "hex": true, "tanggal_update": true, "tanggal_hapus": true,
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
		var c master_color.Color
		if err := rows.Scan(
			&c.ID, &c.Nama, &c.Hex, &c.TanggalUpdate, &c.TanggalHapus,
		); err != nil {
			return nil, err
		}
		colors = append(colors, c)
	}

	return colors, nil
}
