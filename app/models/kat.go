package models

import (
	"time"
)

// Kat represents a category record in the database
type Kat struct {
	ID            int        `json:"id"`
	Value         string     `json:"value"`
	TanggalUpdate time.Time  `json:"tanggal_update"`
	TanggalHapus  *time.Time `json:"tanggal_hapus,omitempty"`
}
