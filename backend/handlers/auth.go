package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"uvcs/modules/db"
	modules "uvcs/modules/password"

	"github.com/gin-gonic/gin"
)

func HandleRegister(c *gin.Context) {
	// Get form data
	firstName := c.PostForm("firstname")
	lastName := c.PostForm("lastname")
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Basic validation
	if firstName == "" || lastName == "" || email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	// Hash password
	hashedPassword, err := modules.Hash(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Insert user into database
	var id int
	err = db.DB.QueryRow(`
		INSERT INTO users (firstname, lastname, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, firstName, lastName, email, hashedPassword).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func HandleLogin(c *gin.Context) {
	// Get form data
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Basic validation
	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	// Find user
	var user struct {
		ID        int
		FirstName string
		LastName  string
		Email     string
		Password  string
	}

	err := db.DB.QueryRow(`
		SELECT id, firstname, lastname, email, password
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	match, err := modules.Verify(password, user.Password)
	if err != nil || !match {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate session keys
	skey1, err := generateRandomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session key"})
		return
	}

	skey2, err := generateRandomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session key"})
		return
	}

	// Get client IP
	clientIP := c.ClientIP()

	// Record login in active_logins table
	_, err = db.DB.Exec(`
		INSERT INTO active_logins (user_id, ip_address, skey1, skey2)
		VALUES ($1, $2, $3, $4)
	`, user.ID, clientIP, skey1, skey2)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"firstname": user.FirstName,
		"lastname":  user.LastName,
		"email":     user.Email,
		"skey1":     skey1,
		"skey2":     skey2,
	})
}

// Helper function to generate random strings for session keys
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
