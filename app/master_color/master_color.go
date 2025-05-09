package master_color

import (
	"time"
)

type Color struct {
	ID            int        `json:"id"`
	Nama          string     `json:"nama"`
	Hex           string     `json:"hex"`
	TanggalUpdate time.Time  `json:"tanggal_update"`
	TanggalHapus  *time.Time `json:"tanggal_hapus"`
}
