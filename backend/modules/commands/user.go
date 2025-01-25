package commands

import (
	"fmt"
	"strings"
	"time"
	"uvcs/modules/db"
	"uvcs/modules/utils"
)

// CreateUser creates a new user with the given credentials
func CreateUser(username, password, firstname, lastname string) error {
	// Generate security keys
	skey1 := utils.GenerateSecurityKey()
	skey2 := utils.GenerateSecurityKey()

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	// Create user
	_, err = db.DB.Exec(`
		INSERT INTO users (username, password, firstname, lastname, skey1, skey2, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, true)
	`, username, hashedPassword, firstname, lastname, skey1, skey2)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	fmt.Printf("Successfully created user '%s' (%s %s)\n", username, firstname, lastname)
	return nil
}

// EnableUser enables a user by their username
func EnableUser(username string) error {
	result, err := db.DB.Exec(`
		UPDATE users 
		SET is_active = true 
		WHERE username = $1 AND is_active = false
	`, username)
	if err != nil {
		return fmt.Errorf("error enabling user: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("user '%s' not found or already enabled", username)
	}

	fmt.Printf("Successfully enabled user '%s'\n", username)
	return nil
}

// DisableUser disables a user by their username
func DisableUser(username string) error {
	result, err := db.DB.Exec(`
		UPDATE users 
		SET is_active = false 
		WHERE username = $1 AND is_active = true
	`, username)
	if err != nil {
		return fmt.Errorf("error disabling user: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("user '%s' not found or already disabled", username)
	}

	fmt.Printf("Successfully disabled user '%s'\n", username)
	return nil
}

// ListUsers lists all users with their status
func ListUsers() error {
	rows, err := db.DB.Query(`
		SELECT username, firstname, lastname, created_at, is_active 
		FROM users 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return fmt.Errorf("error querying users: %v", err)
	}
	defer rows.Close()

	fmt.Printf("\nUsers:\n")
	fmt.Printf("%-20s %-20s %-20s %-20s %s\n", "Username", "First Name", "Last Name", "Created At", "Status")
	fmt.Println(strings.Repeat("-", 100))

	for rows.Next() {
		var (
			username  string
			firstname string
			lastname  string
			createdAt time.Time
			isActive  bool
		)
		err := rows.Scan(&username, &firstname, &lastname, &createdAt, &isActive)
		if err != nil {
			return fmt.Errorf("error scanning user: %v", err)
		}

		status := "Disabled"
		if isActive {
			status = "Enabled"
		}

		fmt.Printf("%-20s %-20s %-20s %-20s %s\n",
			username,
			firstname,
			lastname,
			createdAt.Format("2006-01-02 15:04:05"),
			status,
		)
	}
	fmt.Println()
	return nil
}
