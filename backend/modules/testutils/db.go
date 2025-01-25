package testutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"uvcs/modules/db"
	"uvcs/modules/utils"
)

// TestDB represents a test database connection
type TestDB struct {
	DB *sql.DB
}

// SetupTestDB creates a new test database and returns a cleanup function
func SetupTestDB(t *testing.T) (*TestDB, func()) {
	// Set test environment
	os.Setenv("DB_TYPE", "sqlite")
	os.Setenv("SQLITE_PATH", ":memory:")

	// Initialize database
	if err := db.InitDB(true); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	return &TestDB{DB: db.DB}, func() {
		db.DB.Close()
	}
}

// CreateTestUser creates a test user and returns their credentials
func (tdb *TestDB) CreateTestUser(t *testing.T) (userID int, username, email, skey1, skey2 string) {
	username = fmt.Sprintf("test_user_%d", utils.GenerateRandomInt())
	email = fmt.Sprintf("test_%d@example.com", utils.GenerateRandomInt())
	password := "test_password"
	firstname := "Test"
	lastname := "User"
	skey1 = utils.GenerateSecurityKey()
	skey2 = utils.GenerateSecurityKey()

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	err = tdb.DB.QueryRow(`
		INSERT INTO users (username, email, password, firstname, lastname, skey1, skey2, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		RETURNING id
	`, username, email, hashedPassword, firstname, lastname, skey1, skey2).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return
}

// CreateTestRepository creates a test repository and returns its ID
func (tdb *TestDB) CreateTestRepository(t *testing.T, ownerID int) (repoID int, repoName string) {
	repoName = fmt.Sprintf("test_repo_%d", utils.GenerateRandomInt())
	description := "Test repository"

	err := tdb.DB.QueryRow(`
		INSERT INTO repositories (name, description, owner_id, is_active)
		VALUES ($1, $2, $3, true)
		RETURNING id
	`, repoName, description, ownerID).Scan(&repoID)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}

	return
}

// GrantRepositoryAccess grants access to a repository for a user
func (tdb *TestDB) GrantRepositoryAccess(t *testing.T, repoID, userID int, accessLevel string) {
	_, err := tdb.DB.Exec(`
		INSERT INTO repository_access (repository_id, user_id, access_level)
		VALUES ($1, $2, $3)
	`, repoID, userID, accessLevel)
	if err != nil {
		t.Fatalf("Failed to grant repository access: %v", err)
	}
}

// CreateTestCommit creates a test commit and returns its ID
func (tdb *TestDB) CreateTestCommit(t *testing.T, repoID, userID int, message string, tags []string) (commitID int) {
	err := tdb.DB.QueryRow(`
		INSERT INTO commit_history (repository_id, user_id, commit_hash, commit_message, tags)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, repoID, userID, utils.GenerateCommitHash(), message, tags).Scan(&commitID)
	if err != nil {
		t.Fatalf("Failed to create test commit: %v", err)
	}

	return
}

// CreateTestCommitDetail creates a test commit detail
func (tdb *TestDB) CreateTestCommitDetail(t *testing.T, commitID int, filePath, content string, isCode, isBinary, isDiff bool) {
	contentChanges := map[string]interface{}{
		"content":   content,
		"is_code":   isCode,
		"is_binary": isBinary,
		"is_diff":   isDiff,
	}
	contentJSON, err := json.Marshal(contentChanges)
	if err != nil {
		t.Fatalf("Failed to marshal content changes: %v", err)
	}

	_, err = tdb.DB.Exec(`
		INSERT INTO commit_details (commit_id, file_path, change_type, content_changes)
		VALUES ($1, $2, $3, $4)
	`, commitID, filePath, "M", contentJSON)
	if err != nil {
		t.Fatalf("Failed to create test commit detail: %v", err)
	}
}
