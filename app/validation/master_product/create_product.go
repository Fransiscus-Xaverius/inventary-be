package master_product

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/app/validation"
	"github.com/everysoft/inventary-be/app/validation/common"
	"github.com/everysoft/inventary-be/db"
)

// ValidationSchema contains validation rules for product creation
type ValidationSchema struct {
	ArtikelRequired       bool `json:"artikel_required"`
	ArtikelUnique         bool `json:"artikel_unique"`
	NamaRequired          bool `json:"nama_required"`
	DeskripsiRequired     bool `json:"deskripsi_required"`
	WarnaRequired         bool `json:"warna_required"`
	SizeRequired          bool `json:"size_required"`
	GrupRequired          bool `json:"grup_required"`
	UnitRequired          bool `json:"unit_required"`
	KatRequired           bool `json:"kat_required"`
	ModelRequired         bool `json:"model_required"`
	GenderRequired        bool `json:"gender_required"`
	TipeRequired          bool `json:"tipe_required"`
	HargaRequired         bool `json:"harga_required"`
	MarketplaceRequired   bool `json:"marketplace_required"`
	GambarRequired        bool `json:"gambar_required"`
	TanggalProdukRequired bool `json:"tanggal_produk_required"`
	TanggalTerimaRequired bool `json:"tanggal_terima_required"`
	StatusRequired        bool `json:"status_required"`
	SupplierRequired      bool `json:"supplier_required"`
	DiupdateOlehRequired  bool `json:"diupdate_oleh_required"`
}

// DefaultCreateSchema returns the default validation schema for product creation
func DefaultCreateSchema() ValidationSchema {
	return ValidationSchema{
		ArtikelRequired:       true,
		ArtikelUnique:         true,
		NamaRequired:          true,
		DeskripsiRequired:     true,
		WarnaRequired:         true,
		SizeRequired:          true,
		GrupRequired:          true,
		UnitRequired:          true,
		KatRequired:           true,
		ModelRequired:         true,
		GenderRequired:        true,
		TipeRequired:          true,
		HargaRequired:         true,
		MarketplaceRequired:   true,
		GambarRequired:        false,
		TanggalProdukRequired: true,
		TanggalTerimaRequired: true,
		StatusRequired:        true,
		SupplierRequired:      true,
		DiupdateOlehRequired:  true,
	}
}

// DefaultUpdateSchema returns the default validation schema for product update
func DefaultUpdateSchema() ValidationSchema {
	return ValidationSchema{
		ArtikelRequired:       false,
		ArtikelUnique:         false,
		NamaRequired:          false,
		DeskripsiRequired:     false,
		WarnaRequired:         false,
		SizeRequired:          false,
		GrupRequired:          false,
		UnitRequired:          false,
		KatRequired:           false,
		ModelRequired:         false,
		GenderRequired:        false,
		TipeRequired:          false,
		HargaRequired:         false,
		MarketplaceRequired:   false,
		GambarRequired:        false,
		TanggalProdukRequired: false,
		TanggalTerimaRequired: false,
		StatusRequired:        false,
		SupplierRequired:      false,
		DiupdateOlehRequired:  false,
	}
}

