package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"uvcs/handlers"
	"uvcs/modules/commands"
	"uvcs/modules/db"

	"github.com/gin-gonic/gin"
)

func main() {
	// Command line flags
	var (
		listBranches bool
		createBranch string
		deleteBranch string
		listCommits  string
	)

	// Long and short forms for flags
	flag.BoolVar(&listBranches, "list-branches", false, "List all branches")
	flag.BoolVar(&listBranches, "lb", false, "List all branches (shorthand)")

	flag.StringVar(&createBranch, "create-branch", "", "Create a new branch with given name")
	flag.StringVar(&createBranch, "cb", "", "Create a new branch with given name (shorthand)")

	flag.StringVar(&deleteBranch, "delete-branch", "", "Delete branch with given name")
	flag.StringVar(&deleteBranch, "db", "", "Delete branch with given name (shorthand)")

	flag.StringVar(&listCommits, "list-commits", "", "List commits for given branch name")
	flag.StringVar(&listCommits, "lc", "", "List commits for given branch name (shorthand)")

	flag.Parse()

	// Initialize database with CLI flag
	isCLI := flag.NFlag() > 0
	if err := db.InitDB(isCLI); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Handle command line operations
	if isCLI {
		var err error
		if listBranches {
			err = commands.ListBranches()
		} else if createBranch != "" {
			err = commands.CreateBranch(createBranch)
		} else if deleteBranch != "" {
			err = commands.DeleteBranch(deleteBranch)
		} else if listCommits != "" {
			err = commands.ListCommits(listCommits)
		}

		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}
		return
	}

	// If no command line args, run web server
	r := gin.Default()

	// Auth routes
	r.POST("/register", handlers.HandleRegister)
	r.POST("/login", handlers.HandleLogin)

	// Branch management routes
	r.POST("/api/branches/list", handlers.HandleListBranches)
	r.POST("/api/branches/create", handlers.HandleCreateBranch)
	r.POST("/api/branches/delete/:name", handlers.HandleDeleteBranch)
	r.POST("/api/branches/:name/commits", handlers.HandleListCommits)

	// Get port from environment variable or default to 80
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	// Run server
	r.Run(fmt.Sprintf(":%s", port))
}
