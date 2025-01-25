package handlers_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"uvcs/modules/handlers"
	"uvcs/modules/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleRegister(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name       string
		username   string
		email      string
		password   string
		statusCode int
		errorMsg   string
	}{
		{
			name:       "valid registration",
			username:   "testuser",
			email:      "test@example.com",
			password:   "password123",
			statusCode: http.StatusOK,
			errorMsg:   "",
		},
		{
			name:       "missing username",
			username:   "",
			email:      "test@example.com",
			password:   "password123",
			statusCode: http.StatusBadRequest,
			errorMsg:   "username is required",
		},
		{
			name:       "missing email",
			username:   "testuser",
			email:      "",
			password:   "password123",
			statusCode: http.StatusBadRequest,
			errorMsg:   "email is required",
		},
		{
			name:       "missing password",
			username:   "testuser",
			email:      "test@example.com",
			password:   "",
			statusCode: http.StatusBadRequest,
			errorMsg:   "password is required",
		},
		{
			name:       "duplicate email",
			username:   "testuser2",
			email:      "test@example.com",
			password:   "password123",
			statusCode: http.StatusBadRequest,
			errorMsg:   "email already exists",
		},
		{
			name:       "duplicate username",
			username:   "testuser",
			email:      "test2@example.com",
			password:   "password123",
			statusCode: http.StatusBadRequest,
			errorMsg:   "username already exists",
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", handlers.HandleRegister)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("username", test.username)
			form.Add("email", test.email)
			form.Add("password", test.password)

			req := httptest.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.statusCode, w.Code)

			if test.errorMsg != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, test.errorMsg, response["error"])
			}
		})
	}
}

func TestHandleLogin(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	// Create a test user
	user := testutils.CreateTestUser(t, db)

	tests := []struct {
		name       string
		identifier string
		password   string
		statusCode int
		errorMsg   string
	}{
		{
			name:       "valid login with username",
			identifier: user.Username,
			password:   "testpassword",
			statusCode: http.StatusOK,
			errorMsg:   "",
		},
		{
			name:       "valid login with email",
			identifier: user.Email,
			password:   "testpassword",
			statusCode: http.StatusOK,
			errorMsg:   "",
		},
		{
			name:       "missing identifier",
			identifier: "",
			password:   "testpassword",
			statusCode: http.StatusBadRequest,
			errorMsg:   "username or email is required",
		},
		{
			name:       "missing password",
			identifier: user.Username,
			password:   "",
			statusCode: http.StatusBadRequest,
			errorMsg:   "password is required",
		},
		{
			name:       "invalid credentials",
			identifier: user.Username,
			password:   "wrongpassword",
			statusCode: http.StatusUnauthorized,
			errorMsg:   "invalid credentials",
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", handlers.HandleLogin)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("identifier", test.identifier)
			form.Add("password", test.password)

			req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.statusCode, w.Code)

			if test.errorMsg != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, test.errorMsg, response["error"])
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["skey1"])
				assert.NotEmpty(t, response["skey2"])
			}
		})
	}
}

func TestHandleGetRepository(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	owner := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, owner.ID)
	commit := testutils.CreateTestCommit(t, db, repo.ID, owner.ID)
	detail := testutils.CreateTestCommitDetail(t, db, commit.ID)

	tests := []struct {
		name        string
		repoName    string
		user        *testutils.TestUser
		statusCode  int
		shouldError bool
	}{
		{
			name:        "owner access",
			repoName:    repo.Name,
			user:        owner,
			statusCode:  http.StatusOK,
			shouldError: false,
		},
		{
			name:        "read access",
			repoName:    repo.Name,
			user:        testutils.CreateTestUser(t, db),
			statusCode:  http.StatusOK,
			shouldError: false,
		},
		{
			name:        "no access",
			repoName:    repo.Name,
			user:        testutils.CreateTestUser(t, db),
			statusCode:  http.StatusForbidden,
			shouldError: true,
		},
		{
			name:        "invalid auth",
			repoName:    repo.Name,
			user:        nil,
			statusCode:  http.StatusUnauthorized,
			shouldError: true,
		},
		{
			name:        "missing name",
			repoName:    "",
			user:        owner,
			statusCode:  http.StatusBadRequest,
			shouldError: true,
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/repository", handlers.HandleGetRepository)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/repository?name="+test.repoName, nil)
			if test.user != nil {
				req.Header.Add("skey1", test.user.Skey1)
				req.Header.Add("skey2", test.user.Skey2)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.statusCode, w.Code)

			if !test.shouldError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Contains(t, response, "branches")
				assert.Contains(t, response, "commits")
				assert.Contains(t, response, "content")
				assert.Contains(t, response, "access")

				commits := response["commits"].([]interface{})
				assert.NotEmpty(t, commits)

				firstCommit := commits[0].(map[string]interface{})
				assert.Contains(t, firstCommit, "id")
				assert.Contains(t, firstCommit, "hash")
				assert.Contains(t, firstCommit, "message")
				assert.Contains(t, firstCommit, "datetime")
				assert.Contains(t, firstCommit, "author")
				assert.Contains(t, firstCommit, "changes")
			}
		})
	}
}

