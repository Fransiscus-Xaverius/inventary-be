package db

import (
	"log"

	"github.com/everysoft/inventary-be/app/category_color_label"
)

// CreateCategoryColorLabelsTableIfNotExists ensures the category_color_labels table exists
func CreateCategoryColorLabelsTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS category_color_labels (
			id SERIAL PRIMARY KEY,
			kode_warna TEXT,
			nama_warna TEXT,
			nama_kolom TEXT,
			keterangan TEXT,
			tanggal_update TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return err
		}
	}

	log.Println("Ensured category_color_labels table exists")
	return nil
}

// FetchAllCategoryColorLabels retrieves all category color labels from the database
func FetchAllCategoryColorLabels() ([]category_color_label.CategoryColorLabel, error) {
	labels := []category_color_label.CategoryColorLabel{}

	query := `
	SELECT 
		id, kode_warna, nama_warna, nama_kolom, keterangan, tanggal_update
	FROM category_color_labels
	ORDER BY nama_kolom, keterangan`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var label category_color_label.CategoryColorLabel
		if err := rows.Scan(
			&label.ID, &label.KodeWarna, &label.NamaWarna,
			&label.NamaKolom, &label.Keterangan, &label.TanggalUpdate,
		); err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}

	return labels, nil
}

// FetchCategoryColorLabelsByColumn retrieves category color labels filtered by column name
func FetchCategoryColorLabelsByColumn(columnName string) ([]category_color_label.CategoryColorLabel, error) {
	labels := []category_color_label.CategoryColorLabel{}

	query := `
	SELECT 
		id, kode_warna, nama_warna, nama_kolom, keterangan, tanggal_update
	FROM category_color_labels
	WHERE nama_kolom = $1
	ORDER BY keterangan`

	rows, err := DB.Query(query, columnName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var label category_color_label.CategoryColorLabel
		if err := rows.Scan(
			&label.ID, &label.KodeWarna, &label.NamaWarna,
			&label.NamaKolom, &label.Keterangan, &label.TanggalUpdate,
		); err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}

	return labels, nil
}
