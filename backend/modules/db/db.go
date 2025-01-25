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
			username VARCHAR(128) UNIQUE,
			email VARCHAR(256) UNIQUE NOT NULL,
			password VARCHAR(256) NOT NULL,
			firstname VARCHAR(128) NOT NULL,
			lastname VARCHAR(128) NOT NULL,
			skey1 VARCHAR(128) NOT NULL,
			skey2 VARCHAR(128) NOT NULL,
			created_at %s,
			is_active %s DEFAULT false,
			CHECK (username IS NOT NULL OR email IS NOT NULL)
		)
	`, autoIncrementSyntax, timestampSyntax, booleanType))
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	// Create repositories table
	_, err = DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS repositories (
			id %s,
			name VARCHAR(128) NOT NULL,
			description TEXT,
			owner_id INTEGER REFERENCES users(id) ON DELETE RESTRICT,
			created_at %s,
			is_active %s DEFAULT true,
			UNIQUE(owner_id, name)
		)
	`, autoIncrementSyntax, timestampSyntax, booleanType))
	if err != nil {
		return fmt.Errorf("error creating repositories table: %v", err)
	}

	// Create repository access table
	_, err = DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS repository_access (
			id %s,
			repository_id INTEGER REFERENCES repositories(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			access_level VARCHAR(10) NOT NULL CHECK (access_level IN ('read', 'write')),
			granted_at %s,
			granted_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
			UNIQUE(repository_id, user_id)
		)
	`, autoIncrementSyntax, timestampSyntax))
	if err != nil {
		return fmt.Errorf("error creating repository_access table: %v", err)
	}

	// Create commit_history table with repository_id
	_, err = DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS commit_history (
			id %s,
			repository_id INTEGER REFERENCES repositories(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			commit_hash VARCHAR(128) NOT NULL,
			commit_datetime %s,
			commit_message TEXT NOT NULL,
			tags VARCHAR(128)[] DEFAULT ARRAY[]::VARCHAR(128)[],
			UNIQUE(repository_id, commit_hash)
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

	// Create branches table with repository_id
	_, err = DB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS branches (
			id %s,
			repository_id INTEGER REFERENCES repositories(id) ON DELETE CASCADE,
			name VARCHAR(128) NOT NULL,
			description TEXT,
			created_at %s,
			commit_ids INTEGER[] DEFAULT ARRAY[]::INTEGER[],
			head_commit_id INTEGER REFERENCES commit_history(id) ON DELETE SET NULL,
			is_active %s DEFAULT true,
			UNIQUE(repository_id, name)
		)
	`, autoIncrementSyntax, timestampSyntax, booleanType))
	if err != nil {
		return fmt.Errorf("error creating branches table: %v", err)
	}

	// Create indexes
	_, err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_commit_details_commit_id ON commit_details(commit_id);
		CREATE INDEX IF NOT EXISTS idx_branches_name ON branches(name);
		CREATE INDEX IF NOT EXISTS idx_repositories_owner ON repositories(owner_id);
		CREATE INDEX IF NOT EXISTS idx_repository_access_user ON repository_access(user_id);
		CREATE INDEX IF NOT EXISTS idx_repository_access_repo ON repository_access(repository_id);
		CREATE INDEX IF NOT EXISTS idx_commit_history_repo ON commit_history(repository_id);
		CREATE INDEX IF NOT EXISTS idx_branches_repo ON branches(repository_id);
	`)
	if err != nil {
		return fmt.Errorf("error creating indexes: %v", err)
	}

	return nil
}
