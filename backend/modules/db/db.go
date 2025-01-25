package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}

	createTables()
}

func createTables() {
	// Create users table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			firstname VARCHAR(100) NOT NULL,
			lastname VARCHAR(100) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create active_logins table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS active_logins (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			ip_address VARCHAR(45) NOT NULL,
			skey1 VARCHAR(64) NOT NULL,
			skey2 VARCHAR(64) NOT NULL,
			login_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_action_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create commit_history table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS commit_history (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			commit_hash VARCHAR(128) NOT NULL,
			commit_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			commit_message TEXT NOT NULL,
			tags VARCHAR(128)[] DEFAULT ARRAY[]::VARCHAR(128)[],
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create commit_details table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS commit_details (
			id SERIAL PRIMARY KEY,
			commit_id INTEGER REFERENCES commit_history(id),
			file_path TEXT NOT NULL,
			change_type CHAR(1) NOT NULL CHECK (change_type IN ('M', 'D', 'A')),
			content_changes JSONB NOT NULL,
			FOREIGN KEY (commit_id) REFERENCES commit_history(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create branches table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS branches (
			id SERIAL PRIMARY KEY,
			name VARCHAR(128) NOT NULL UNIQUE,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			commit_ids INTEGER[] DEFAULT ARRAY[]::INTEGER[],
			head_commit_id INTEGER REFERENCES commit_history(id),
			is_active BOOLEAN DEFAULT true,
			FOREIGN KEY (head_commit_id) REFERENCES commit_history(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create index on commit_id for faster lookups
	_, err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_commit_details_commit_id ON commit_details(commit_id)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create index on branch name for faster lookups
	_, err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_branches_name ON branches(name)
	`)
	if err != nil {
		log.Fatal(err)
	}
}
