package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB initializes the database connection and schema
func InitDB(isCLI bool) error {
	config, err := loadConfig(isCLI)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	// Connect to database
	DB, err = sql.Open(config.Type, config.GetDSN())
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	// Test connection
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	// Initialize schema
	err = initSchema(config.Type)
	if err != nil {
		return fmt.Errorf("error initializing schema: %v", err)
	}

	log.Printf("Connected to %s database successfully", config.Type)
	return nil
}

func initSchema(dbType string) error {
	var autoIncrementSyntax string
	var timestampSyntax string
	var booleanType string

	if dbType == "sqlite" {
		autoIncrementSyntax = "INTEGER PRIMARY KEY AUTOINCREMENT"
		timestampSyntax = "DATETIME DEFAULT CURRENT_TIMESTAMP"
		booleanType = "INTEGER"
	} else {
		autoIncrementSyntax = "SERIAL PRIMARY KEY"
		timestampSyntax = "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"
		booleanType = "BOOLEAN"
	}

	// Create users table
	_, err := DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS users (
			id %s,
			username VARCHAR(128) UNIQUE NOT NULL,
			password VARCHAR(256) NOT NULL,
			firstname VARCHAR(128) NOT NULL,
			lastname VARCHAR(128) NOT NULL,
			skey1 VARCHAR(128) NOT NULL,
			skey2 VARCHAR(128) NOT NULL,
			created_at %s,
			is_active %s DEFAULT false
		)
	`, autoIncrementSyntax, timestampSyntax, booleanType))
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	// Create commit_history table
	_, err = DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS commit_history (
			id %s,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			commit_hash VARCHAR(128) NOT NULL,
			commit_datetime %s,
			commit_message TEXT NOT NULL,
			tags VARCHAR(128)[] DEFAULT ARRAY[]::VARCHAR(128)[]
		)
	`, autoIncrementSyntax, timestampSyntax))
	if err != nil {
		return fmt.Errorf("error creating commit_history table: %v", err)
	}

	// Create commit_details table
	_, err = DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS commit_details (
			id %s,
			commit_id INTEGER REFERENCES commit_history(id) ON DELETE CASCADE,
			file_path TEXT NOT NULL,
			change_type CHAR(1) NOT NULL CHECK (change_type IN ('M', 'D', 'A')),
			content_changes JSONB NOT NULL
		)
	`, autoIncrementSyntax))
	if err != nil {
		return fmt.Errorf("error creating commit_details table: %v", err)
	}

	// Create branches table
	_, err = DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS branches (
			id %s,
			name VARCHAR(128) NOT NULL UNIQUE,
			description TEXT,
			created_at %s,
			commit_ids INTEGER[] DEFAULT ARRAY[]::INTEGER[],
			head_commit_id INTEGER REFERENCES commit_history(id) ON DELETE SET NULL,
			is_active %s DEFAULT true
		)
	`, autoIncrementSyntax, timestampSyntax, booleanType))
	if err != nil {
		return fmt.Errorf("error creating branches table: %v", err)
	}

	// Create indexes
	_, err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_commit_details_commit_id ON commit_details(commit_id);
		CREATE INDEX IF NOT EXISTS idx_branches_name ON branches(name);
	`)
	if err != nil {
		return fmt.Errorf("error creating indexes: %v", err)
	}

	return nil
}
