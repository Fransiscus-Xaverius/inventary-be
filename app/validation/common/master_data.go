package common

import (
	"fmt"

	"github.com/everysoft/inventary-be/app/validation"
	"github.com/everysoft/inventary-be/db"
)

// MasterDataExists verifies if a given ID exists in the specified master data table
// tableName is the name of the master data table (e.g., "master_grups", "master_units")
// id is the ID to check for existence
func MasterDataExists(tableName string, id string) (bool, error) {
	return db.CheckMasterDataExists(tableName, id)
}

// ValidateMasterDataID checks if an ID exists in a master data table and returns an error object if not
func ValidateMasterDataID(tableName string, fieldName string, id string) *validation.ValidationError {
	if id == "" {
		return &validation.ValidationError{
			Error:      fmt.Sprintf("%s is required", fieldName),
			ErrorField: fieldName,
		}
	}

	exists, err := MasterDataExists(tableName, id)
	if err != nil {
		return &validation.ValidationError{
			Error:      fmt.Sprintf("Error checking %s: %s", fieldName, err.Error()),
			ErrorField: fieldName,
		}
	}

	if !exists {
		return &validation.ValidationError{
			Error:      fmt.Sprintf("%s ID does not exist in master data", fieldName),
			ErrorField: fieldName,
		}
	}

	return nil
}
