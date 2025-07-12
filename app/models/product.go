package models

import (
	"time"
)

// ColorInfo represents color information from master_colors
type ColorInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

type MarketplaceInfo struct {
	Tokopedia string `json:"tokopedia"`
	Shopee    string `json:"shopee"`
	Lazada    string `json:"lazada"`
	Tiktok    string `json:"tiktok"`
	Bukalapak string `json:"bukalapak"`
}

type Product struct {
	Artikel       string          `json:"artikel"`          // ARTIKEL = PRODUCT_NAME
	Nama          string          `json:"nama"`             // NAMA
	Deskripsi     string          `json:"deskripsi"`        // DESKRIPSI
	Rating        float64         `json:"rating"`           // RATING
	No            string          `json:"no"`               // NO
	Warna         string          `json:"warna"`            // WARNA - Now stores comma-separated IDs
	Size          string          `json:"size"`             // SIZE
	Grup          string          `json:"grup"`             // GRUP
	Unit          string          `json:"unit"`             // UNIT
	Kat           string          `json:"kat"`              // KAT
	Model         string          `json:"model"`            // MODEL
	Gender        string          `json:"gender"`           // GENDER
	Tipe          string          `json:"tipe"`             // TIPE
	Harga         float64         `json:"harga"`            // HARGA
	HargaDiskon   float64         `json:"harga_diskon"`     // HARGA DISKON
	Marketplace   MarketplaceInfo `json:"marketplace"`      // MARKETPLACE
	Gambar        []string        `json:"gambar"`           // GAMBAR
	TanggalProduk time.Time       `json:"tanggal_produk"`   // TANGGAL PRODUK
	TanggalTerima time.Time       `json:"tanggal_terima"`   // TANGGAL TERIMA
	Usia          string          `json:"usia,omitempty"`   // Calculated dynamically: "Fresh" under 1 year, "Normal" under 2 years, "Aging" over 2 years
	Status        string          `json:"status"`           // STATUS
	Supplier      string          `json:"supplier"`         // SUPPLIER
	DiupdateOleh  string          `json:"diupdate_oleh"`    // DIUPDATE OLEH
	TanggalUpdate time.Time       `json:"tanggal_update"`   // TANGGAL UPDATE
	TanggalHapus  *time.Time      `json:"tanggal_hapus"`    // TANGGAL HAPUS - Null if product is active, contains timestamp when soft-deleted
	Colors        []ColorInfo     `json:"colors,omitempty"` // Additional color information
}
