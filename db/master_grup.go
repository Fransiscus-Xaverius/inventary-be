package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/everysoft/inventary-be/app/models"
)

// CreateMasterGrupsTableIfNotExists ensures the master_grups table exists
func CreateMasterGrupsTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_grups (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			tanggal_hapus TIMESTAMPTZ
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_grups_value ON master_grups(value);`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_grups table exists")
	return nil
}

// CountAllGrups counts all available grups matching the search query
func CountAllGrups(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_grups WHERE tanggal_hapus IS NULL"
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

// FetchAllGrups retrieves all grups matching the search query with pagination
func FetchAllGrups(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Grup, error) {
	grups := []models.Grup{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_grups
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
		var g models.Grup
		if err := rows.Scan(
			&g.ID, &g.Value, &g.TanggalUpdate, &g.TanggalHapus,
		); err != nil {
			return nil, err
		}
		grups = append(grups, g)
	}

	return grups, nil
}

// FetchGrupByID retrieves a grup by its ID
func FetchGrupByID(id int) (models.Grup, error) {
	var g models.Grup
	err := DB.QueryRow(`SELECT id, value, tanggal_update, tanggal_hapus FROM master_grups WHERE id = $1 AND tanggal_hapus IS NULL`, id).
		Scan(&g.ID, &g.Value, &g.TanggalUpdate, &g.TanggalHapus)

	if err == sql.ErrNoRows {
		return g, errors.New("not_found")
	}
	return g, err
}

// InsertGrup inserts a new grup record
func InsertGrup(g *models.Grup) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_grups 
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

// UpdateGrup updates an existing grup record
func UpdateGrup(id int, g *models.Grup) (models.Grup, error) {
	// First check if the grup exists
	_, err := FetchGrupByID(id)
	if err != nil {
		return *g, err
	}

	// Build dynamic query with only fields that need to be updated
	query := "UPDATE master_grups SET"
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
	return FetchGrupByID(id)
}

// DeleteGrup soft-deletes a grup by setting tanggal_hapus
func DeleteGrup(id int) error {
	_, err := DB.Exec(`UPDATE master_grups SET tanggal_hapus = $1 WHERE id = $2 AND tanggal_hapus IS NULL`,
		time.Now(), id)
	return err
}

// RestoreGrup restores a soft-deleted grup
func RestoreGrup(id int) error {
	_, err := DB.Exec(`UPDATE master_grups SET tanggal_hapus = NULL WHERE id = $1`, id)
	return err
}

// CountDeletedGrups counts all deleted grups matching the search query
func CountDeletedGrups(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_grups WHERE tanggal_hapus IS NOT NULL"
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

// FetchDeletedGrups retrieves all deleted grups matching the search query with pagination
func FetchDeletedGrups(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Grup, error) {
	grups := []models.Grup{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_grups
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
		var g models.Grup
		if err := rows.Scan(
			&g.ID, &g.Value, &g.TanggalUpdate, &g.TanggalHapus,
		); err != nil {
			return nil, err
		}
		grups = append(grups, g)
	}

	return grups, nil
}
