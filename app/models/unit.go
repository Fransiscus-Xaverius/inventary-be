package models

import (
	"time"
)

// Unit represents a unit record in the database
type Unit struct {
	ID            int        `json:"id"`
	Value         string     `json:"value"`
	TanggalUpdate time.Time  `json:"tanggal_update"`
	TanggalHapus  *time.Time `json:"tanggal_hapus,omitempty"`
}
