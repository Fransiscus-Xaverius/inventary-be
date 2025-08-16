package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/everysoft/inventary-be/app/models"
)

// CreateMasterNewsletterTableIfNotExists ensures the master_newsletter table exists
func CreateMasterNewsletterTableIfNotExists() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS master_newsletter (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL,
			whatsapp TEXT NOT NULL,
			message TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMPTZ DEFAULT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_master_newsletter_email ON master_newsletter(email);`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	log.Println("Ensured master_newsletter table exists")
	return nil
}

// InsertNewsletter inserts a new newsletter subscription
func InsertNewsletter(n *models.Newsletter) error {
	stmt, err := DB.Prepare(`
		INSERT INTO master_newsletter
		(email, whatsapp, message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRow(
		n.Email,
		n.Whatsapp,
		n.Message,
		time.Now(),
		time.Now(),
	).Scan(&n.ID)
}

// CountAllNewsletters counts all available newsletter entries matching the search query
func CountAllNewsletters(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_newsletter WHERE deleted_at IS NULL"
	args := []interface{}{}
	paramCount := 1

	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR email ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR whatsapp ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR message ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

// FetchAllNewsletters retrieves all newsletter entries matching the search query with pagination
func FetchAllNewsletters(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Newsletter, error) {
	newsletters := []models.Newsletter{}

	baseQuery := `
	SELECT
		id, email, whatsapp, message, created_at, updated_at, deleted_at
	FROM master_newsletter
	WHERE deleted_at IS NULL`

	args := []interface{}{}
	paramCount := 1

	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR email ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR whatsapp ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR message ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	orderBy := " ORDER BY "
	validColumns := map[string]bool{
		"id": true, "email": true, "whatsapp": true, "created_at": true, "updated_at": true,
	}

	if sortColumn == "" || !validColumns[sortColumn] {
		sortColumn = "created_at"
	}

	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	orderBy += sortColumn + " " + sortDirection

	paginationQuery := orderBy + ` LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	baseQuery += paginationQuery
	args = append(args, limit, offset)

	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n models.Newsletter
		if err := rows.Scan(
			&n.ID, &n.Email, &n.Whatsapp, &n.Message, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt,
		); err != nil {
			return nil, err
		}
		newsletters = append(newsletters, n)
	}

	return newsletters, nil
}

// FetchNewsletterByID retrieves a newsletter entry by its ID
func FetchNewsletterByID(id int) (models.Newsletter, error) {
	var n models.Newsletter
	err := DB.QueryRow(`SELECT id, email, whatsapp, message, created_at, updated_at, deleted_at FROM master_newsletter WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&n.ID, &n.Email, &n.Whatsapp, &n.Message, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt)

	if err == sql.ErrNoRows {
		return n, errors.New("not_found")
	}
	return n, err
}

// UpdateNewsletter updates an existing newsletter entry
func UpdateNewsletter(id int, n *models.Newsletter) (models.Newsletter, error) {
	query := "UPDATE master_newsletter SET email = $1, whatsapp = $2, message = $3, updated_at = $4 WHERE id = $5"
	_, err := DB.Exec(query, n.Email, n.Whatsapp, n.Message, time.Now(), id)
	if err != nil {
		return *n, err
	}
	return FetchNewsletterByID(id)
}

// DeleteNewsletter soft-deletes a newsletter entry
func DeleteNewsletter(id int) error {
	currentTime := time.Now()
	result, err := DB.Exec("UPDATE master_newsletter SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL", currentTime, id)
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

// RestoreNewsletter restores a soft-deleted newsletter entry
func RestoreNewsletter(id int) error {
	result, err := DB.Exec("UPDATE master_newsletter SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL", id)
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

// CountDeletedNewsletters counts all soft-deleted newsletter entries
func CountDeletedNewsletters(queryStr string) (int, error) {
	baseQuery := "SELECT COUNT(id) FROM master_newsletter WHERE deleted_at IS NOT NULL"
	args := []interface{}{}
	paramCount := 1

	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR email ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	var count int
	err := DB.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

// FetchDeletedNewsletters retrieves all soft-deleted newsletter entries
func FetchDeletedNewsletters(limit, offset int, queryStr string, sortColumn string, sortDirection string) ([]models.Newsletter, error) {
	newsletters := []models.Newsletter{}

	baseQuery := `
	SELECT
		id, email, whatsapp, message, created_at, updated_at, deleted_at
	FROM master_newsletter
	WHERE deleted_at IS NOT NULL`

	args := []interface{}{}
	paramCount := 1

	if queryStr != "" {
		searchQuery := ` AND (
			CAST(id AS TEXT) ILIKE $` + fmt.Sprintf("%d", paramCount) + `
			OR email ILIKE $` + fmt.Sprintf("%d", paramCount) + `
		)`
		baseQuery += searchQuery
		args = append(args, "%"+queryStr+"%")
		paramCount++
	}

	orderBy := " ORDER BY "
	validColumns := map[string]bool{
		"id": true, "email": true, "deleted_at": true,
	}

	if sortColumn == "" || !validColumns[sortColumn] {
		sortColumn = "deleted_at"
	}

	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	orderBy += sortColumn + " " + sortDirection

	paginationQuery := orderBy + ` LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	baseQuery += paginationQuery
	args = append(args, limit, offset)

	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n models.Newsletter
		if err := rows.Scan(
			&n.ID, &n.Email, &n.Whatsapp, &n.Message, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt,
		); err != nil {
			return nil, err
		}
		newsletters = append(newsletters, n)
	}

	return newsletters, nil
}
