package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/everysoft/inventary-be/app/master_product"
)

func CountAllProducts() (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(no) FROM master_products").Scan(&count)
	return count, err
}

func CreateMasterProductsTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_products (
			no SERIAL PRIMARY KEY,
			artikel TEXT NOT NULL,
			warna TEXT,
			size TEXT,
			grup TEXT,
			unit TEXT,
			kat TEXT,
			model TEXT,
			gender TEXT,
			tipe TEXT,
			harga NUMERIC(15,2),
			tanggal_produk DATE,
			tanggal_terima DATE,
			usia INT,
			status TEXT,
			supplier TEXT,
			diupdate_oleh TEXT,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_products_artikel ON master_products(artikel);`,
		`CREATE INDEX IF NOT EXISTS idx_master_products_grup ON master_products(grup);`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'uq_master_products_artikel'
			) THEN
				ALTER TABLE master_products ADD CONSTRAINT uq_master_products_artikel UNIQUE (artikel);
			END IF;
		END$$;`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_products table exists")
	return nil
}

func FetchAllProducts(limit, offset int) ([]master_product.Product, error) {
	products := []master_product.Product{}

	query := `
	SELECT no, artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, usia, status, supplier, diupdate_oleh, tanggal_update
	FROM master_products
	ORDER BY no
	LIMIT $1 OFFSET $2`

	rows, err := DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p master_product.Product
		if err := rows.Scan(
			&p.No, &p.Artikel, &p.Warna, &p.Size, &p.Grup, &p.Unit, &p.Kat,
			&p.Model, &p.Gender, &p.Tipe, &p.Harga, &p.TanggalProduk,
			&p.TanggalTerima, &p.Usia, &p.Status, &p.Supplier,
			&p.DiupdateOleh, &p.TanggalUpdate,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func FetchProductByArtikel(artikel string) (master_product.Product, error) {
	var p master_product.Product
	err := DB.QueryRow(`SELECT no, artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, usia, status, supplier, diupdate_oleh, tanggal_update FROM master_products WHERE artikel = $1`, artikel).
		Scan(&p.No, &p.Artikel, &p.Warna, &p.Size, &p.Grup, &p.Unit, &p.Kat, &p.Model, &p.Gender, &p.Tipe, &p.Harga, &p.TanggalProduk, &p.TanggalTerima, &p.Usia, &p.Status, &p.Supplier, &p.DiupdateOleh, &p.TanggalUpdate)

	if err == sql.ErrNoRows {
		return p, errors.New("not_found")
	}
	return p, err
}

func InsertProduct(p *master_product.Product) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_products 
		(artikel, warna, size, grup, unit, kat, model, gender, tipe, harga, tanggal_produk, tanggal_terima, usia, status, supplier, diupdate_oleh, tanggal_update) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING no`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(
		p.Artikel,
		p.Warna,
		p.Size,
		p.Grup,
		p.Unit,
		p.Kat,
		p.Model,
		p.Gender,
		p.Tipe,
		p.Harga,
		p.TanggalProduk,
		p.TanggalTerima,
		p.Usia,
		p.Status,
		p.Supplier,
		p.DiupdateOleh,
		p.TanggalUpdate,
	).Scan(&p.No)
}

func UpdateProduct(artikel string, p *master_product.Product) (master_product.Product, error) {
	result, err := DB.Exec(`
		UPDATE master_products SET
			warna = $1,
			size = $2,
			grup = $3,
			unit = $4,
			kat = $5,
			model = $6,
			gender = $7,
			tipe = $8,
			harga = $9,
			tanggal_produk = $10,
			tanggal_terima = $11,
			usia = $12,
			status = $13,
			supplier = $14,
			diupdate_oleh = $15,
			tanggal_update = $16
		WHERE artikel = $17`,
		p.Warna, p.Size, p.Grup, p.Unit, p.Kat, p.Model, p.Gender, p.Tipe, p.Harga,
		p.TanggalProduk, p.TanggalTerima, p.Usia, p.Status, p.Supplier,
		p.DiupdateOleh, p.TanggalUpdate, artikel)

	if err != nil {
		return *p, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return *p, err
	}
	if rowsAffected == 0 {
		return *p, errors.New("not_found")
	}

	// Fetch updated product
	return FetchProductByArtikel(artikel)
}

func DeleteProduct(artikel string) error {
	result, err := DB.Exec("DELETE FROM master_products WHERE artikel = $1", artikel)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("not_found")
	}

	return nil
}
