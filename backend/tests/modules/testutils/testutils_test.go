package testutils_test

import (
	"testing"
	"uvcs/modules/testutils"

	"github.com/stretchr/testify/assert"
)

func TestSetupTestDB(t *testing.T) {
	// Test database setup
	db := testutils.SetupTestDB(t)
	assert.NotNil(t, db)
	defer db.Close()

	// Verify tables exist
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'users'
		)
	`).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCreateTestUser(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	// Test user creation
	user := testutils.CreateTestUser(t, db)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.ID)
	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Email)
	assert.NotEmpty(t, user.Skey1)
	assert.NotEmpty(t, user.Skey2)

	// Verify user exists in database
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM users 
			WHERE id = $1
		)
	`, user.ID).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCreateTestRepository(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	user := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, user.ID)

	assert.NotNil(t, repo)
	assert.NotEmpty(t, repo.ID)
	assert.NotEmpty(t, repo.Name)
	assert.Equal(t, user.ID, repo.OwnerID)

	// Verify repository exists in database
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM repositories 
			WHERE id = $1
		)
	`, repo.ID).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestGrantRepositoryAccess(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	owner := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, owner.ID)
	user := testutils.CreateTestUser(t, db)

	// Test granting access
	testutils.GrantRepositoryAccess(t, db, repo.ID, user.ID, "read")

	// Verify access exists in database
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM repository_access 
			WHERE repository_id = $1 
			AND user_id = $2
			AND access_level = 'read'
		)
	`, repo.ID, user.ID).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCreateTestCommit(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	user := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, user.ID)
	commit := testutils.CreateTestCommit(t, db, repo.ID, user.ID)

	assert.NotNil(t, commit)
	assert.NotEmpty(t, commit.ID)
	assert.NotEmpty(t, commit.Hash)
	assert.Equal(t, repo.ID, commit.RepositoryID)
	assert.Equal(t, user.ID, commit.UserID)

	// Verify commit exists in database
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM commit_history 
			WHERE id = $1
		)
	`, commit.ID).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCreateTestCommitDetail(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	user := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, user.ID)
	commit := testutils.CreateTestCommit(t, db, repo.ID, user.ID)
	detail := testutils.CreateTestCommitDetail(t, db, commit.ID)

	assert.NotNil(t, detail)
	assert.NotEmpty(t, detail.ID)
	assert.Equal(t, commit.ID, detail.CommitID)
	assert.NotEmpty(t, detail.FilePath)
	assert.NotEmpty(t, detail.ChangeType)
	assert.NotEmpty(t, detail.ContentChange)

	// Verify commit detail exists in database
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM commit_details 
			WHERE id = $1
		)
	`, detail.ID).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCreateTestBranch(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	user := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, user.ID)
	branch := testutils.CreateTestBranch(t, db, repo.ID, "feature/test")

	assert.NotNil(t, branch)
	assert.NotEmpty(t, branch.ID)
	assert.Equal(t, repo.ID, branch.RepositoryID)
	assert.Equal(t, "feature/test", branch.Name)

	// Verify branch exists in database
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM branches 
			WHERE id = $1
		)
	`, branch.ID).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)
}
