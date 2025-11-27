package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/everysoft/inventary-be/app/models"
)

// CreateMasterKatsTableIfNotExists ensures the master_kats table exists
func CreateMasterKatsTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_kats (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			tanggal_hapus TIMESTAMPTZ
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_kats_value ON master_kats(value);`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_kats table exists")
	return nil
}

// CountAllKats counts all available categories matching the search query
func CountAllKats(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_kats WHERE tanggal_hapus IS NULL"
	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR value ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

// FetchAllKats retrieves all categories matching the search query with pagination
func FetchAllKats(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Kat, error) {
	kats := []models.Kat{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_kats
	WHERE tanggal_hapus IS NULL`

	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR value ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add sorting
	orderBy := " ORDER BY "
	// Map of valid column names to prevent SQL injection
	validColumns := map[string]bool{
		"id": true, "value": true, "tanggal_update": true,
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
		var k models.Kat
		if err := rows.Scan(
			&k.ID, &k.Value, &k.TanggalUpdate, &k.TanggalHapus,
		); err != nil {
			return nil, err
		}
		kats = append(kats, k)
	}

	return kats, nil
}

// FetchKatByID retrieves a category by its ID
func FetchKatByID(id int) (models.Kat, error) {
	var k models.Kat
	err := DB.QueryRow(`SELECT id, value, tanggal_update, tanggal_hapus FROM master_kats WHERE id = $1 AND tanggal_hapus IS NULL`, id).
		Scan(&k.ID, &k.Value, &k.TanggalUpdate, &k.TanggalHapus)

	if err == sql.ErrNoRows {
		return k, errors.New("not_found")
	}
	return k, err
}

// InsertKat inserts a new category record
func InsertKat(k *models.Kat) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_kats 
		(value, tanggal_update) 
		VALUES ($1, $2)
		RETURNING id`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(
		k.Value,
		time.Now(),
	).Scan(&k.ID)
}

// UpdateKat updates an existing category record
func UpdateKat(id int, k *models.Kat) (models.Kat, error) {
	// First check if the category exists
	_, err := FetchKatByID(id)
	if err != nil {
		return *k, err
	}

	// Build dynamic query with only fields that need to be updated
	query := "UPDATE master_kats SET"
	args := []interface{}{}
	paramCount := 1
	fieldsToUpdate := 0

	// Check if value is provided
	if k.Value != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" value = $%d,", paramCount)
		args = append(args, k.Value)
		paramCount++
	}

	// Add tanggal_update
	query += fmt.Sprintf(" tanggal_update = $%d", paramCount)
	args = append(args, time.Now())
	paramCount++

	// Add WHERE clause
	query += fmt.Sprintf(" WHERE id = $%d", paramCount)
	args = append(args, id)

	// Execute update
	_, err = DB.Exec(query, args...)
	if err != nil {
		return *k, err
	}

	// Fetch the updated record
	return FetchKatByID(id)
}

// DeleteKat soft-deletes a category by setting tanggal_hapus
func DeleteKat(id int) error {
	_, err := DB.Exec(`UPDATE master_kats SET tanggal_hapus = $1 WHERE id = $2 AND tanggal_hapus IS NULL`,
		time.Now(), id)
	return err
}

// RestoreKat restores a soft-deleted category
func RestoreKat(id int) error {
	_, err := DB.Exec(`UPDATE master_kats SET tanggal_hapus = NULL WHERE id = $1`, id)
	return err
}

// CountDeletedKats counts all deleted categories matching the search query
func CountDeletedKats(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_kats WHERE tanggal_hapus IS NOT NULL"
	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR value ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

// FetchDeletedKats retrieves all deleted categories matching the search query with pagination
func FetchDeletedKats(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Kat, error) {
	kats := []models.Kat{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_kats
	WHERE tanggal_hapus IS NOT NULL`

	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR value ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add sorting
	orderBy := " ORDER BY "
	// Map of valid column names to prevent SQL injection
	validColumns := map[string]bool{
		"id": true, "value": true, "tanggal_update": true, "tanggal_hapus": true,
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
		var k models.Kat
		if err := rows.Scan(
			&k.ID, &k.Value, &k.TanggalUpdate, &k.TanggalHapus,
		); err != nil {
			return nil, err
		}
		kats = append(kats, k)
	}

	return kats, nil
}
