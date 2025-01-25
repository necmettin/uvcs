package handlers

import (
	"net/http"
	"uvcs/modules/commands"
	"uvcs/modules/db"

	"github.com/gin-gonic/gin"
)

// validateKeys checks if the provided skey1 and skey2 are valid
func validateKeys(c *gin.Context) bool {
	skey1 := c.PostForm("skey1")
	skey2 := c.PostForm("skey2")

	if skey1 == "" || skey2 == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "skey1 and skey2 are required"})
		return false
	}

	var exists bool
	err := db.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE skey1 = $1 AND skey2 = $2
		)
	`, skey1, skey2).Scan(&exists)

	if err != nil || !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication keys"})
		return false
	}

	return true
}

// HandleListBranches handles GET /api/branches
func HandleListBranches(c *gin.Context) {
	if !validateKeys(c) {
		return
	}

	branches, err := commands.ListBranchesAPI()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branches": branches})
}

// HandleCreateBranch handles POST /api/branches
func HandleCreateBranch(c *gin.Context) {
	if !validateKeys(c) {
		return
	}

	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Branch name is required"})
		return
	}

	err := commands.CreateBranch(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Branch created successfully"})
}

// HandleDeleteBranch handles DELETE /api/branches/:name
func HandleDeleteBranch(c *gin.Context) {
	if !validateKeys(c) {
		return
	}

	name := c.Param("name")
	err := commands.DeleteBranch(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Branch deleted successfully"})
}

// HandleListCommits handles GET /api/branches/:name/commits
func HandleListCommits(c *gin.Context) {
	if !validateKeys(c) {
		return
	}

	name := c.Param("name")
	commits, err := commands.ListCommitsAPI(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"commits": commits})
}
