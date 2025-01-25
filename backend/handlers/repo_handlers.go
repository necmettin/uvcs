package handlers

import (
	"encoding/json"
	"net/http"
	"uvcs/modules/db"

	"github.com/gin-gonic/gin"
)

type RepositoryResponse struct {
	Branches []BranchInfo `json:"branches"`
	Commits  []CommitInfo `json:"commits"`
	Content  RepoContent  `json:"content"`
}

type BranchInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	CommitIDs   []int  `json:"commit_ids"`
	HeadCommit  int    `json:"head_commit"`
	IsActive    bool   `json:"is_active"`
}

type CommitInfo struct {
	ID       int          `json:"id"`
	Hash     string       `json:"hash"`
	Message  string       `json:"message"`
	DateTime string       `json:"datetime"`
	Tags     []string     `json:"tags"`
	Author   Author       `json:"author"`
	Changes  []FileChange `json:"changes"`
}

type Author struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
}

type FileChange struct {
	FilePath      string          `json:"file_path"`
	ChangeType    string          `json:"change_type"` // "M", "D", "A"
	ContentChange json.RawMessage `json:"content_change"`
}

type RepoContent struct {
	Files map[string]FileContent `json:"files"`
}

type FileContent struct {
	Content   string `json:"content"`
	CommitID  int    `json:"commit_id"`
	Timestamp string `json:"timestamp"`
}

// HandleGetRepository returns complete repository information
func HandleGetRepository(c *gin.Context) {
	if !validateKeys(c) {
		return
	}

	// Get all commits with author info and tags
	commits, err := db.DB.Query(`
		SELECT 
			ch.id,
			ch.commit_hash,
			ch.commit_message,
			ch.commit_datetime,
			ch.tags,
			u.id as user_id,
			u.firstname || ' ' || u.lastname as full_name
		FROM commit_history ch
		JOIN users u ON ch.user_id = u.id
		ORDER BY ch.commit_datetime DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching commits"})
		return
	}
	defer commits.Close()

	var commitInfos []CommitInfo
	for commits.Next() {
		var (
			commit   CommitInfo
			author   Author
			datetime string
			tags     []string
		)
		err := commits.Scan(
			&commit.ID,
			&commit.Hash,
			&commit.Message,
			&datetime,
			&tags,
			&author.ID,
			&author.FullName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning commits"})
			return
		}
		commit.DateTime = datetime
		commit.Tags = tags
		commit.Author = author

		// Get file changes for this commit
		changes, err := db.DB.Query(`
			SELECT file_path, change_type, content_changes
			FROM commit_details
			WHERE commit_id = $1
		`, commit.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching commit details"})
			return
		}
		defer changes.Close()

		for changes.Next() {
			var change FileChange
			var contentChanges []byte
			err := changes.Scan(&change.FilePath, &change.ChangeType, &contentChanges)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning commit details"})
				return
			}
			change.ContentChange = json.RawMessage(contentChanges)
			commit.Changes = append(commit.Changes, change)
		}

		commitInfos = append(commitInfos, commit)
	}

	// Get all branches
	branches, err := db.DB.Query(`
		SELECT 
			id,
			name,
			description,
			created_at,
			commit_ids,
			head_commit_id,
			is_active
		FROM branches
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching branches"})
		return
	}
	defer branches.Close()

	var branchInfos []BranchInfo
	for branches.Next() {
		var (
			branch     BranchInfo
			createdAt  string
			commitIDs  []int
			headCommit int
		)
		err := branches.Scan(
			&branch.ID,
			&branch.Name,
			&branch.Description,
			&createdAt,
			&commitIDs,
			&headCommit,
			&branch.IsActive,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning branches"})
			return
		}
		branch.CreatedAt = createdAt
		branch.CommitIDs = commitIDs
		branch.HeadCommit = headCommit
		branchInfos = append(branchInfos, branch)
	}

	// Build current repository content
	content := RepoContent{
		Files: make(map[string]FileContent),
	}

	// Get latest content for each file
	files, err := db.DB.Query(`
		WITH RankedChanges AS (
			SELECT 
				cd.file_path,
				cd.change_type,
				cd.content_changes,
				cd.commit_id,
				ch.commit_datetime,
				ROW_NUMBER() OVER (PARTITION BY cd.file_path ORDER BY ch.commit_datetime DESC) as rn
			FROM commit_details cd
			JOIN commit_history ch ON cd.commit_id = ch.id
			WHERE cd.change_type != 'D'
		)
		SELECT 
			file_path,
			content_changes,
			commit_id,
			commit_datetime
		FROM RankedChanges
		WHERE rn = 1
		ORDER BY file_path
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching repository content"})
		return
	}
	defer files.Close()

	for files.Next() {
		var (
			filePath    string
			contentJSON []byte
			commitID    int
			timestamp   string
		)
		err := files.Scan(&filePath, &contentJSON, &commitID, &timestamp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning repository content"})
			return
		}

		var contentMap map[string]interface{}
		err = json.Unmarshal(contentJSON, &contentMap)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing file content"})
			return
		}

		content.Files[filePath] = FileContent{
			Content:   contentMap["content"].(string),
			CommitID:  commitID,
			Timestamp: timestamp,
		}
	}

	response := RepositoryResponse{
		Branches: branchInfos,
		Commits:  commitInfos,
		Content:  content,
	}

	c.JSON(http.StatusOK, response)
}