func TestHandleCommit(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	owner := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, owner.ID)
	writeUser := testutils.CreateTestUser(t, db)
	readUser := testutils.CreateTestUser(t, db)

	testutils.GrantRepositoryAccess(t, db, repo.ID, writeUser.ID, "write")
	testutils.GrantRepositoryAccess(t, db, repo.ID, readUser.ID, "read")

	tests := []struct {
		name        string
		repoName    string
		message     string
		user        *testutils.TestUser
		files       []testutils.TestFile
		statusCode  int
		shouldError bool
	}{
		{
			name:     "valid commit with new files",
			repoName: repo.Name,
			message:  "test commit",
			user:     owner,
			files: []testutils.TestFile{
				{Name: "test.go", Content: []byte("package main\n\nfunc main() {}\n")},
				{Name: "README.md", Content: []byte("# Test Repository")},
			},
			statusCode:  http.StatusOK,
			shouldError: false,
		},
		{
			name:     "valid commit with write access",
			repoName: repo.Name,
			message:  "test commit",
			user:     writeUser,
			files: []testutils.TestFile{
				{Name: "test2.go", Content: []byte("package main\n\nfunc test() {}\n")},
			},
			statusCode:  http.StatusOK,
			shouldError: false,
		},
		{
			name:     "commit without write access",
			repoName: repo.Name,
			message:  "test commit",
			user:     readUser,
			files: []testutils.TestFile{
				{Name: "test3.go", Content: []byte("package main\n\nfunc test() {}\n")},
			},
			statusCode:  http.StatusForbidden,
			shouldError: true,
		},
		{
			name:        "commit with invalid auth",
			repoName:    repo.Name,
			message:     "test commit",
			user:        nil,
			files:       []testutils.TestFile{},
			statusCode:  http.StatusUnauthorized,
			shouldError: true,
		},
		{
			name:        "commit without files",
			repoName:    repo.Name,
			message:     "test commit",
			user:        owner,
			files:       []testutils.TestFile{},
			statusCode:  http.StatusBadRequest,
			shouldError: true,
		},
		{
			name:     "commit without message",
			repoName: repo.Name,
			message:  "",
			user:     owner,
			files: []testutils.TestFile{
				{Name: "test4.go", Content: []byte("package main\n\nfunc test() {}\n")},
			},
			statusCode:  http.StatusBadRequest,
			shouldError: true,
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/repository/commit", handlers.HandleCommit)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			writer.WriteField("name", test.repoName)
			writer.WriteField("message", test.message)

			for _, file := range test.files {
				part, err := writer.CreateFormFile("files", file.Name)
				assert.NoError(t, err)
				_, err = part.Write(file.Content)
				assert.NoError(t, err)
			}

			writer.Close()

			req := httptest.NewRequest("POST", "/api/repository/commit", body)
			req.Header.Add("Content-Type", writer.FormDataContentType())
			if test.user != nil {
				req.Header.Add("skey1", test.user.Skey1)
				req.Header.Add("skey2", test.user.Skey2)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.statusCode, w.Code)

			if !test.shouldError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "commit_id")
				assert.Contains(t, response, "commit_hash")
			}
		})
	}
}

func TestHandleCommitWithDiffs(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

	owner := testutils.CreateTestUser(t, db)
	repo := testutils.CreateTestRepository(t, db, owner.ID)

	// First commit with initial content
	initialContent := []byte("package main\n\nfunc main() {}\n")
	firstCommit := testutils.CreateTestCommit(t, db, repo.ID, owner.ID)
	testutils.CreateTestCommitDetail(t, db, firstCommit.ID, "test.go", "A", string(initialContent))

	// Modified content for second commit
	modifiedContent := []byte("package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/repository/commit", handlers.HandleCommit)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("name", repo.Name)
	writer.WriteField("message", "Update main function")

	part, err := writer.CreateFormFile("files", "test.go")
	assert.NoError(t, err)
	_, err = part.Write(modifiedContent)
	assert.NoError(t, err)

	writer.Close()

	req := httptest.NewRequest("POST", "/api/repository/commit", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("skey1", owner.Skey1)
	req.Header.Add("skey2", owner.Skey2)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "commit_id")

	// Verify that a diff was created in the database
	var isDiff bool
	err = db.QueryRow(`
		SELECT is_diff FROM commit_details
		WHERE commit_id = $1 AND file_path = 'test.go'
	`, response["commit_id"]).Scan(&isDiff)
	assert.NoError(t, err)
	assert.True(t, isDiff, "Second commit should store a diff")
}
