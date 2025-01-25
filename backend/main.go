package main

import (
	"fmt"
	"os"
	"uvcs/handlers"
	"uvcs/modules/db"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	db.InitDB()

	r := gin.Default()

	// Routes
	r.POST("/register", handlers.HandleRegister)
	r.POST("/login", handlers.HandleLogin)

	// Get port from environment variable or default to 80
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	// Run server
	r.Run(fmt.Sprintf(":%s", port))
}
