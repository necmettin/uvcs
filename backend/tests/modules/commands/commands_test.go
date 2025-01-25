package commands_test

import (
	"testing"
	"uvcs/modules/commands"
	"uvcs/modules/testutils"

	"github.com/stretchr/testify/assert"
)

func TestListBranches(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	user := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, user.ID)

	// Test listing branches for existing repository
	err := commands.ListBranches(repo.Name, user.Username)
	assert.NoError(t, err)

	// Test listing branches for non-existent repository
	err = commands.ListBranches("non-existent", user.Username)
	assert.Error(t, err)

	// Test listing branches without permission
	otherUser := testutils.CreateTestUser(t, db)
	err = commands.ListBranches(repo.Name, otherUser.Username)
	assert.Error(t, err)
}

func TestCreateBranch(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	user := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, user.ID)

	tests := []struct {
		name        string
		branchName  string
		username    string
		shouldError bool
	}{
		{
			name:        "valid branch creation",
			branchName:  "feature/new-branch",
			username:    user.Username,
			shouldError: false,
		},
		{
			name:        "duplicate branch",
			branchName:  "feature/new-branch",
			username:    user.Username,
			shouldError: true,
		},
		{
			name:        "empty branch name",
			branchName:  "",
			username:    user.Username,
			shouldError: true,
		},
		{
			name:        "unauthorized user",
			branchName:  "feature/another-branch",
			username:    "unauthorized",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := commands.CreateBranch(repo.Name, test.branchName, test.username)
			if test.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteBranch(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	user := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, user.ID)
	branch := testutils.CreateTestBranch(t, db, repo.ID, "feature/test")

	tests := []struct {
		name        string
		branchName  string
		username    string
		shouldError bool
	}{
		{
			name:        "valid branch deletion",
			branchName:  branch.Name,
			username:    user.Username,
			shouldError: false,
		},
		{
			name:        "non-existent branch",
			branchName:  "non-existent",
			username:    user.Username,
			shouldError: true,
		},
		{
			name:        "unauthorized user",
			branchName:  branch.Name,
			username:    "unauthorized",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := commands.DeleteBranch(repo.Name, test.branchName, test.username)
			if test.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListCommits(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	user := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, user.ID)
	commit := testutils.CreateTestCommit(t, db, repo.ID, user.ID)

	// Test listing commits for existing repository
	err := commands.ListCommits(repo.Name, user.Username)
	assert.NoError(t, err)

	// Test listing commits for non-existent repository
	err = commands.ListCommits("non-existent", user.Username)
	assert.Error(t, err)

	// Test listing commits without permission
	otherUser := testutils.CreateTestUser(t, db)
	err = commands.ListCommits(repo.Name, otherUser.Username)
	assert.Error(t, err)
}

func TestGrantAccess(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	owner := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, owner.ID)
	user := testutils.CreateTestUser(t, db)

	tests := []struct {
		name        string
		username    string
		accessLevel string
		shouldError bool
	}{
		{
			name:        "grant read access",
			username:    user.Username,
			accessLevel: "read",
			shouldError: false,
		},
		{
			name:        "grant write access",
			username:    user.Username,
			accessLevel: "write",
			shouldError: false,
		},
		{
			name:        "invalid access level",
			username:    user.Username,
			accessLevel: "invalid",
			shouldError: true,
		},
		{
			name:        "non-existent user",
			username:    "non-existent",
			accessLevel: "read",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := commands.GrantAccess(repo.Name, owner.Username, test.username, test.accessLevel)
			if test.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRevokeAccess(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	owner := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, owner.ID)
	user := testutils.CreateTestUser(t, db)
	testutils.GrantRepositoryAccess(t, db, repo.ID, user.ID, "read")

	tests := []struct {
		name        string
		username    string
		shouldError bool
	}{
		{
			name:        "revoke existing access",
			username:    user.Username,
			shouldError: false,
		},
		{
			name:        "revoke non-existent access",
			username:    "non-existent",
			shouldError: true,
		},
		{
			name:        "revoke owner access",
			username:    owner.Username,
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := commands.RevokeAccess(repo.Name, owner.Username, test.username)
			if test.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListAccess(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	owner := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, owner.ID)
	user1 := testutils.CreateTestUser(t, db)
	user2 := testutils.CreateTestUser(t, db)

	testutils.GrantRepositoryAccess(t, db, repo.ID, user1.ID, "read")
	testutils.GrantRepositoryAccess(t, db, repo.ID, user2.ID, "write")

	// Test listing access for existing repository
	err := commands.ListAccess(repo.Name, owner.Username)
	assert.NoError(t, err)

	// Test listing access for non-existent repository
	err = commands.ListAccess("non-existent", owner.Username)
	assert.Error(t, err)

	// Test listing access without permission
	err = commands.ListAccess(repo.Name, "unauthorized")
	assert.Error(t, err)
}
