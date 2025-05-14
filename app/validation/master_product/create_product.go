package master_product

import (
	"strconv"
	"strings"
	"time"

	"github.com/everysoft/inventary-be/app/models"
	"github.com/everysoft/inventary-be/app/validation/common"
	"github.com/everysoft/inventary-be/db"
	"github.com/gin-gonic/gin"
)

// ValidationSchema contains validation rules for product creation
type ValidationSchema struct {
	ArtikelRequired       bool `json:"artikel_required"`
	WarnaRequired         bool `json:"warna_required"`
	SizeRequired          bool `json:"size_required"`
	GrupRequired          bool `json:"grup_required"`
	UnitRequired          bool `json:"unit_required"`
	KatRequired           bool `json:"kat_required"`
	ModelRequired         bool `json:"model_required"`
	GenderRequired        bool `json:"gender_required"`
	TipeRequired          bool `json:"tipe_required"`
	HargaRequired         bool `json:"harga_required"`
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
		WarnaRequired:         true,
		SizeRequired:          true,
		GrupRequired:          true,
		UnitRequired:          true,
		KatRequired:           true,
		ModelRequired:         true,
		GenderRequired:        true,
		TipeRequired:          true,
		HargaRequired:         true,
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
		WarnaRequired:         false,
		SizeRequired:          false,
		GrupRequired:          false,
		UnitRequired:          false,
		KatRequired:           false,
		ModelRequired:         false,
		GenderRequired:        false,
		TipeRequired:          false,
		HargaRequired:         false,
		TanggalProdukRequired: false,
		TanggalTerimaRequired: false,
		StatusRequired:        false,
		SupplierRequired:      false,
		DiupdateOlehRequired:  false,
	}
}

