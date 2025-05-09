package master_color

import (
	"time"
)

type Color struct {
	ID            int        `json:"id"`
	Nama          string     `json:"nama"`
	TanggalUpdate time.Time  `json:"tanggal_update"`
	TanggalHapus  *time.Time `json:"tanggal_hapus"`
}
