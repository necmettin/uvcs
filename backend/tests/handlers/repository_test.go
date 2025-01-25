package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"uvcs/handlers"
	"uvcs/modules/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleGetRepository(t *testing.T) {
	// Setup
	tdb, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Create test users and repository
	ownerID, _, _, ownerSkey1, ownerSkey2 := tdb.CreateTestUser(t)
	userID, _, _, userSkey1, userSkey2 := tdb.CreateTestUser(t)
	repoID, repoName := tdb.CreateTestRepository(t, ownerID)

	// Create some test commits
	commitID := tdb.CreateTestCommit(t, repoID, ownerID, "Initial commit", []string{"v1.0"})
	tdb.CreateTestCommitDetail(t, commitID, "main.go", "package main", true, false, false)
	tdb.CreateTestCommitDetail(t, commitID, "README.md", "# Test", false, false, false)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/repository", handlers.HandleGetRepository)

	tests := []struct {
		name       string
		form       url.Values
		setup      func()
		wantStatus int
		wantError  bool
	}{
		{
			name: "owner access",
			form: url.Values{
				"name":  {repoName},
				"skey1": {ownerSkey1},
				"skey2": {ownerSkey2},
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "read access",
			form: url.Values{
				"name":  {repoName},
				"skey1": {userSkey1},
				"skey2": {userSkey2},
			},
			setup: func() {
				tdb.GrantRepositoryAccess(t, repoID, userID, "read")
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "no access",
			form: url.Values{
				"name":  {repoName},
				"skey1": {userSkey1},
				"skey2": {userSkey2},
			},
			wantStatus: http.StatusForbidden,
			wantError:  true,
		},
		{
			name: "invalid auth",
			form: url.Values{
				"name":  {repoName},
				"skey1": {"invalid"},
				"skey2": {"invalid"},
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name: "missing name",
			form: url.Values{
				"skey1": {ownerSkey1},
				"skey2": {ownerSkey2},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/repository", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.wantError {
				assert.Contains(t, response, "error")
			} else {
				// Verify repository data
				assert.Contains(t, response, "branches")
				assert.Contains(t, response, "commits")
				assert.Contains(t, response, "content")
				assert.Contains(t, response, "access")

				// Verify commits
				commits := response["commits"].([]interface{})
				assert.Equal(t, 1, len(commits))
				commit := commits[0].(map[string]interface{})
				assert.Contains(t, commit, "hash")
				assert.Contains(t, commit, "message")
				assert.Contains(t, commit, "datetime")
				assert.Contains(t, commit, "tags")
				assert.Contains(t, commit, "author")
				assert.Contains(t, commit, "changes")

				// Verify file changes
				changes := commit["changes"].([]interface{})
				assert.Equal(t, 2, len(changes))
				for _, change := range changes {
					changeMap := change.(map[string]interface{})
					assert.Contains(t, changeMap, "file_path")
					assert.Contains(t, changeMap, "change_type")
					assert.Contains(t, changeMap, "content_change")
				}
			}
		})
	}
}
