package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/everysoft/inventary-be/app/models"
)

// CreateBannersTableIfNotExists ensures the banners table exists
func CreateBannersTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS banners (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			cta_text TEXT,
			cta_link TEXT,
			image_url TEXT,
			order_index INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMPTZ DEFAULT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_banners_order_index ON banners(order_index);`,
		`CREATE INDEX IF NOT EXISTS idx_banners_is_active ON banners(is_active);`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured banners table exists")
	return nil
}

// CountAllBanners counts all available banners matching the search query
func CountAllBanners(queryStr string, isActive *bool) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM banners WHERE deleted_at IS NULL"
	args := []interface{}{} // Use a slice for args
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR title ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR description ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add isActive filter
	if isActive != nil {
		baseQuery += fmt.Sprintf(" AND is_active = $%d", paramCount)
		args = append(args, *isActive)
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

// FetchAllBanners retrieves all banners matching the search query with pagination
func FetchAllBanners(limit, offset int, queryStr string, isActive *bool, sortColumn string, sortDirection string) ([]models.Banner, error) {
	banners := []models.Banner{}

	baseQuery := `
	SELECT
		id, title, description, cta_text, cta_link, image_url, order_index, is_active, created_at, updated_at, deleted_at
	FROM banners
	WHERE deleted_at IS NULL`

	args := []interface{}{} // Use a slice for args
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR title ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR description ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add isActive filter
	if isActive != nil {
		baseQuery += fmt.Sprintf(" AND is_active = $%d", paramCount)
		args = append(args, *isActive)
		paramCount++
	}

	// Add sorting
	orderBy := " ORDER BY "
	validColumns := map[string]bool{
		"id": true, "title": true, "order_index": true, "is_active": true, "created_at": true, "updated_at": true,
	}

	// Default sort
	if sortColumn == "" || !validColumns[sortColumn] {
		sortColumn = "order_index"
	}

	// Default direction
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "asc"
	}

	orderBy += sortColumn + " " + sortDirection

	// Add pagination
	paginationQuery := orderBy + ` LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	baseQuery += paginationQuery
	args = append(args, limit, offset)

	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var b models.Banner
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Description, &b.CtaText, &b.CtaLink, &b.ImageUrl, &b.OrderIndex, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
		); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}

	return banners, nil
}

