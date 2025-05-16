package models

import (
	"time"
)

// Tipe represents a tipe record in the database
type Tipe struct {
	ID            int        `json:"id"`
	Value         string     `json:"value"`
	TanggalUpdate time.Time  `json:"tanggal_update"`
	TanggalHapus  *time.Time `json:"tanggal_hapus,omitempty"`
}
