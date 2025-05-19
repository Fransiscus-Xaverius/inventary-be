package models

import (
	"time"
)

// CategoryColorLabel represents a color mapping for column values
type CategoryColorLabel struct {
	ID            int       `json:"id"`
	KodeWarna     string    `json:"kode_warna"`
	NamaWarna     string    `json:"nama_warna"`
	NamaKolom     string    `json:"nama_kolom"`
	Keterangan    string    `json:"keterangan"`
	TanggalUpdate time.Time `json:"tanggal_update"`
}
