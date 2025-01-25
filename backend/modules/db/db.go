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
}
