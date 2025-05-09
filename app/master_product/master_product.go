package master_product

import (
	"time"
)

type Product struct {
	ID            string     `json:"id"`
	Artikel       string     `json:"artikel"`        // ARTIKEL = PRODUCT_NAME
	No            string     `json:"no"`             // NO
	Warna         string     `json:"warna"`          // WARNA
	Size          string     `json:"size"`           // SIZE
	Grup          string     `json:"grup"`           // GRUP
	Unit          string     `json:"unit"`           // UNIT
	Kat           string     `json:"kat"`            // KAT
	Model         string     `json:"model"`          // MODEL
	Gender        string     `json:"gender"`         // GENDER
	Tipe          string     `json:"tipe"`           // TIPE
	Harga         float64    `json:"harga"`          // HARGA
	TanggalProduk time.Time  `json:"tanggal_produk"` // TANGGAL PRODUK
	TanggalTerima time.Time  `json:"tanggal_terima"` // TANGGAL TERIMA
	Usia          int        `json:"usia"`           // USIA
	Status        string     `json:"status"`         // STATUS
	Supplier      string     `json:"supplier"`       // SUPPLIER
	DiupdateOleh  string     `json:"diupdate_oleh"`  // DIUPDATE OLEH
	TanggalUpdate time.Time  `json:"tanggal_update"` // TANGGAL UPDATE
	TanggalHapus  *time.Time `json:"tanggal_hapus"`  // TANGGAL HAPUS - Null if product is active, contains timestamp when soft-deleted
}
