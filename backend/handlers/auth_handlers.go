package handlers

import (
	"database/sql"
	"net/http"
	"uvcs/modules/db"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	Username  string
	FirstName string
	LastName  string
	Skey1     string
	Skey2     string
	IsActive  bool
}

// HandleLogin handles the login request
func HandleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	var user User
	var hashedPassword string
	err := db.DB.QueryRow(`
		SELECT id, username, firstname, lastname, password, skey1, skey2, is_active
		FROM users 
		WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &hashedPassword, &user.Skey1, &user.Skey2, &user.IsActive)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled. Please contact administrator."})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"firstname": user.FirstName,
			"lastname":  user.LastName,
			"skey1":     user.Skey1,
			"skey2":     user.Skey2,
		},
	})
}