// ValidateProduct performs all validations for product creation and update
func ValidateProduct(p *models.Product, schema ValidationSchema) gin.H {
	// Validate Artikel (required and unique)
	if schema.ArtikelRequired && strings.TrimSpace(p.Artikel) == "" {
		return gin.H{
			"column":  "artikel",
			"message": "Artikel is required",
		}
	}

	// Check for duplicate artikel if not empty
	if strings.TrimSpace(p.Artikel) != "" {
		_, err := db.FetchProductByArtikelIncludeDeleted(p.Artikel)
		if err == nil || (err != nil && err.Error() != "not_found") {
			// If no error, then product exists
			// Or if there's an error but it's not "not_found", then there's a different issue
			if err == nil {
				return gin.H{
					"column":  "artikel",
					"message": "Product with this artikel already exists",
				}
			}
			// Otherwise, there was a database error
			return gin.H{
				"column":  "artikel",
				"message": "Error checking artikel uniqueness: " + err.Error(),
			}
		}
	}

	// Validate warna (required, comma-separated color IDs)
	if schema.WarnaRequired && strings.TrimSpace(p.Warna) == "" {
		return gin.H{
			"column":  "warna",
			"message": "Warna is required",
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
				return gin.H{
					"column":  "warna",
					"message": "Invalid color ID format: " + colorIDStr,
				}
			}

			_, err = db.FetchColorByID(colorID)
			if err != nil {
				if err.Error() == "not_found" {
					return gin.H{
						"column":  "warna",
						"message": "Color ID not found: " + colorIDStr,
					}
				}
				return gin.H{
					"column":  "warna",
					"message": "Error checking color ID: " + err.Error(),
				}
			}
		}
	}

	// Validate size (required, EU shoe sizes or ranges)
	if schema.SizeRequired && strings.TrimSpace(p.Size) == "" {
		return gin.H{
			"column":  "size",
			"message": "Size is required",
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
					return gin.H{
						"column":  "size",
						"message": "Invalid size range format: " + s,
					}
				}

				start, errStart := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				end, errEnd := strconv.Atoi(strings.TrimSpace(rangeParts[1]))

				if errStart != nil || errEnd != nil || start >= end {
					return gin.H{
						"column":  "size",
						"message": "Invalid size range values in: " + s,
					}
				}
			} else {
				// Single size
				_, err := strconv.Atoi(s)
				if err != nil {
					return gin.H{
						"column":  "size",
						"message": "Size must be numeric: " + s,
					}
				}
			}
		}
	}

	// Validate master data fields

	// Grup
	if schema.GrupRequired && strings.TrimSpace(p.Grup) == "" {
		return gin.H{
			"column":  "grup",
			"message": "Grup is required",
		}
	}

	if strings.TrimSpace(p.Grup) != "" {
		if errObj := common.ValidateMasterDataID("master_grups", "grup", p.Grup); errObj != nil {
			return gin.H(errObj)
		}
	}

	// Unit
	if schema.UnitRequired && strings.TrimSpace(p.Unit) == "" {
		return gin.H{
			"column":  "unit",
			"message": "Unit is required",
		}
	}

	if strings.TrimSpace(p.Unit) != "" {
		if errObj := common.ValidateMasterDataID("master_units", "unit", p.Unit); errObj != nil {
			return gin.H(errObj)
		}
	}

	// Kat
	if schema.KatRequired && strings.TrimSpace(p.Kat) == "" {
		return gin.H{
			"column":  "kat",
			"message": "Kat is required",
		}
	}

	if strings.TrimSpace(p.Kat) != "" {
		if errObj := common.ValidateMasterDataID("master_kats", "kat", p.Kat); errObj != nil {
			return gin.H(errObj)
		}
	}

	// Gender
	if schema.GenderRequired && strings.TrimSpace(p.Gender) == "" {
		return gin.H{
			"column":  "gender",
			"message": "Gender is required",
		}
	}

	if strings.TrimSpace(p.Gender) != "" {
		if errObj := common.ValidateMasterDataID("master_genders", "gender", p.Gender); errObj != nil {
			return gin.H(errObj)
		}
	}

	// Tipe
	if schema.TipeRequired && strings.TrimSpace(p.Tipe) == "" {
		return gin.H{
			"column":  "tipe",
			"message": "Tipe is required",
		}
	}

	if strings.TrimSpace(p.Tipe) != "" {
		if errObj := common.ValidateMasterDataID("master_tipes", "tipe", p.Tipe); errObj != nil {
			return gin.H(errObj)
		}
	}

	// Model (required)
	if schema.ModelRequired && strings.TrimSpace(p.Model) == "" {
		return gin.H{
			"column":  "model",
			"message": "Model is required",
		}
	}

	// Validate harga (required, numeric)
	if schema.HargaRequired && p.Harga <= 0 {
		return gin.H{
			"column":  "harga",
			"message": "Harga is required and must be a positive number",
		}
	}

	// Validate tanggal_produk (required, valid date)
	zeroTime := time.Time{}
	if schema.TanggalProdukRequired && p.TanggalProduk == zeroTime {
		return gin.H{
			"column":  "tanggal_produk",
			"message": "Tanggal produk is required",
		}
	}

	// Validate tanggal_terima (required, valid date)
	if schema.TanggalTerimaRequired && p.TanggalTerima == zeroTime {
		return gin.H{
			"column":  "tanggal_terima",
			"message": "Tanggal terima is required",
		}
	}

	// Validate status (required, must be one of "active", "inactive", or "discontinued")
	if schema.StatusRequired && strings.TrimSpace(p.Status) == "" {
		return gin.H{
			"column":  "status",
			"message": "Status is required",
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
			return gin.H{
				"column":  "status",
				"message": "Status must be either 'active', 'inactive', or 'discontinued'",
			}
		}
	}

	// Validate supplier (required)
	if schema.SupplierRequired && strings.TrimSpace(p.Supplier) == "" {
		return gin.H{
			"column":  "supplier",
			"message": "Supplier is required",
		}
	}

	// Validate diupdate_oleh (required)
	if schema.DiupdateOlehRequired && strings.TrimSpace(p.DiupdateOleh) == "" {
		return gin.H{
			"column":  "diupdate_oleh",
			"message": "Diupdate oleh is required",
		}
	}

	// All validations passed
	return nil
}

// ValidateCreate performs validation for product creation
func ValidateCreate(p *models.Product) gin.H {
	return ValidateProduct(p, DefaultCreateSchema())
}

// ValidateUpdate performs validation for product update
func ValidateUpdate(p *models.Product) gin.H {
	// Skip artikel uniqueness check for updates
	schema := DefaultUpdateSchema()
	return ValidateProduct(p, schema)
}
