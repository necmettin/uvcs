package commands

import (
	"fmt"
	"strings"
	"time"
	"uvcs/modules/db"
	"uvcs/modules/utils"
)

// CreateUser creates a new user with the given credentials
func CreateUser(identifier, identifierType, password, firstname, lastname string) error {
	// Generate security keys
	skey1 := utils.GenerateSecurityKey()
	skey2 := utils.GenerateSecurityKey()

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	var query string
	var args []interface{}

	if identifierType == "email" {
		query = `
			INSERT INTO users (email, password, firstname, lastname, skey1, skey2, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, true)`
		args = []interface{}{identifier, hashedPassword, firstname, lastname, skey1, skey2}
	} else {
		query = `
			INSERT INTO users (username, password, firstname, lastname, skey1, skey2, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, true)`
		args = []interface{}{identifier, hashedPassword, firstname, lastname, skey1, skey2}
	}

	// Create user
	_, err = db.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	fmt.Printf("Successfully created user '%s' (%s %s)\n", identifier, firstname, lastname)
	return nil
}

// EnableUser enables a user by their identifier (username or email)
func EnableUser(identifier string) error {
	result, err := db.DB.Exec(`
		UPDATE users 
		SET is_active = true 
		WHERE (username = $1 OR email = $1) AND is_active = false
	`, identifier)
	if err != nil {
		return fmt.Errorf("error enabling user: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("user '%s' not found or already enabled", identifier)
	}

	fmt.Printf("Successfully enabled user '%s'\n", identifier)
	return nil
}

// DisableUser disables a user by their identifier (username or email)
func DisableUser(identifier string) error {
	result, err := db.DB.Exec(`
		UPDATE users 
		SET is_active = false 
		WHERE (username = $1 OR email = $1) AND is_active = true
	`, identifier)
	if err != nil {
		return fmt.Errorf("error disabling user: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("user '%s' not found or already disabled", identifier)
	}

	fmt.Printf("Successfully disabled user '%s'\n", identifier)
	return nil
}

// ListUsers lists all users with their status
func ListUsers() error {
	rows, err := db.DB.Query(`
		SELECT 
			COALESCE(username, email) as identifier,
			firstname, 
			lastname, 
			created_at, 
			is_active,
			username IS NOT NULL as has_username,
			email IS NOT NULL as has_email
		FROM users 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return fmt.Errorf("error querying users: %v", err)
	}
	defer rows.Close()

	fmt.Printf("\nUsers:\n")
	fmt.Printf("%-30s %-20s %-20s %-20s %-10s %s\n",
		"Identifier", "First Name", "Last Name", "Created At", "Status", "Type")
	fmt.Println(strings.Repeat("-", 120))

	for rows.Next() {
		var (
			identifier  string
			firstname   string
			lastname    string
			createdAt   time.Time
			isActive    bool
			hasUsername bool
			hasEmail    bool
		)
		err := rows.Scan(&identifier, &firstname, &lastname, &createdAt, &isActive, &hasUsername, &hasEmail)
		if err != nil {
			return fmt.Errorf("error scanning user: %v", err)
		}

		status := "Disabled"
		if isActive {
			status = "Enabled"
		}

		idType := "email"
		if hasUsername {
			idType = "username"
		}

		fmt.Printf("%-30s %-20s %-20s %-20s %-10s %s\n",
			identifier,
			firstname,
			lastname,
			createdAt.Format("2006-01-02 15:04:05"),
			status,
			idType,
		)
	}
	fmt.Println()
	return nil
}
