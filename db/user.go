package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/everysoft/inventary-be/app/auth"
)

// CreateUser inserts a new user into the database
func CreateUser(user auth.User) error {
	// Use a transaction for safety
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert user
	query := `
		INSERT INTO users (id, username, email, password, role)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(query, user.ID, user.Username, user.Email, user.Password, user.Role)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return tx.Commit()
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (auth.User, error) {
	var user auth.User
	query := `
		SELECT id, username, email, password, role
		FROM users
		WHERE username = $1
	`
	err := DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.User{}, fmt.Errorf("user not found: %w", err)
		}
		return auth.User{}, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (auth.User, error) {
	var user auth.User
	query := `
		SELECT id, username, email, password, role
		FROM users
		WHERE email = $1
	`
	err := DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.User{}, fmt.Errorf("user not found: %w", err)
		}
		return auth.User{}, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(id string) (auth.User, error) {
	var user auth.User
	query := `
		SELECT id, username, email, password, role
		FROM users
		WHERE id = $1
	`
	err := DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.User{}, fmt.Errorf("user not found: %w", err)
		}
		return auth.User{}, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}

// UserExists checks if a user with the given username or email already exists
func UserExists(username, email string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE username = $1 OR email = $2
		)
	`
	err := DB.QueryRow(query, username, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if user exists: %w", err)
	}
	
	return exists, nil
}

// CreateUsersTableIfNotExists creates the users table if it doesn't exist
func CreateUsersTableIfNotExists() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`
	
	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	
	log.Println("Ensured users table exists")
	return nil
}