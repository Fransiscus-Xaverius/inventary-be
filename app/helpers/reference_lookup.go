package helpers

import (
	"log"
	"reflect"
	"strconv"

	"github.com/everysoft/inventary-be/db"
)

// ValueGetter is an interface for objects that can return their value
type ValueGetter interface {
	GetValue() string
}

// ValueWrapper is a helper struct to implement the ValueGetter interface
type ValueWrapper struct {
	Value string
}

// GetValue returns the wrapped value
func (v ValueWrapper) GetValue() string {
	return v.Value
}

// GetValueFromID converts an ID string to its actual value from a reference table
// It handles errors gracefully by returning the original value if conversion fails
func GetValueFromID(idStr string, fetchFunc func(int) (ValueGetter, error)) string {
	if idStr == "" {
		return idStr
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("GetValueFromID: Invalid ID format for '%s': %v", idStr, err)
		return idStr // Return original if not a valid integer
	}

	entity, err := fetchFunc(id)
	if err != nil {
		log.Printf("GetValueFromID: Failed to fetch entity for ID %d: %v", id, err)
		return idStr // Return original if entity not found
	}

	return entity.GetValue()
}

// ConvertIDsToValues takes a map of field names to IDs and their corresponding fetch functions
// It returns a map of field names to their resolved values
func ConvertIDsToValues(fieldToID map[string]string, fieldToFetchFunc map[string]func(int) (ValueGetter, error)) map[string]string {
	result := make(map[string]string)

	for field, id := range fieldToID {
		if fetchFunc, ok := fieldToFetchFunc[field]; ok {
			result[field] = GetValueFromID(id, fetchFunc)
		} else {
			result[field] = id // Keep original if no fetch function
		}
	}

	return result
}

// ConvertProductFields converts the specified fields in a product from IDs to their values
// It automatically determines the appropriate fetch function for each field
func ConvertProductFields(product interface{}, fieldNames []string) {
	// Create a map of field names to fetch functions
	fetchFuncs := map[string]func(int) (ValueGetter, error){
		"Grup": func(id int) (ValueGetter, error) {
			grup, err := db.FetchGrupByID(id)
			return ValueWrapper{Value: grup.Value}, err
		},
		"Unit": func(id int) (ValueGetter, error) {
			unit, err := db.FetchUnitByID(id)
			return ValueWrapper{Value: unit.Value}, err
		},
		"Kat": func(id int) (ValueGetter, error) {
			kat, err := db.FetchKatByID(id)
			return ValueWrapper{Value: kat.Value}, err
		},
		"Gender": func(id int) (ValueGetter, error) {
			gender, err := db.FetchGenderByID(id)
			return ValueWrapper{Value: gender.Value}, err
		},
		"Tipe": func(id int) (ValueGetter, error) {
			tipe, err := db.FetchTipeByID(id)
			return ValueWrapper{Value: tipe.Value}, err
		},
	}

	// Get reflect value of product
	val := reflect.ValueOf(product).Elem()

	// Process each field name
	for _, fieldName := range fieldNames {
		// Get the field
		field := val.FieldByName(fieldName)

		if field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
			// Get current ID value
			idStr := field.String()

			// Get the appropriate fetch function
			if fetchFunc, ok := fetchFuncs[fieldName]; ok {
				// Convert ID to actual value
				value := GetValueFromID(idStr, fetchFunc)

				// Set the field to the converted value
				field.SetString(value)
			}
		}
	}
}
