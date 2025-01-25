package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"uvcs/modules/db"
	"uvcs/modules/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// HandleRegister handles user registration
func HandleRegister(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	firstname := c.PostForm("firstname")
	lastname := c.PostForm("lastname")

	if password == "" || firstname == "" || lastname == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields",
		})
		return
	}

	if username == "" && email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either username or email is required",
		})
		return
	}

	// Generate security keys
	skey1 := utils.GenerateSecurityKey()
	skey2 := utils.GenerateSecurityKey()

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error hashing password: %v", err),
		})
		return
	}

	// Create user
	var userID int
	var query string
	var args []interface{}

	if username != "" && email != "" {
		query = `
			INSERT INTO users (username, email, password, firstname, lastname, skey1, skey2, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, true)
			RETURNING id`
		args = []interface{}{username, email, hashedPassword, firstname, lastname, skey1, skey2}
	} else if email != "" {
		query = `
			INSERT INTO users (email, password, firstname, lastname, skey1, skey2, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, true)
			RETURNING id`
		args = []interface{}{email, hashedPassword, firstname, lastname, skey1, skey2}
	} else {
		query = `
			INSERT INTO users (username, password, firstname, lastname, skey1, skey2, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, true)
			RETURNING id`
		args = []interface{}{username, hashedPassword, firstname, lastname, skey1, skey2}
	}

	err = db.DB.QueryRow(query, args...).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error creating user: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":        userID,
			"username":  username,
			"email":     email,
			"firstname": firstname,
			"lastname":  lastname,
			"skey1":     skey1,
			"skey2":     skey2,
		},
	})
}

// HandleLogin handles user login
func HandleLogin(c *gin.Context) {
	identifier := c.PostForm("identifier") // can be username or email
	password := c.PostForm("password")

	if identifier == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing identifier or password",
		})
		return
	}

	// Get user by username or email
	var (
		userID    int
		username  sql.NullString
		email     sql.NullString
		firstname string
		lastname  string
		skey1     string
		skey2     string
		dbPass    string
		isActive  bool
	)

	err := db.DB.QueryRow(`
		SELECT id, username, email, firstname, lastname, password, skey1, skey2, is_active
		FROM users
		WHERE (username = $1 OR email = $1)
	`, identifier).Scan(&userID, &username, &email, &firstname, &lastname, &dbPass, &skey1, &skey2, &isActive)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Database error: %v", err),
		})
		return
	}

	if !isActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Account is disabled",
		})
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(dbPass), []byte(password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":        userID,
			"username":  username.String,
			"email":     email.String,
			"firstname": firstname,
			"lastname":  lastname,
			"skey1":     skey1,
			"skey2":     skey2,
		},
	})
}