// FetchBannerByID retrieves a banner by its ID
func FetchBannerByID(id int) (models.Banner, error) {
	var b models.Banner
	err := DB.QueryRow(`SELECT id, title, description, cta_text, cta_link, image_url, order_index, is_active, created_at, updated_at, deleted_at FROM banners WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&b.ID, &b.Title, &b.Description, &b.CtaText, &b.CtaLink, &b.ImageUrl, &b.OrderIndex, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt)

	if err == sql.ErrNoRows {
		return b, errors.New("not_found")
	}
	return b, err
}

// InsertBanner inserts a new banner record
func InsertBanner(b *models.Banner) error {
	stmt, err := DB.Prepare(`
		INSERT INTO banners
		(title, description, cta_text, cta_link, image_url, order_index, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(
		b.Title,
		b.Description,
		b.CtaText,
		b.CtaLink,
		b.ImageUrl,
		b.OrderIndex,
		b.IsActive,
		time.Now(),
		time.Now(),
	).Scan(&b.ID)
}

// UpdateBanner updates an existing banner record
func UpdateBanner(id int, b *models.Banner) (models.Banner, error) {
	// First fetch the existing banner to get current values
	currentBanner, err := FetchBannerByID(id)
	if err != nil {
		return *b, err
	}

	// Build dynamic query with only fields that need to be updated
	query := "UPDATE banners SET"
	args := []interface{}{} // Use a slice for args
	paramCount := 1
	fieldsToUpdate := 0

	// Check each field and only update if it's not empty in the request
	if b.Title != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" title = $%d,", paramCount)
		args = append(args, b.Title)
		paramCount++
	}
	if b.Description != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" description = $%d,", paramCount)
		args = append(args, b.Description)
		paramCount++
	}
	if b.CtaText != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" cta_text = $%d,", paramCount)
		args = append(args, b.CtaText)
		paramCount++
	}
	if b.CtaLink != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" cta_link = $%d,", paramCount)
		args = append(args, b.CtaLink)
		paramCount++
	}
	if b.ImageUrl != "" {
		fieldsToUpdate++
		query += fmt.Sprintf(" image_url = $%d,", paramCount)
		args = append(args, b.ImageUrl)
		paramCount++
	}
	// For int and bool fields, check if they are provided (not zero-value or default-value if that implies no change)
	// For simplicity, assuming if they are in the request body, they should be updated.
	// A more robust solution might involve pointers or checking for explicit presence in JSON.
	// For now, we'll update if the value is different from the current one.
	if b.OrderIndex != currentBanner.OrderIndex && b.OrderIndex != 0 {
		fieldsToUpdate++
		query += fmt.Sprintf(" order_index = $%d,", paramCount)
		args = append(args, b.OrderIndex)
		paramCount++
	}
	if b.IsActive != currentBanner.IsActive {
		fieldsToUpdate++
		query += fmt.Sprintf(" is_active = $%d,", paramCount)
		args = append(args, b.IsActive)
		paramCount++
	}

	// Always update updated_at
	fieldsToUpdate++
	query += fmt.Sprintf(" updated_at = $%d,", paramCount)
	args = append(args, time.Now())
	paramCount++

	// Remove the trailing comma and complete the query
	query = query[:len(query)-1] + " WHERE id = $" + fmt.Sprintf("%d", paramCount)
	args = append(args, id)

	// If no fields to update (other than updated_at), return the current banner
	if fieldsToUpdate == 1 { // Only updated_at was added
		return currentBanner, nil
	}

	result, err := DB.Exec(query, args...)
	if err != nil {
		return *b, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return *b, err
	}
	if rowsAffected == 0 {
		return *b, errors.New("not_found")
	}

	// Fetch updated banner
	return FetchBannerByID(id)
}

// DeleteBanner soft-deletes a banner
func DeleteBanner(id int) error {
	currentTime := time.Now()
	result, err := DB.Exec("UPDATE banners SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL", currentTime, id)
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

// RestoreBanner restores a soft-deleted banner
func RestoreBanner(id int) error {
	result, err := DB.Exec("UPDATE banners SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL", id)
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

// CountDeletedBanners counts all soft-deleted banners
func CountDeletedBanners(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM banners WHERE deleted_at IS NOT NULL"
	args := []interface{}{} // Use a slice for args
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR title ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR description ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

// FetchDeletedBanners retrieves all soft-deleted banners
func FetchDeletedBanners(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Banner, error) {
	banners := []models.Banner{}

	baseQuery := `
	SELECT
		id, title, description, cta_text, cta_link, image_url, order_index, is_active, created_at, updated_at, deleted_at
	FROM banners
	WHERE deleted_at IS NOT NULL`

	args := []interface{}{} // Use a slice for args
	paramCount := 1

	// Add search condition if query string is provided
	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR title ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR description ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	// Add sorting
	orderBy := " ORDER BY "
	validColumns := map[string]bool{
		"id": true, "title": true, "order_index": true, "is_active": true, "created_at": true, "updated_at": true, "deleted_at": true,
	}

	// Default sort
	if sortColumn == "" || !validColumns[sortColumn] {
		sortColumn = "deleted_at"
	}

	// Default direction
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	orderBy += sortColumn + " " + sortDirection

	// Add pagination
	paginationQuery := orderBy + ` LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	baseQuery += paginationQuery
	args = append(args, limit, offset)

	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var b models.Banner
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Description, &b.CtaText, &b.CtaLink, &b.ImageUrl, &b.OrderIndex, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
		); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}

	return banners, nil
}
