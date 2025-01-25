package commands

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"uvcs/modules/db"
)

type Branch struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
	CommitIDs   []int
	HeadCommit  int
	IsActive    bool
}

// ListBranches returns all active branches
func ListBranches() error {
	rows, err := db.DB.Query(`
		SELECT id, name, description, created_at, commit_ids, head_commit_id, is_active 
		FROM branches 
		WHERE is_active = true
		ORDER BY created_at DESC
	`)
	if err != nil {
		return fmt.Errorf("error querying branches: %v", err)
	}
	defer rows.Close()

	fmt.Printf("\nBranches:\n")
	fmt.Printf("%-5s %-20s %-30s %-20s %s\n", "ID", "Name", "Description", "Created At", "Commits")
	fmt.Println(strings.Repeat("-", 100))

	for rows.Next() {
		var b Branch
		var desc sql.NullString
		err := rows.Scan(&b.ID, &b.Name, &desc, &b.CreatedAt, &b.CommitIDs, &b.HeadCommit, &b.IsActive)
		if err != nil {
			return fmt.Errorf("error scanning branch: %v", err)
		}
		b.Description = desc.String
		fmt.Printf("%-5d %-20s %-30s %-20s %d commits\n",
			b.ID,
			b.Name,
			truncateString(b.Description, 27),
			b.CreatedAt.Format("2006-01-02 15:04:05"),
			len(b.CommitIDs),
		)
	}
	fmt.Println()
	return nil
}

// CreateBranch creates a new branch
func CreateBranch(name string) error {
	// Check if branch already exists
	var exists bool
	err := db.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM branches WHERE name = $1)
	`, name).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking branch existence: %v", err)
	}
	if exists {
		return fmt.Errorf("branch '%s' already exists", name)
	}

	// Create new branch
	_, err = db.DB.Exec(`
		INSERT INTO branches (name, description, commit_ids)
		VALUES ($1, $2, $3)
	`, name, fmt.Sprintf("Branch created via CLI on %s", time.Now().Format("2006-01-02")), []int{})

	if err != nil {
		return fmt.Errorf("error creating branch: %v", err)
	}

	fmt.Printf("Successfully created branch '%s'\n", name)
	return nil
}

// DeleteBranch soft deletes a branch by setting is_active to false
func DeleteBranch(name string) error {
	result, err := db.DB.Exec(`
		UPDATE branches 
		SET is_active = false 
		WHERE name = $1 AND is_active = true
	`, name)
	if err != nil {
		return fmt.Errorf("error deleting branch: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("branch '%s' not found or already deleted", name)
	}

	fmt.Printf("Successfully deleted branch '%s'\n", name)
	return nil
}

// ListCommits lists all commits in a branch
func ListCommits(branchName string) error {
	// First check if branch exists and is active
	var branchID int
	var commitIDs []int
	err := db.DB.QueryRow(`
		SELECT id, commit_ids 
		FROM branches 
		WHERE name = $1 AND is_active = true
	`, branchName).Scan(&branchID, &commitIDs)
	if err == sql.ErrNoRows {
		return fmt.Errorf("branch '%s' not found or is inactive", branchName)
	}
	if err != nil {
		return fmt.Errorf("error querying branch: %v", err)
	}

	if len(commitIDs) == 0 {
		fmt.Printf("\nNo commits found in branch '%s'\n\n", branchName)
		return nil
	}

	// Query commits
	rows, err := db.DB.Query(`
		SELECT ch.id, ch.commit_hash, ch.commit_message, ch.commit_datetime, u.firstname, u.lastname
		FROM commit_history ch
		JOIN users u ON ch.user_id = u.id
		WHERE ch.id = ANY($1)
		ORDER BY ch.commit_datetime DESC
	`, commitIDs)
	if err != nil {
		return fmt.Errorf("error querying commits: %v", err)
	}
	defer rows.Close()

	fmt.Printf("\nCommits in branch '%s':\n", branchName)
	fmt.Printf("%-5s %-12s %-20s %-50s %s\n", "ID", "Date", "Author", "Message", "Hash")
	fmt.Println(strings.Repeat("-", 120))

	for rows.Next() {
		var (
			id        int
			hash      string
			message   string
			datetime  time.Time
			firstname string
			lastname  string
		)
		err := rows.Scan(&id, &hash, &message, &datetime, &firstname, &lastname)
		if err != nil {
			return fmt.Errorf("error scanning commit: %v", err)
		}

		fmt.Printf("%-5d %-12s %-20s %-50s %s\n",
			id,
			datetime.Format("2006-01-02"),
			fmt.Sprintf("%s %s", firstname, lastname),
			truncateString(message, 47),
			truncateString(hash, 12),
		)
	}
	fmt.Println()
	return nil
}

// Helper function to truncate strings
func truncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length-3] + "..."
}

// ListBranchesAPI returns all active branches as structured data
func ListBranchesAPI() ([]Branch, error) {
	rows, err := db.DB.Query(`
		SELECT id, name, description, created_at, commit_ids, head_commit_id, is_active 
		FROM branches 
		WHERE is_active = true
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying branches: %v", err)
	}
	defer rows.Close()

	var branches []Branch
	for rows.Next() {
		var b Branch
		var desc sql.NullString
		err := rows.Scan(&b.ID, &b.Name, &desc, &b.CreatedAt, &b.CommitIDs, &b.HeadCommit, &b.IsActive)
		if err != nil {
			return nil, fmt.Errorf("error scanning branch: %v", err)
		}
		b.Description = desc.String
		branches = append(branches, b)
	}

	return branches, nil
}

type Commit struct {
	ID       int       `json:"id"`
	Hash     string    `json:"hash"`
	Message  string    `json:"message"`
	DateTime time.Time `json:"datetime"`
	Author   string    `json:"author"`
}

// ListCommitsAPI returns all commits in a branch as structured data
func ListCommitsAPI(branchName string) ([]Commit, error) {
	// First check if branch exists and is active
	var branchID int
	var commitIDs []int
	err := db.DB.QueryRow(`
		SELECT id, commit_ids 
		FROM branches 
		WHERE name = $1 AND is_active = true
	`, branchName).Scan(&branchID, &commitIDs)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("branch '%s' not found or is inactive", branchName)
	}
	if err != nil {
		return nil, fmt.Errorf("error querying branch: %v", err)
	}

	if len(commitIDs) == 0 {
		return []Commit{}, nil
	}

	// Query commits
	rows, err := db.DB.Query(`
		SELECT ch.id, ch.commit_hash, ch.commit_message, ch.commit_datetime, u.firstname, u.lastname
		FROM commit_history ch
		JOIN users u ON ch.user_id = u.id
		WHERE ch.id = ANY($1)
		ORDER BY ch.commit_datetime DESC
	`, commitIDs)
	if err != nil {
		return nil, fmt.Errorf("error querying commits: %v", err)
	}
	defer rows.Close()

	var commits []Commit
	for rows.Next() {
		var (
			id        int
			hash      string
			message   string
			datetime  time.Time
			firstname string
			lastname  string
		)
		err := rows.Scan(&id, &hash, &message, &datetime, &firstname, &lastname)
		if err != nil {
			return nil, fmt.Errorf("error scanning commit: %v", err)
		}

		commits = append(commits, Commit{
			ID:       id,
			Hash:     hash,
			Message:  message,
			DateTime: datetime,
			Author:   fmt.Sprintf("%s %s", firstname, lastname),
		})
	}

	return commits, nil
}
