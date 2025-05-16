package models

import (
	"time"
)

// Color represents a color record in the database
type Color struct {
	ID            int        `json:"id"`
	Nama          string     `json:"nama"`
	Hex           string     `json:"hex"`
	TanggalUpdate time.Time  `json:"tanggal_update"`
	TanggalHapus  *time.Time `json:"tanggal_hapus"`
}
