package models

import (
	"time"
)

// Grup represents a grup record in the database
type Grup struct {
	ID            int        `json:"id"`
	Value         string     `json:"value"`
	TanggalUpdate time.Time  `json:"tanggal_update"`
	TanggalHapus  *time.Time `json:"tanggal_hapus,omitempty"`
}
