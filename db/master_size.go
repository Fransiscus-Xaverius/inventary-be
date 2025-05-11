package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/everysoft/inventary-be/app/master_size"
)

// CreateMasterSizesTableIfNotExists creates the master_sizes table if it doesn't exist
func CreateMasterSizesTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_sizes (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL UNIQUE,
			unit TEXT NOT NULL,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			tanggal_hapus TIMESTAMPTZ
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_sizes_value ON master_sizes(value);`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'uq_master_sizes_value_unit'
			) THEN
				ALTER TABLE master_sizes ADD CONSTRAINT uq_master_sizes_value_unit UNIQUE (value, unit);
			END IF;
		END$$;`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_sizes table exists")
	return nil
}

// CountAllSizes returns the total count of sizes matching the search query
func CountAllSizes(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_sizes WHERE tanggal_hapus IS NULL"
	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR value ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR unit ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

// GetAllSizes retrieves all sizes with pagination and search
func GetAllSizes(page, limit int, queryStr string) ([]*master_size.Size, error) {
	offset := (page - 1) * limit
	baseQuery := `
		SELECT id, value, unit, tanggal_update, tanggal_hapus
		FROM master_sizes
		WHERE tanggal_hapus IS NULL
	`
	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR value ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR unit ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add pagination
	baseQuery += ` ORDER BY id ASC LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	args = append(args, limit, offset)

	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sizes: %w", err)
	}
	defer rows.Close()

	var sizes []*master_size.Size
	for rows.Next() {
		var size master_size.Size
		var tanggalHapus sql.NullTime
		if err := rows.Scan(&size.ID, &size.Value, &size.Unit, &size.TanggalUpdate, &tanggalHapus); err != nil {
			return nil, fmt.Errorf("failed to scan size row: %w", err)
		}
		if tanggalHapus.Valid {
			size.TanggalHapus = &tanggalHapus.Time
		}
		sizes = append(sizes, &size)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating size rows: %w", err)
	}

	return sizes, nil
}

// GetSizeByID retrieves a size by its ID
func GetSizeByID(id int) (*master_size.Size, error) {
	var size master_size.Size
	var tanggalHapus sql.NullTime

	err := DB.QueryRow(`
		SELECT id, value, unit, tanggal_update, tanggal_hapus
		FROM master_sizes
		WHERE id = $1 AND tanggal_hapus IS NULL
	`, id).Scan(&size.ID, &size.Value, &size.Unit, &size.TanggalUpdate, &tanggalHapus)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get size: %w", err)
	}

	if tanggalHapus.Valid {
		size.TanggalHapus = &tanggalHapus.Time
	}

	return &size, nil
}

// CreateSize creates a new size
func CreateSize(size *master_size.Size) error {
	err := DB.QueryRow(`
		INSERT INTO master_sizes (value, unit)
		VALUES ($1, $2)
		RETURNING id, tanggal_update
	`, size.Value, size.Unit).Scan(&size.ID, &size.TanggalUpdate)

	if err != nil {
		return fmt.Errorf("failed to create size: %w", err)
	}

	return nil
}

// UpdateSize updates an existing size
func UpdateSize(size *master_size.Size) error {
	result, err := DB.Exec(`
		UPDATE master_sizes
		SET value = $1, unit = $2, tanggal_update = CURRENT_TIMESTAMP
		WHERE id = $3 AND tanggal_hapus IS NULL
	`, size.Value, size.Unit, size.ID)

	if err != nil {
		return fmt.Errorf("failed to update size: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("size not found or already deleted")
	}

	// Get the updated timestamp
	err = DB.QueryRow(`
		SELECT tanggal_update
		FROM master_sizes
		WHERE id = $1
	`, size.ID).Scan(&size.TanggalUpdate)

	if err != nil {
		return fmt.Errorf("failed to get updated timestamp: %w", err)
	}

	return nil
}

// DeleteSize soft deletes a size by setting tanggal_hapus
func DeleteSize(id int) error {
	result, err := DB.Exec(`
		UPDATE master_sizes
		SET tanggal_hapus = CURRENT_TIMESTAMP
		WHERE id = $1 AND tanggal_hapus IS NULL
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete size: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("size not found or already deleted")
	}

	return nil
}

// RestoreSize restores a soft-deleted size by clearing tanggal_hapus
func RestoreSize(id int) error {
	result, err := DB.Exec(`
		UPDATE master_sizes
		SET tanggal_hapus = NULL, tanggal_update = CURRENT_TIMESTAMP
		WHERE id = $1 AND tanggal_hapus IS NOT NULL
	`, id)

	if err != nil {
		return fmt.Errorf("failed to restore size: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("size not found or not deleted")
	}

	return nil
}

// CountDeletedSizes returns the total count of deleted sizes matching the search query
func CountDeletedSizes(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_sizes WHERE tanggal_hapus IS NOT NULL"
	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR value ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR unit ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

// GetDeletedSizes retrieves all deleted sizes with pagination and search
func GetDeletedSizes(page, limit int, queryStr string) ([]*master_size.Size, error) {
	offset := (page - 1) * limit
	baseQuery := `
		SELECT id, value, unit, tanggal_update, tanggal_hapus
		FROM master_sizes
		WHERE tanggal_hapus IS NOT NULL
	`
	args := []interface{}{}
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR value ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR unit ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add pagination
	baseQuery += ` ORDER BY tanggal_hapus DESC LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	args = append(args, limit, offset)

	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query deleted sizes: %w", err)
	}
	defer rows.Close()

	var sizes []*master_size.Size
	for rows.Next() {
		var size master_size.Size
		var tanggalHapus sql.NullTime
		if err := rows.Scan(&size.ID, &size.Value, &size.Unit, &size.TanggalUpdate, &tanggalHapus); err != nil {
			return nil, fmt.Errorf("failed to scan size row: %w", err)
		}
		if tanggalHapus.Valid {
			size.TanggalHapus = &tanggalHapus.Time
		}
		sizes = append(sizes, &size)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating size rows: %w", err)
	}

	return sizes, nil
}

// GetSizeByIDIncludeDeleted retrieves a size by its ID, including deleted ones
func GetSizeByIDIncludeDeleted(id int) (*master_size.Size, error) {
	var size master_size.Size
	var tanggalHapus sql.NullTime

	err := DB.QueryRow(`
		SELECT id, value, unit, tanggal_update, tanggal_hapus
		FROM master_sizes
		WHERE id = $1
	`, id).Scan(&size.ID, &size.Value, &size.Unit, &size.TanggalUpdate, &tanggalHapus)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("not_found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get size: %w", err)
	}

	if tanggalHapus.Valid {
		size.TanggalHapus = &tanggalHapus.Time
	}

	return &size, nil
}
