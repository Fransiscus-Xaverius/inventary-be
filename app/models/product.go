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

// ProductRating represents admin-set specifications/stats for a product
type ProductRating struct {
	Comfort int      `json:"comfort" validate:"min=0,max=10"`
	Style   int      `json:"style" validate:"min=0,max=10"`
	Support int      `json:"support" validate:"min=0,max=10"`
	Purpose []string `json:"purpose" validate:"min=1"`
}

type MarketplaceInfo struct {
	Tokopedia *string `json:"tokopedia" optional:"true"`
	Shopee    *string `json:"shopee" optional:"true"`
	Lazada    *string `json:"lazada" optional:"true"`
	Tiktok    *string `json:"tiktok" optional:"true"`
	Bukalapak *string `json:"bukalapak" optional:"true"`
}

type CreateProductRequest struct {
	Artikel     string   `form:"artikel" binding:"required"`
	Nama        string   `form:"nama" binding:"required"`
	Deskripsi   string   `form:"deskripsi" binding:"required"`
	Warna       string   `form:"warna" binding:"required"` // Comma-separated IDs
	Size        string   `form:"size" binding:"required"`
	Grup        string   `form:"grup" binding:"required"`
	Unit        string   `form:"unit" binding:"required"`
	Kat         string   `form:"kat" binding:"required"`
	Model       string   `form:"model" binding:"required"`
	Gender      string   `form:"gender" binding:"required"`
	Tipe        string   `form:"tipe" binding:"required"`
	Harga       float64  `form:"harga" binding:"required,gt=0"`
	HargaDiskon *float64 `form:"harga_diskon"`
	Rating      string   `form:"rating"`      // JSON string
	Marketplace string   `form:"marketplace"` // JSON string
	// Gambar        []*multipart.FileHeader `form:"gambar"`
	TanggalProduk string `form:"tanggal_produk"`
	TanggalTerima string `form:"tanggal_terima"`
	Status        string `form:"status" binding:"required"`
	Supplier      string `form:"supplier" binding:"required"`
	DiupdateOleh  string `form:"diupdate_oleh" binding:"required"`
}

type Product struct {
	Artikel       string          `json:"artikel"`          // ARTIKEL = PRODUCT_NAME
	Nama          string          `json:"nama"`             // NAMA
	Deskripsi     string          `json:"deskripsi"`        // DESKRIPSI
	Rating        ProductRating   `json:"rating"`           // RATING - Admin-set specifications
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
	HargaDiskon   *float64        `json:"harga_diskon"`     // HARGA DISKON
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
