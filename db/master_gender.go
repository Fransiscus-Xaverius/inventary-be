package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/everysoft/inventary-be/app/models"
)

// CreateMasterGendersTableIfNotExists ensures the master_genders table exists
func CreateMasterGendersTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_genders (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			tanggal_hapus TIMESTAMPTZ
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_genders_value ON master_genders(value);`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_genders table exists")
	return nil
}

// CountAllGenders counts all available genders matching the search query
func CountAllGenders(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_genders WHERE tanggal_hapus IS NULL"
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

// FetchAllGenders retrieves all genders matching the search query with pagination
func FetchAllGenders(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Gender, error) {
	genders := []models.Gender{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_genders
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
		var g models.Gender
		if err := rows.Scan(
			&g.ID, &g.Value, &g.TanggalUpdate, &g.TanggalHapus,
		); err != nil {
			return nil, err
		}
		genders = append(genders, g)
	}

	return genders, nil
}

// FetchGenderByID retrieves a gender by its ID
func FetchGenderByID(id int) (models.Gender, error) {
	var g models.Gender
	err := DB.QueryRow(`SELECT id, value, tanggal_update, tanggal_hapus FROM master_genders WHERE id = $1 AND tanggal_hapus IS NULL`, id).
		Scan(&g.ID, &g.Value, &g.TanggalUpdate, &g.TanggalHapus)

	if err == sql.ErrNoRows {
		return g, errors.New("not_found")
	}
	return g, err
}

// InsertGender inserts a new gender record
func InsertGender(g *models.Gender) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_genders 
		(value, tanggal_update) 
		VALUES ($1, $2)
		RETURNING id`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(
		g.Value,
		time.Now(),
	).Scan(&g.ID)
}

// UpdateGender updates an existing gender record
func UpdateGender(id int, g *models.Gender) (models.Gender, error) {
	// First check if the gender exists
	_, err := FetchGenderByID(id)
	if err != nil {
		return *g, err
	}

	// Build dynamic query with only fields that need to be updated
	query := "UPDATE master_genders SET"
	args := []interface{}{}
	paramCount := 1
	fieldsToUpdate := 0

	// Check if value is provided
	if g.Value != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" value = $%d,", paramCount)
		args = append(args, g.Value)
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
		return *g, err
	}

	// Fetch the updated record
	return FetchGenderByID(id)
}

// DeleteGender soft-deletes a gender by setting tanggal_hapus
func DeleteGender(id int) error {
	_, err := DB.Exec(`UPDATE master_genders SET tanggal_hapus = $1 WHERE id = $2 AND tanggal_hapus IS NULL`,
		time.Now(), id)
	return err
}

// RestoreGender restores a soft-deleted gender
func RestoreGender(id int) error {
	_, err := DB.Exec(`UPDATE master_genders SET tanggal_hapus = NULL WHERE id = $1`, id)
	return err
}

// CountDeletedGenders counts all deleted genders matching the search query
func CountDeletedGenders(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_genders WHERE tanggal_hapus IS NOT NULL"
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

// FetchDeletedGenders retrieves all deleted genders matching the search query with pagination
func FetchDeletedGenders(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Gender, error) {
	genders := []models.Gender{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_genders
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
		var g models.Gender
		if err := rows.Scan(
			&g.ID, &g.Value, &g.TanggalUpdate, &g.TanggalHapus,
		); err != nil {
			return nil, err
		}
		genders = append(genders, g)
	}

	return genders, nil
}
