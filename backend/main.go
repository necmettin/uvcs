package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
		listUsers    bool
		enableUser   string
		disableUser  string
		createUser   string
		userEmail    string
		userPass     string
		firstName    string
		lastName     string

		// Repository management flags
		createRepo   string
		repoDesc     string
		repoOwner    string
		listRepos    string
		grantAccess  string
		revokeAccess string
		listAccess   string
		accessUser   string
		accessLevel  string
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

	// User management flags
	flag.BoolVar(&listUsers, "list-users", false, "List all users")
	flag.BoolVar(&listUsers, "lu", false, "List all users (shorthand)")

	flag.StringVar(&createUser, "create-user", "", "Create a new user with given username (optional if email is provided)")
	flag.StringVar(&createUser, "cu", "", "Create a new user with given username (shorthand)")
	flag.StringVar(&userEmail, "email", "", "Email for new user (optional if username is provided)")
	flag.StringVar(&userEmail, "e", "", "Email for new user (shorthand)")
	flag.StringVar(&userPass, "password", "", "Password for new user")
	flag.StringVar(&userPass, "p", "", "Password for new user (shorthand)")
	flag.StringVar(&firstName, "firstname", "", "First name for new user")
	flag.StringVar(&firstName, "f", "", "First name for new user (shorthand)")
	flag.StringVar(&lastName, "lastname", "", "Last name for new user")
	flag.StringVar(&lastName, "l", "", "Last name for new user (shorthand)")

	flag.StringVar(&enableUser, "enable-user", "", "Enable user by username or email")
	flag.StringVar(&enableUser, "eu", "", "Enable user by username (shorthand)")

	flag.StringVar(&disableUser, "disable-user", "", "Disable user by username")
	flag.StringVar(&disableUser, "du", "", "Disable user by username (shorthand)")

	// Repository management flags
	flag.StringVar(&createRepo, "create-repo", "", "Create a new repository")
	flag.StringVar(&createRepo, "cr", "", "Create a new repository (shorthand)")
	flag.StringVar(&repoDesc, "desc", "", "Repository description (used with --create-repo)")
	flag.StringVar(&repoOwner, "owner", "", "Repository owner (used with --create-repo)")

	flag.StringVar(&listRepos, "list-repos", "", "List repositories for given username")
	flag.StringVar(&listRepos, "lr", "", "List repositories for given username (shorthand)")

	flag.StringVar(&grantAccess, "grant-access", "", "Grant repository access (format: owner/repo)")
	flag.StringVar(&grantAccess, "ga", "", "Grant repository access (shorthand)")

	flag.StringVar(&revokeAccess, "revoke-access", "", "Revoke repository access (format: owner/repo)")
	flag.StringVar(&revokeAccess, "ra", "", "Revoke repository access (shorthand)")

	flag.StringVar(&listAccess, "list-access", "", "List repository access (format: owner/repo)")
	flag.StringVar(&listAccess, "la", "", "List repository access (shorthand)")

	flag.StringVar(&accessUser, "user", "", "Username for access operations")
	flag.StringVar(&accessLevel, "level", "read", "Access level (read/write) for grant operations")

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
		} else if listUsers {
			err = commands.ListUsers()
		} else if enableUser != "" {
			err = commands.EnableUser(enableUser)
		} else if disableUser != "" {
			err = commands.DisableUser(disableUser)
		} else if createUser != "" || userEmail != "" {
			if userPass == "" {
				log.Fatal("Password (--password) is required")
			}
			if firstName == "" {
				log.Fatal("First name (--firstname) is required")
			}
			if lastName == "" {
				log.Fatal("Last name (--lastname) is required")
			}
			if createUser == "" && userEmail == "" {
				log.Fatal("Either username (--create-user) or email (--email) is required")
			}

			identifier := createUser
			identifierType := "username"
			if createUser == "" {
				identifier = userEmail
				identifierType = "email"
			}
			err = commands.CreateUser(identifier, identifierType, userPass, firstName, lastName)
		} else if createRepo != "" {
			if repoOwner == "" {
				log.Fatal("Repository owner (--owner) is required")
			}
			err = commands.CreateRepository(createRepo, repoDesc, repoOwner)
		} else if listRepos != "" {
			err = commands.ListRepositories(listRepos)
		} else if grantAccess != "" {
			if accessUser == "" {
				log.Fatal("Username (--user) is required")
			}
			parts := strings.Split(grantAccess, "/")
			if len(parts) != 2 {
				log.Fatal("Repository must be in format owner/repo")
			}
			err = commands.GrantAccess(parts[0], parts[1], accessUser, accessLevel)
		} else if revokeAccess != "" {
			if accessUser == "" {
				log.Fatal("Username (--user) is required")
			}
			parts := strings.Split(revokeAccess, "/")
			if len(parts) != 2 {
				log.Fatal("Repository must be in format owner/repo")
			}
			err = commands.RevokeAccess(parts[0], parts[1], accessUser)
		} else if listAccess != "" {
			parts := strings.Split(listAccess, "/")
			if len(parts) != 2 {
				log.Fatal("Repository must be in format owner/repo")
			}
			err = commands.ListAccess(parts[0], parts[1])
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

	// Repository routes
	r.POST("/api/repository", handlers.HandleGetRepository)

	// Get port from environment variable or default to 80
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	// Run server
	r.Run(fmt.Sprintf(":%s", port))
}