// ValidateProduct performs all validations for product creation and update
func ValidateProduct(p *models.Product, schema ValidationSchema) *validation.ValidationError {
	// Validate Artikel (required and unique)
	if schema.ArtikelRequired && strings.TrimSpace(p.Artikel) == "" {
		return &validation.ValidationError{
			Error:      "Artikel is required",
			ErrorField: "artikel",
		}
	}

	// Check for duplicate artikel if not empty
	if schema.ArtikelUnique && strings.TrimSpace(p.Artikel) != "" {
		_, err := db.FetchProductByArtikelIncludeDeleted(p.Artikel)
		if err == nil || (err != nil && err.Error() != "not_found") {
			// If no error, then product exists
			// Or if there's an error but it's not "not_found", then there's a different issue
			if err == nil {
				return &validation.ValidationError{
					Error:      "Product with this artikel already exists",
					ErrorField: "artikel",
				}
			}
			// Otherwise, there was a database error
			return &validation.ValidationError{
				Error:      "Error checking artikel uniqueness: " + err.Error(),
				ErrorField: "artikel",
			}
		}
	}

	// Validate warna (required, comma-separated color IDs)
	if schema.WarnaRequired && strings.TrimSpace(p.Warna) == "" {
		return &validation.ValidationError{
			Error:      "Warna is required",
			ErrorField: "warna",
		}
	}

	// Validate that all color IDs exist if warna is not empty
	if strings.TrimSpace(p.Warna) != "" {
		warna := strings.TrimSpace(p.Warna)
		p.Warna = warna // Store the trimmed value

		// Check if all color IDs are valid
		colorIDs := strings.Split(warna, ",")
		for _, colorIDStr := range colorIDs {
			colorIDStr = strings.TrimSpace(colorIDStr)
			if colorIDStr == "" {
				continue
			}

			colorID, err := strconv.Atoi(colorIDStr)
			if err != nil {
				return &validation.ValidationError{
					Error:      "Invalid color ID format: " + colorIDStr,
					ErrorField: "warna",
				}
			}

			_, err = db.FetchColorByID(colorID)
			if err != nil {
				if err.Error() == "not_found" {
					return &validation.ValidationError{
						Error:      "Color ID not found: " + colorIDStr,
						ErrorField: "warna",
					}
				}
				return &validation.ValidationError{
					Error:      "Error checking color ID: " + err.Error(),
					ErrorField: "warna",
				}
			}
		}
	}

	// Validate size (required, EU shoe sizes or ranges)
	if schema.SizeRequired && strings.TrimSpace(p.Size) == "" {
		return &validation.ValidationError{
			Error:      "Size is required",
			ErrorField: "size",
		}
	}

	// Validate EU shoe sizes or ranges if size is not empty
	if strings.TrimSpace(p.Size) != "" {
		size := strings.TrimSpace(p.Size)
		p.Size = size // Store the trimmed value

		// Validate EU shoe sizes or ranges
		sizes := strings.Split(size, ",")
		for _, s := range sizes {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}

			// Check if it's a range (e.g., "38-40")
			if strings.Contains(s, "-") {
				rangeParts := strings.Split(s, "-")
				if len(rangeParts) != 2 {
					return &validation.ValidationError{
						Error:      "Invalid size range format: " + s,
						ErrorField: "size",
					}
				}

				start, errStart := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				end, errEnd := strconv.Atoi(strings.TrimSpace(rangeParts[1]))

				if errStart != nil || errEnd != nil || start >= end {
					return &validation.ValidationError{
						Error:      "Invalid size range values in: " + s,
						ErrorField: "size",
					}
				}
			} else {
				// Single size
				_, err := strconv.Atoi(s)
				if err != nil {
					return &validation.ValidationError{
						Error:      "Size must be numeric: " + s,
						ErrorField: "size",
					}
				}
			}
		}
	}

	// Validate master data fields

	// Nama
	if schema.NamaRequired && strings.TrimSpace(p.Nama) == "" {
		return &validation.ValidationError{
			Error:      "Nama is required",
			ErrorField: "nama",
		}
	}

	// Deskripsi
	if schema.DeskripsiRequired && strings.TrimSpace(p.Deskripsi) == "" {
		return &validation.ValidationError{
			Error:      "Deskripsi is required",
			ErrorField: "deskripsi",
		}
	}

	// Marketplace
	if schema.MarketplaceRequired && p.Marketplace == (models.MarketplaceInfo{}) {
		return &validation.ValidationError{
			Error:      "Marketplace is required",
			ErrorField: "marketplace",
		}
	}

	// Gambar
	// Grup
	if schema.GrupRequired && strings.TrimSpace(p.Grup) == "" {
		return &validation.ValidationError{
			Error:      "Grup is required",
			ErrorField: "grup",
		}
	}

	if strings.TrimSpace(p.Grup) != "" {
		if validationError := common.ValidateMasterDataID("master_grups", "grup", p.Grup); validationError != nil {
			return validationError
		}
	}

	// Unit
	if schema.UnitRequired && strings.TrimSpace(p.Unit) == "" {
		return &validation.ValidationError{
			Error:      "Unit is required",
			ErrorField: "unit",
		}
	}

	if strings.TrimSpace(p.Unit) != "" {
		if errObj := common.ValidateMasterDataID("master_units", "unit", p.Unit); errObj != nil {
			return errObj
		}
	}

	// Kat
	if schema.KatRequired && strings.TrimSpace(p.Kat) == "" {
		return &validation.ValidationError{
			Error:      "Kat is required",
			ErrorField: "kat",
		}
	}

	if strings.TrimSpace(p.Kat) != "" {
		if validationError := common.ValidateMasterDataID("master_kats", "kat", p.Kat); validationError != nil {
			return validationError
		}
	}

	// Gender
	if schema.GenderRequired && strings.TrimSpace(p.Gender) == "" {
		return &validation.ValidationError{
			Error:      "Gender is required",
			ErrorField: "gender",
		}
	}

	if strings.TrimSpace(p.Gender) != "" {
		if validationError := common.ValidateMasterDataID("master_genders", "gender", p.Gender); validationError != nil {
			return validationError
		}
	}

	// Tipe
	if schema.TipeRequired && strings.TrimSpace(p.Tipe) == "" {
		return &validation.ValidationError{
			Error:      "Tipe is required",
			ErrorField: "tipe",
		}
	}

	if strings.TrimSpace(p.Tipe) != "" {
		if validationError := common.ValidateMasterDataID("master_tipes", "tipe", p.Tipe); validationError != nil {
			return validationError
		}
	}

	// Model (required)
	if schema.ModelRequired && strings.TrimSpace(p.Model) == "" {
		return &validation.ValidationError{
			Error:      "Model is required",
			ErrorField: "model",
		}
	}

	// Validate harga (required, numeric)
	if schema.HargaRequired && p.Harga <= 0 {
		return &validation.ValidationError{
			Error:      "Harga is required and must be a positive number",
			ErrorField: "harga",
		}
	}

	// Validate tanggal_produk (required, valid date)
	zeroTime := time.Time{}
	if schema.TanggalProdukRequired && p.TanggalProduk == zeroTime {
		return &validation.ValidationError{
			Error:      "Tanggal produk is required",
			ErrorField: "tanggal_produk",
		}
	}

	// Validate tanggal_terima (required, valid date)
	if schema.TanggalTerimaRequired && p.TanggalTerima == zeroTime {
		return &validation.ValidationError{
			Error:      "Tanggal terima is required",
			ErrorField: "tanggal_terima",
		}
	}

	// Validate status (required, must be one of "active", "inactive", or "discontinued")
	if schema.StatusRequired && strings.TrimSpace(p.Status) == "" {
		return &validation.ValidationError{
			Error:      "Status is required",
			ErrorField: "status",
		}
	}

	if strings.TrimSpace(p.Status) != "" {
		status := strings.TrimSpace(p.Status)
		p.Status = status // Store the trimmed value

		validStatuses := map[string]bool{
			"active":       true,
			"inactive":     true,
			"discontinued": true,
		}

		if !validStatuses[status] {
			return &validation.ValidationError{
				Error:      "Status must be either 'active', 'inactive', or 'discontinued'",
				ErrorField: "status",
			}
		}
	}

	// Validate supplier (required)
	if schema.SupplierRequired && strings.TrimSpace(p.Supplier) == "" {
		return &validation.ValidationError{
			Error:      "Supplier is required",
			ErrorField: "supplier",
		}
	}

	// Validate diupdate_oleh (required)
	if schema.DiupdateOlehRequired && strings.TrimSpace(p.DiupdateOleh) == "" {
		return &validation.ValidationError{
			Error:      "Diupdate oleh is required",
			ErrorField: "diupdate_oleh",
		}
	}

	// Validate rating (required structure with valid values)
	if p.Rating.Comfort < 0 || p.Rating.Comfort > 10 {
		return &validation.ValidationError{
			Error:      "Comfort rating must be between 0 and 10",
			ErrorField: "rating.comfort",
		}
	}

	if p.Rating.Style < 0 || p.Rating.Style > 10 {
		return &validation.ValidationError{
			Error:      "Style rating must be between 0 and 10",
			ErrorField: "rating.style",
		}
	}

	if p.Rating.Support < 0 || p.Rating.Support > 10 {
		return &validation.ValidationError{
			Error:      "Support rating must be between 0 and 10",
			ErrorField: "rating.support",
		}
	}

	if len(p.Rating.Purpose) == 0 {
		return &validation.ValidationError{
			Error:      "Purpose array cannot be empty",
			ErrorField: "rating.purpose",
		}
	}

	// Validate purpose array - check for empty strings
	for i, purpose := range p.Rating.Purpose {
		if strings.TrimSpace(purpose) == "" && len(p.Rating.Purpose) == 1 {
			// Allow single empty string as default
			continue
		}
		if strings.TrimSpace(purpose) == "" {
			return &validation.ValidationError{
				Error:      fmt.Sprintf("Purpose at index %d cannot be empty", i),
				ErrorField: "rating.purpose",
			}
		}
	}

	// All validations passed
	return nil
}

// ValidateCreate performs validation for product creation
func ValidateCreate(p *models.Product) *validation.ValidationError {
	return ValidateProduct(p, DefaultCreateSchema())
}

// ValidateUpdate performs validation for product update
func ValidateUpdate(p *models.Product) *validation.ValidationError {
	// Skip artikel uniqueness check for updates
	schema := DefaultUpdateSchema()
	return ValidateProduct(p, schema)
}
