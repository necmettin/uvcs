package handlers_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"uvcs/handlers"
	"uvcs/modules/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleCommit(t *testing.T) {
	// Setup
	tdb, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Create test users and repository
	ownerID, _, _, ownerSkey1, ownerSkey2 := tdb.CreateTestUser(t)
	userID, _, _, userSkey1, userSkey2 := tdb.CreateTestUser(t)
	repoID, repoName := tdb.CreateTestRepository(t, ownerID)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/commit", handlers.HandleCommit)

	tests := []struct {
		name       string
		setup      func() (*bytes.Buffer, *multipart.Writer)
		auth       map[string]string
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid commit with new files",
			setup: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.WriteField("name", repoName)
				writer.WriteField("message", "Initial commit")
				writer.WriteField("tags", "v1.0,initial")

				// Add a code file
				part, _ := writer.CreateFormFile("files[]", "main.go")
				part.Write([]byte("package main\n\nfunc main() {}\n"))

				// Add a text file
				part, _ = writer.CreateFormFile("files[]", "README.md")
				part.Write([]byte("# Test Repository\n"))

				writer.Close()
				return body, writer
			},
			auth: map[string]string{
				"skey1": ownerSkey1,
				"skey2": ownerSkey2,
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "valid commit with write access",
			setup: func() (*bytes.Buffer, *multipart.Writer) {
				tdb.GrantRepositoryAccess(t, repoID, userID, "write")
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.WriteField("name", repoName)
				writer.WriteField("message", "Update README")

				part, _ := writer.CreateFormFile("files[]", "README.md")
				part.Write([]byte("# Updated Test Repository\n"))

				writer.Close()
				return body, writer
			},
			auth: map[string]string{
				"skey1": userSkey1,
				"skey2": userSkey2,
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "commit without write access",
			setup: func() (*bytes.Buffer, *multipart.Writer) {
				tdb.GrantRepositoryAccess(t, repoID, userID, "read")
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.WriteField("name", repoName)
				writer.WriteField("message", "Unauthorized commit")

				part, _ := writer.CreateFormFile("files[]", "test.txt")
				part.Write([]byte("test\n"))

				writer.Close()
				return body, writer
			},
			auth: map[string]string{
				"skey1": userSkey1,
				"skey2": userSkey2,
			},
			wantStatus: http.StatusForbidden,
			wantError:  true,
		},
		{
			name: "commit with invalid auth",
			setup: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.WriteField("name", repoName)
				writer.WriteField("message", "Invalid auth")

				part, _ := writer.CreateFormFile("files[]", "test.txt")
				part.Write([]byte("test\n"))

				writer.Close()
				return body, writer
			},
			auth: map[string]string{
				"skey1": "invalid",
				"skey2": "invalid",
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name: "commit without files",
			setup: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.WriteField("name", repoName)
				writer.WriteField("message", "Empty commit")
				writer.Close()
				return body, writer
			},
			auth: map[string]string{
				"skey1": ownerSkey1,
				"skey2": ownerSkey2,
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "commit without message",
			setup: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.WriteField("name", repoName)

				part, _ := writer.CreateFormFile("files[]", "test.txt")
				part.Write([]byte("test\n"))

				writer.Close()
				return body, writer
			},
			auth: map[string]string{
				"skey1": ownerSkey1,
				"skey2": ownerSkey2,
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, writer := tt.setup()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/commit", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			for k, v := range tt.auth {
				req.PostForm.Set(k, v)
			}
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.wantError {
				assert.Contains(t, response, "error")
			} else {
				assert.Contains(t, response, "commit_id")
				assert.Contains(t, response, "commit_hash")
			}
		})
	}
}

func TestHandleCommitWithDiffs(t *testing.T) {
	// Setup
	tdb, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Create test users and repository
	ownerID, _, _, ownerSkey1, ownerSkey2 := tdb.CreateTestUser(t)
	repoID, repoName := tdb.CreateTestRepository(t, ownerID)

	// Create initial commit
	commitID := tdb.CreateTestCommit(t, repoID, ownerID, "Initial commit", []string{"v1.0"})
	tdb.CreateTestCommitDetail(t, commitID, "main.go", "package main\n\nfunc main() {}\n", true, false, false)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/commit", handlers.HandleCommit)

	// Test commit with modified file
	t.Run("commit with file modifications", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("name", repoName)
		writer.WriteField("message", "Update main function")

		part, _ := writer.CreateFormFile("files[]", "main.go")
		part.Write([]byte("package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n"))

		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/commit", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.PostForm.Set("skey1", ownerSkey1)
		req.PostForm.Set("skey2", ownerSkey2)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "commit_id")
		assert.Contains(t, response, "commit_hash")

		// Verify that a diff was created
		rows, err := tdb.DB.Query(`
			SELECT cd.is_diff, cd.content_changes
			FROM commit_details cd
			JOIN commit_history ch ON cd.commit_id = ch.id
			WHERE ch.id = $1 AND cd.file_path = 'main.go'
		`, response["commit_id"])
		assert.NoError(t, err)
		defer rows.Close()

		assert.True(t, rows.Next())
		var isDiff bool
		var contentChanges string
		err = rows.Scan(&isDiff, &contentChanges)
		assert.NoError(t, err)
		assert.True(t, isDiff)
		assert.Contains(t, contentChanges, "fmt.Println")
	})
}

func TestIsCodeFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"main.go", true},
		{"script.js", true},
		{"style.css", false},
		{"README.md", false},
		{"image.png", false},
		{"test.py", true},
		{"program.java", true},
		{"header.h", true},
		{"source.cpp", true},
		{"module.rs", true},
		{"script.rb", true},
		{"index.php", true},
		{"program.cs", true},
		{"app.swift", true},
		{"service.kt", true},
		{"main.scala", true},
		{"objc.m", true},
		{"objc.mm", true},
		{".gitignore", false},
		{"Dockerfile", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := handlers.IsCodeFile(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{
			name:    "text file",
			content: []byte("Hello, World!"),
			want:    false,
		},
		{
			name:    "binary file with null byte",
			content: []byte{72, 101, 108, 108, 111, 0, 87, 111, 114, 108, 100},
			want:    true,
		},
		{
			name:    "empty file",
			content: []byte{},
			want:    false,
		},
		{
			name:    "unicode text",
			content: []byte("Hello, 世界!"),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.IsBinaryFile(tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProcessCodeFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "remove indentation and empty lines",
			content: `package main

			import "fmt"

			func main() {
				fmt.Println("Hello")
				
				fmt.Println("World")
			}`,
			expected: "package main\nimport \"fmt\"\nfunc main() {\nfmt.Println(\"Hello\")\nfmt.Println(\"World\")\n}",
		},
		{
			name:     "already clean",
			content:  "package main\nfunc main() {\n}",
			expected: "package main\nfunc main() {\n}",
		},
		{
			name: "multiple empty lines",
			content: `line1


			line2


			line3`,
			expected: "line1\nline2\nline3",
		},
		{
			name:     "only whitespace",
			content:  "   \n\t\n  \n",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.ProcessCodeFile(tt.content)
			assert.Equal(t, tt.expected, got)
		})
	}
}
