package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/everysoft/inventary-be/app/models"
)

// CreateMasterTipesTableIfNotExists ensures the master_tipes table exists
func CreateMasterTipesTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_tipes (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			tanggal_hapus TIMESTAMPTZ
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_tipes_value ON master_tipes(value);`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_tipes table exists")
	return nil
}

// CountAllTipes counts all available tipes matching the search query
func CountAllTipes(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_tipes WHERE tanggal_hapus IS NULL"
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

// FetchAllTipes retrieves all tipes matching the search query with pagination
func FetchAllTipes(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Tipe, error) {
	tipes := []models.Tipe{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_tipes
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
		var t models.Tipe
		if err := rows.Scan(
			&t.ID, &t.Value, &t.TanggalUpdate, &t.TanggalHapus,
		); err != nil {
			return nil, err
		}
		tipes = append(tipes, t)
	}

	return tipes, nil
}

// FetchTipeByID retrieves a tipe by its ID
func FetchTipeByID(id int) (models.Tipe, error) {
	var t models.Tipe
	err := DB.QueryRow(`SELECT id, value, tanggal_update, tanggal_hapus FROM master_tipes WHERE id = $1 AND tanggal_hapus IS NULL`, id).
		Scan(&t.ID, &t.Value, &t.TanggalUpdate, &t.TanggalHapus)

	if err == sql.ErrNoRows {
		return t, errors.New("not_found")
	}
	return t, err
}

// InsertTipe inserts a new tipe record
func InsertTipe(t *models.Tipe) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_tipes 
		(value, tanggal_update) 
		VALUES ($1, $2)
		RETURNING id`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(
		t.Value,
		time.Now(),
	).Scan(&t.ID)
}

// UpdateTipe updates an existing tipe record
func UpdateTipe(id int, t *models.Tipe) (models.Tipe, error) {
	// First check if the tipe exists
	_, err := FetchTipeByID(id)
	if err != nil {
		return *t, err
	}

	// Build dynamic query with only fields that need to be updated
	query := "UPDATE master_tipes SET"
	args := []interface{}{}
	paramCount := 1
	fieldsToUpdate := 0

	// Check if value is provided
	if t.Value != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" value = $%d,", paramCount)
		args = append(args, t.Value)
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
		return *t, err
	}

	// Fetch the updated record
	return FetchTipeByID(id)
}

// DeleteTipe soft-deletes a tipe by setting tanggal_hapus
func DeleteTipe(id int) error {
	_, err := DB.Exec(`UPDATE master_tipes SET tanggal_hapus = $1 WHERE id = $2 AND tanggal_hapus IS NULL`,
		time.Now(), id)
	return err
}

// RestoreTipe restores a soft-deleted tipe
func RestoreTipe(id int) error {
	_, err := DB.Exec(`UPDATE master_tipes SET tanggal_hapus = NULL WHERE id = $1`, id)
	return err
}

// CountDeletedTipes counts all deleted tipes matching the search query
func CountDeletedTipes(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_tipes WHERE tanggal_hapus IS NOT NULL"
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

// FetchDeletedTipes retrieves all deleted tipes matching the search query with pagination
func FetchDeletedTipes(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Tipe, error) {
	tipes := []models.Tipe{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_tipes
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
		var t models.Tipe
		if err := rows.Scan(
			&t.ID, &t.Value, &t.TanggalUpdate, &t.TanggalHapus,
		); err != nil {
			return nil, err
		}
		tipes = append(tipes, t)
	}

	return tipes, nil
}
