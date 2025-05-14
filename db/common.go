package db

import (
	"fmt"
)

// CheckMasterDataExists verifies if a given ID exists in the specified master data table
// tableName is the name of the master data table (e.g., "master_grups", "master_units")
// id is the ID to check for existence
func CheckMasterDataExists(tableName string, id string) (bool, error) {
	// Build the query to check if the ID exists in the specified table
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1 AND tanggal_hapus IS NULL LIMIT 1", tableName)

	var exists int
	err := DB.QueryRow(query, id).Scan(&exists)

	if err != nil {
		// If no rows found, the ID doesn't exist
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
