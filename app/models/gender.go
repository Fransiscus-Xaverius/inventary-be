package models

import (
	"time"
)

// Gender represents a gender record in the database
type Gender struct {
	ID            int        `json:"id"`
	Value         string     `json:"value"`
	TanggalUpdate time.Time  `json:"tanggal_update"`
	TanggalHapus  *time.Time `json:"tanggal_hapus,omitempty"`
}
