package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/everysoft/inventary-be/app/models"
)

// CreateMasterUnitsTableIfNotExists ensures the master_units table exists
func CreateMasterUnitsTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_units (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			tanggal_hapus TIMESTAMPTZ
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_units_value ON master_units(value);`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'uq_master_units_value'
			) THEN
				ALTER TABLE master_units ADD CONSTRAINT uq_master_units_value UNIQUE (value);
			END IF;
		END$$;`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_units table exists")
	return nil
}

// CountAllUnits counts all available units matching the search query
func CountAllUnits(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_units WHERE tanggal_hapus IS NULL"
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

// FetchAllUnits retrieves all units matching the search query with pagination
func FetchAllUnits(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Unit, error) {
	units := []models.Unit{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_units
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
		var u models.Unit
		if err := rows.Scan(
			&u.ID, &u.Value, &u.TanggalUpdate, &u.TanggalHapus,
		); err != nil {
			return nil, err
		}
		units = append(units, u)
	}

	return units, nil
}

// FetchUnitByID retrieves a unit by its ID
func FetchUnitByID(id int) (models.Unit, error) {
	var u models.Unit
	err := DB.QueryRow(`SELECT id, value, tanggal_update, tanggal_hapus FROM master_units WHERE id = $1 AND tanggal_hapus IS NULL`, id).
		Scan(&u.ID, &u.Value, &u.TanggalUpdate, &u.TanggalHapus)

	if err == sql.ErrNoRows {
		return u, errors.New("not_found")
	}
	return u, err
}

// InsertUnit inserts a new unit record
func InsertUnit(u *models.Unit) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_units 
		(value, tanggal_update) 
		VALUES ($1, $2)
		RETURNING id`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(
		u.Value,
		time.Now(),
	).Scan(&u.ID)
}

// UpdateUnit updates an existing unit record
func UpdateUnit(id int, u *models.Unit) (models.Unit, error) {
	// First check if the unit exists
	_, err := FetchUnitByID(id)
	if err != nil {
		return *u, err
	}

	// Build dynamic query with only fields that need to be updated
	query := "UPDATE master_units SET"
	args := []interface{}{}
	paramCount := 1
	fieldsToUpdate := 0

	// Check if value is provided
	if u.Value != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" value = $%d,", paramCount)
		args = append(args, u.Value)
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
		return *u, err
	}

	// Fetch the updated record
	return FetchUnitByID(id)
}

// DeleteUnit soft-deletes a unit by setting tanggal_hapus
func DeleteUnit(id int) error {
	_, err := DB.Exec(`UPDATE master_units SET tanggal_hapus = $1 WHERE id = $2 AND tanggal_hapus IS NULL`,
		time.Now(), id)
	return err
}

// RestoreUnit restores a soft-deleted unit
func RestoreUnit(id int) error {
	_, err := DB.Exec(`UPDATE master_units SET tanggal_hapus = NULL WHERE id = $1`, id)
	return err
}

// CountDeletedUnits counts all deleted units matching the search query
func CountDeletedUnits(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_units WHERE tanggal_hapus IS NOT NULL"
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

// FetchDeletedUnits retrieves all deleted units matching the search query with pagination
func FetchDeletedUnits(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Unit, error) {
	units := []models.Unit{}

	// Start building the query with parameters
	baseQuery := `
	SELECT 
		id, value, tanggal_update, tanggal_hapus
	FROM master_units
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
		var u models.Unit
		if err := rows.Scan(
			&u.ID, &u.Value, &u.TanggalUpdate, &u.TanggalHapus,
		); err != nil {
			return nil, err
		}
		units = append(units, u)
	}

	return units, nil
}
