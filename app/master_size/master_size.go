package master_size

import (
	"time"
)

// Size represents a shoe size in the system
type Size struct {
	ID            int        `json:"id"`
	Value         string     `json:"value"`
	Unit          string     `json:"unit"`
	TanggalUpdate time.Time  `json:"tanggal_update"`
	TanggalHapus  *time.Time `json:"tanggal_hapus,omitempty"`
}
