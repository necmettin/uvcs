package db_test

import (
	"database/sql"
	"testing"
	"uvcs/modules/db"

	"github.com/stretchr/testify/assert"
)

func TestInitDB(t *testing.T) {
	// Test with valid connection string
	db, err := db.InitDB("postgres://postgres:postgres@localhost:5432/uvcs_test?sslmode=disable")
	assert.NoError(t, err)
	assert.NotNil(t, db)
	if db != nil {
		db.Close()
	}

	// Test with invalid connection string
	db, err = db.InitDB("invalid://connection/string")
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestCreateTables(t *testing.T) {
	// Setup test database
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/uvcs_test?sslmode=disable")
	assert.NoError(t, err)
	defer db.Close()

	// Drop all tables if they exist
	_, err = db.Exec(`
		DROP TABLE IF EXISTS repository_access CASCADE;
		DROP TABLE IF EXISTS commit_details CASCADE;
		DROP TABLE IF EXISTS commit_history CASCADE;
		DROP TABLE IF EXISTS branches CASCADE;
		DROP TABLE IF EXISTS repositories CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
	`)
	assert.NoError(t, err)

	// Test creating tables
	err = db.CreateTables(db)
	assert.NoError(t, err)

	// Verify tables were created
	tables := []string{
		"users",
		"repositories",
		"repository_access",
		"commit_history",
		"commit_details",
		"branches",
	}

	for _, table := range tables {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`, table).Scan(&exists)
		assert.NoError(t, err)
		assert.True(t, exists, "Table %s should exist", table)
	}

	// Verify foreign key constraints
	constraints := []struct {
		table     string
		column    string
		refTable  string
		refColumn string
	}{
		{"repositories", "owner_id", "users", "id"},
		{"repository_access", "repository_id", "repositories", "id"},
		{"repository_access", "user_id", "users", "id"},
		{"repository_access", "granted_by", "users", "id"},
		{"commit_history", "repository_id", "repositories", "id"},
		{"commit_history", "user_id", "users", "id"},
		{"commit_details", "commit_id", "commit_history", "id"},
		{"branches", "repository_id", "repositories", "id"},
		{"branches", "head_commit_id", "commit_history", "id"},
	}

	for _, c := range constraints {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
					ON tc.constraint_name = kcu.constraint_name
				JOIN information_schema.constraint_column_usage ccu
					ON ccu.constraint_name = tc.constraint_name
				WHERE tc.constraint_type = 'FOREIGN KEY'
				AND kcu.table_name = $1
				AND kcu.column_name = $2
				AND ccu.table_name = $3
				AND ccu.column_name = $4
			)
		`, c.table, c.column, c.refTable, c.refColumn).Scan(&exists)
		assert.NoError(t, err)
		assert.True(t, exists, "Foreign key constraint should exist from %s.%s to %s.%s",
			c.table, c.column, c.refTable, c.refColumn)
	}
}

func TestGetDB(t *testing.T) {
	// Test when DB is not initialized
	db.DB = nil
	db := db.GetDB()
	assert.Nil(t, db)

	// Test when DB is initialized
	testDB, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/uvcs_test?sslmode=disable")
	assert.NoError(t, err)
	defer testDB.Close()

	db.DB = testDB
	db = db.GetDB()
	assert.NotNil(t, db)
	assert.Equal(t, db.DB, db)
}
