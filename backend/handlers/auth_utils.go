package handlers

import (
	"database/sql"
	"fmt"
	"uvcs/modules/db"

	"github.com/gin-gonic/gin"
)

// authenticateRequest verifies the user's authentication using skey1 and skey2
func authenticateRequest(c *gin.Context) (int, error) {
	skey1 := c.PostForm("skey1")
	skey2 := c.PostForm("skey2")

	if skey1 == "" || skey2 == "" {
		return 0, fmt.Errorf("missing authentication keys")
	}

	var userID int
	var isActive bool
	err := db.DB.QueryRow(`
		SELECT id, is_active 
		FROM users 
		WHERE skey1 = $1 AND skey2 = $2
	`, skey1, skey2).Scan(&userID, &isActive)

	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid authentication keys")
	}
	if err != nil {
		return 0, fmt.Errorf("database error: %v", err)
	}
	if !isActive {
		return 0, fmt.Errorf("account is disabled")
	}

	return userID, nil
}
