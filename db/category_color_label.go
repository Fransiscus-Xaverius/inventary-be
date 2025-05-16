package db

import (
	"fmt"
	"log"

	"github.com/everysoft/inventary-be/app/models"
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
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'uq_category_color_labels_column_desc'
			) THEN
				ALTER TABLE category_color_labels ADD CONSTRAINT uq_category_color_labels_column_desc UNIQUE (nama_kolom, keterangan);
			END IF;
		END$$;`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured category_color_labels table exists")
	return nil
}

// FetchAllCategoryColorLabels retrieves all category color labels from the database
func FetchAllCategoryColorLabels() ([]models.CategoryColorLabel, error) {
	labels := []models.CategoryColorLabel{}

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
		var label models.CategoryColorLabel
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
func FetchCategoryColorLabelsByColumn(columnName string) ([]models.CategoryColorLabel, error) {
	labels := []models.CategoryColorLabel{}

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
		var label models.CategoryColorLabel
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

// FetchCategoryColorLabelByColumnAndValue retrieves a specific category color label by column name and category value
func FetchCategoryColorLabelByColumnAndValue(columnName, categoryValue string) (*models.CategoryColorLabel, error) {
	var label models.CategoryColorLabel

	query := `
	SELECT 
		id, kode_warna, nama_warna, nama_kolom, keterangan, tanggal_update
	FROM category_color_labels
	WHERE nama_kolom = $1 AND LOWER(keterangan) = LOWER($2)
	LIMIT 1`

	row := DB.QueryRow(query, columnName, categoryValue)
	err := row.Scan(
		&label.ID, &label.KodeWarna, &label.NamaWarna,
		&label.NamaKolom, &label.Keterangan, &label.TanggalUpdate,
	)
	if err != nil {
		return nil, err
	}

	return &label, nil
}
