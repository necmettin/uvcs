package commands

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"uvcs/modules/db"
)

// CreateRepository creates a new repository for a user
func CreateRepository(name, description string, ownerUsername string) error {
	var ownerID int
	err := db.DB.QueryRow(`
		SELECT id FROM users 
		WHERE username = $1 AND is_active = true
	`, ownerUsername).Scan(&ownerID)
	if err != nil {
		return fmt.Errorf("error finding user: %v", err)
	}

	// Create repository
	_, err = db.DB.Exec(`
		INSERT INTO repositories (name, description, owner_id)
		VALUES ($1, $2, $3)
	`, name, description, ownerID)
	if err != nil {
		return fmt.Errorf("error creating repository: %v", err)
	}

	fmt.Printf("Successfully created repository '%s' owned by '%s'\n", name, ownerUsername)
	return nil
}

// ListRepositories lists all repositories a user has access to
func ListRepositories(username string) error {
	rows, err := db.DB.Query(`
		WITH user_repos AS (
			-- Repositories owned by the user
			SELECT r.*, 'owner' as access_type
			FROM repositories r
			JOIN users u ON r.owner_id = u.id
			WHERE u.username = $1
			UNION
			-- Repositories the user has access to
			SELECT r.*, ra.access_level as access_type
			FROM repositories r
			JOIN repository_access ra ON r.id = ra.repository_id
			JOIN users u ON ra.user_id = u.id
			WHERE u.username = $1
		)
		SELECT 
			name,
			description,
			created_at,
			access_type,
			(SELECT username FROM users WHERE id = owner_id) as owner
		FROM user_repos
		WHERE is_active = true
		ORDER BY created_at DESC
	`, username)
	if err != nil {
		return fmt.Errorf("error querying repositories: %v", err)
	}
	defer rows.Close()

	fmt.Printf("\nRepositories:\n")
	fmt.Printf("%-30s %-30s %-20s %-10s %s\n", "Name", "Description", "Created At", "Access", "Owner")
	fmt.Println(strings.Repeat("-", 120))

	for rows.Next() {
		var (
			name        string
			description string
			createdAt   time.Time
			accessType  string
			owner       string
		)
		err := rows.Scan(&name, &description, &createdAt, &accessType, &owner)
		if err != nil {
			return fmt.Errorf("error scanning repository: %v", err)
		}

		fmt.Printf("%-30s %-30s %-20s %-10s %s\n",
			name,
			truncateString(description, 27),
			createdAt.Format("2006-01-02 15:04:05"),
			accessType,
			owner,
		)
	}
	fmt.Println()
	return nil
}

// GrantAccess grants repository access to a user
func GrantAccess(repoOwner, repoName, username, accessLevel string) error {
	var repoID int
	err := db.DB.QueryRow(`
		SELECT r.id 
		FROM repositories r
		JOIN users u ON r.owner_id = u.id
		WHERE u.username = $1 AND r.name = $2 AND r.is_active = true
	`, repoOwner, repoName).Scan(&repoID)
	if err != nil {
		return fmt.Errorf("error finding repository: %v", err)
	}

	var userID int
	err = db.DB.QueryRow(`
		SELECT id FROM users 
		WHERE username = $1 AND is_active = true
	`, username).Scan(&userID)
	if err != nil {
		return fmt.Errorf("error finding user: %v", err)
	}

	// Grant access
	_, err = db.DB.Exec(`
		INSERT INTO repository_access (repository_id, user_id, access_level)
		VALUES ($1, $2, $3)
		ON CONFLICT (repository_id, user_id) 
		DO UPDATE SET access_level = $3
	`, repoID, userID, accessLevel)
	if err != nil {
		return fmt.Errorf("error granting access: %v", err)
	}

	fmt.Printf("Successfully granted %s access to '%s/%s' for user '%s'\n",
		accessLevel, repoOwner, repoName, username)
	return nil
}

// RevokeAccess revokes repository access from a user
func RevokeAccess(repoOwner, repoName, username string) error {
	result, err := db.DB.Exec(`
		DELETE FROM repository_access ra
		USING repositories r, users u, users target
		WHERE r.id = ra.repository_id
		AND u.id = r.owner_id
		AND target.id = ra.user_id
		AND u.username = $1
		AND r.name = $2
		AND target.username = $3
	`, repoOwner, repoName, username)
	if err != nil {
		return fmt.Errorf("error revoking access: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("no access found for user '%s' on repository '%s/%s'",
			username, repoOwner, repoName)
	}

	fmt.Printf("Successfully revoked access to '%s/%s' from user '%s'\n",
		repoOwner, repoName, username)
	return nil
}

// ListAccess lists all users who have access to a repository
func ListAccess(repoOwner, repoName string) error {
	rows, err := db.DB.Query(`
		WITH repo_users AS (
			-- Owner
			SELECT 
				u.username,
				u.firstname || ' ' || u.lastname as full_name,
				'owner' as access_level,
				r.created_at as granted_at,
				NULL as granted_by
			FROM repositories r
			JOIN users u ON r.owner_id = u.id
			JOIN users owner ON owner.id = r.owner_id
			WHERE owner.username = $1 AND r.name = $2
			UNION
			-- Users with explicit access
			SELECT 
				u.username,
				u.firstname || ' ' || u.lastname as full_name,
				ra.access_level,
				ra.granted_at,
				granter.username as granted_by
			FROM repository_access ra
			JOIN repositories r ON ra.repository_id = r.id
			JOIN users u ON ra.user_id = u.id
			JOIN users owner ON r.owner_id = owner.id
			LEFT JOIN users granter ON ra.granted_by = granter.id
			WHERE owner.username = $1 AND r.name = $2
		)
		SELECT * FROM repo_users
		ORDER BY 
			CASE WHEN access_level = 'owner' THEN 1
				 WHEN access_level = 'write' THEN 2
				 ELSE 3 
			END,
			username
	`, repoOwner, repoName)
	if err != nil {
		return fmt.Errorf("error querying access: %v", err)
	}
	defer rows.Close()

	fmt.Printf("\nAccess for repository '%s/%s':\n", repoOwner, repoName)
	fmt.Printf("%-20s %-30s %-10s %-20s %s\n",
		"Username", "Full Name", "Access", "Granted At", "Granted By")
	fmt.Println(strings.Repeat("-", 100))

	for rows.Next() {
		var (
			username  string
			fullName  string
			access    string
			grantedAt time.Time
			grantedBy sql.NullString
		)
		err := rows.Scan(&username, &fullName, &access, &grantedAt, &grantedBy)
		if err != nil {
			return fmt.Errorf("error scanning access: %v", err)
		}

		grantedByStr := "-"
		if grantedBy.Valid {
			grantedByStr = grantedBy.String
		}

		fmt.Printf("%-20s %-30s %-10s %-20s %s\n",
			username,
			fullName,
			access,
			grantedAt.Format("2006-01-02 15:04:05"),
			grantedByStr,
		)
	}
	fmt.Println()
	return nil
}
