package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"uvcs/modules/db"
	"uvcs/modules/utils"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// FileCommit represents a file to be committed
type FileCommit struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	IsCode     bool   `json:"is_code"`
	IsBinary   bool   `json:"is_binary"`
	ChangeType string `json:"change_type"` // "M", "D", "A"
	IsDiff     bool   `json:"is_diff"`     // whether content is full file or diff
}

// FileHistory represents a file's previous state
type FileHistory struct {
	Content    string
	IsCode     bool
	IsBinary   bool
	CommitID   int
	CommitHash string
}

// CommitRequest represents the commit request structure
type CommitRequest struct {
	Files   []FileCommit `json:"files"`
	Message string       `json:"message"`
	Tags    []string     `json:"tags"`
}

// isCodeFile checks if a file is a code file based on its extension
func isCodeFile(filename string) bool {
	codeExtensions := map[string]bool{
		".go":    true,
		".js":    true,
		".ts":    true,
		".jsx":   true,
		".tsx":   true,
		".py":    true,
		".java":  true,
		".cpp":   true,
		".c":     true,
		".h":     true,
		".hpp":   true,
		".rs":    true,
		".rb":    true,
		".php":   true,
		".cs":    true,
		".swift": true,
		".kt":    true,
		".scala": true,
		".m":     true,
		".mm":    true,
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return codeExtensions[ext]
}

// isBinaryFile checks if a file is binary by looking at its content
func isBinaryFile(content []byte) bool {
	// Check for null bytes and other binary indicators
	for _, b := range content {
		if b == 0 {
			return true
		}
	}
	return false
}

// processCodeFile removes unnecessary whitespace and empty lines
func processCodeFile(content string) string {
	// Split into lines
	lines := strings.Split(content, "\n")

	// Process each line
	var processedLines []string
	for _, line := range lines {
		// Trim whitespace
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Add to processed lines
		processedLines = append(processedLines, trimmed)
	}

	// Join lines back together
	return strings.Join(processedLines, "\n")
}

// getFileHistory gets the most recent version of a file from previous commits
func getFileHistory(tx *sql.Tx, repoID int, filePath string) (*FileHistory, error) {
	var history FileHistory
	err := tx.QueryRow(`
		WITH latest_commit AS (
			SELECT cd.commit_id, cd.content_changes, ch.commit_hash
			FROM commit_details cd
			JOIN commit_history ch ON cd.commit_id = ch.id
			WHERE ch.repository_id = $1 AND cd.file_path = $2
			ORDER BY ch.commit_datetime DESC
			LIMIT 1
		)
		SELECT 
			lc.commit_id,
			lc.content_changes,
			lc.commit_hash
		FROM latest_commit lc
	`, repoID, filePath).Scan(&history.CommitID, &history.Content, &history.CommitHash)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Parse content changes
	var contentChanges struct {
		Content  string `json:"content"`
		IsCode   bool   `json:"is_code"`
		IsBinary bool   `json:"is_binary"`
		IsDiff   bool   `json:"is_diff"`
	}
	err = json.Unmarshal([]byte(history.Content), &contentChanges)
	if err != nil {
		return nil, err
	}

	history.Content = contentChanges.Content
	history.IsCode = contentChanges.IsCode
	history.IsBinary = contentChanges.IsBinary

	// If the content is a diff, we need to reconstruct the full file
	if contentChanges.IsDiff {
		// Recursively get previous version and apply diffs
		prevHistory, err := getFileHistory(tx, repoID, filePath)
		if err != nil {
			return nil, err
		}

		dmp := diffmatchpatch.New()
		patches, err := dmp.PatchFromText(history.Content)
		if err != nil {
			return nil, err
		}

		result, _ := dmp.PatchApply(patches, prevHistory.Content)
		history.Content = result
	}

	return &history, nil
}

// HandleCommit handles file commits
func HandleCommit(c *gin.Context) {
	// Verify authentication
	userID, err := authenticateRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Get repository information
	repoName := c.PostForm("name")
	message := c.PostForm("message")
	tagsStr := c.PostForm("tags") // comma-separated list of tags

	if repoName == "" || message == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Repository name and commit message are required",
		})
		return
	}

	// Parse tags
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// Get repository ID and verify access
	var repoID int
	err = db.DB.QueryRow(`
		SELECT r.id 
		FROM repositories r
		LEFT JOIN repository_access ra ON r.id = ra.repository_id AND ra.user_id = $1
		WHERE r.name = $2 AND r.is_active = true
		AND (r.owner_id = $1 OR ra.access_level = 'write')
	`, userID, repoName).Scan(&repoID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "No write access to repository",
		})
		return
	}

	// Process files
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	files := form.File["files[]"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided"})
		return
	}

	// Begin transaction
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	var processedFiles []FileCommit
	for _, file := range files {
		// Open the file
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error opening file %s: %v", file.Filename, err),
			})
			return
		}
		defer f.Close()

		// Read the content
		content, err := io.ReadAll(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error reading file %s: %v", file.Filename, err),
			})
			return
		}

		// Get file history
		history, err := getFileHistory(tx, repoID, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error getting file history: %v", err),
			})
			return
		}

		// Determine file type
		isCode := isCodeFile(file.Filename)
		isBinary := isBinaryFile(content)

		var processedContent string
		var isDiff bool
		changeType := "M"

		if history == nil {
			// New file
			changeType = "A"
			isDiff = false
			if isBinary {
				processedContent = string(content)
			} else if isCode {
				processedContent = processCodeFile(string(content))
			} else {
				processedContent = string(content)
			}
		} else {
			// Existing file
			if isBinary {
				// For binary files, always store full content
				processedContent = string(content)
				isDiff = false
			} else {
				// For text files, store diff
				var currentContent string
				if isCode {
					currentContent = processCodeFile(string(content))
				} else {
					currentContent = string(content)
				}

				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(history.Content, currentContent, false)
				patches := dmp.PatchMake(diffs)
				processedContent = dmp.PatchToText(patches)
				isDiff = true
			}
		}

		processedFiles = append(processedFiles, FileCommit{
			Path:       file.Filename,
			Content:    processedContent,
			IsCode:     isCode,
			IsBinary:   isBinary,
			ChangeType: changeType,
			IsDiff:     isDiff,
		})
	}

	// Create commit record
	var commitID int
	err = tx.QueryRow(`
		INSERT INTO commit_history (repository_id, user_id, commit_hash, commit_message, tags)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, repoID, userID, utils.GenerateCommitHash(), message, pq.Array(tags)).Scan(&commitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create commit"})
		return
	}

	// Insert file changes
	stmt, err := tx.Prepare(`
		INSERT INTO commit_details (commit_id, file_path, change_type, content_changes)
		VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare statement"})
		return
	}
	defer stmt.Close()

	for _, file := range processedFiles {
		contentChanges := map[string]interface{}{
			"content":   file.Content,
			"is_code":   file.IsCode,
			"is_binary": file.IsBinary,
			"is_diff":   file.IsDiff,
		}
		contentJSON, err := json.Marshal(contentChanges)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error encoding file content: %v", err),
			})
			return
		}

		_, err = stmt.Exec(commitID, file.Path, file.ChangeType, contentJSON)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error saving file changes: %v", err),
			})
			return
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Commit created successfully",
		"commit_id": commitID,
		"files":     processedFiles,
	})
}
