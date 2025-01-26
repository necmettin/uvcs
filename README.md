# UVCS - Universal Version Control System

UVCS is a versatile version control system that operates both as a command-line tool and a web server. It provides functionality for managing code branches, commits, and user authentication.

## Building the Application

```bash
cd backend
go build
```

This will create the `uvcs` binary in the backend directory.

## Database Configuration

UVCS supports both SQLite3 and PostgreSQL databases. The choice of database can be configured using environment variables:

```bash
# Database type (sqlite3 or postgres)
DB_TYPE=sqlite3  # CLI default
DB_TYPE=postgres # Server default

# SQLite3 configuration
SQLITE_PATH=/path/to/database  # Default: ~/.uvcs

# PostgreSQL configuration
DB_HOST=localhost      # Default: localhost
DB_PORT=5432          # Default: 5432
POSTGRES_USER=user    # Default: postgres
POSTGRES_PASSWORD=pwd # Default: postgres
POSTGRES_DB=dbname    # Default: uvcs
```

You can set these variables in a `.env` file or in your environment. The CLI defaults to SQLite3 for portability, while the server defaults to PostgreSQL for better concurrent access.

## Command Line Interface (CLI)

The application can be used as a CLI tool with the following commands:

### Repository Management

1. Create Repository
   ```bash
   ./uvcs --create-repo myproject --owner johndoe --desc "My awesome project"
   # or shorter form:
   ./uvcs --cr myproject --owner johndoe --desc "My awesome project"
   ```
   Creates a new repository owned by the specified user.

2. List Repositories
   ```bash
   ./uvcs --list-repos johndoe  # or shorter form: --lr johndoe
   ```
   Shows all repositories the user has access to, including ownership status.

3. Grant Access
   ```bash
   ./uvcs --grant-access johndoe/myproject --user janedoe --level write
   # or shorter form:
   ./uvcs --ga johndoe/myproject --user janedoe --level write
   ```
   Grants repository access to a user. Access levels can be 'read' or 'write'.

4. Revoke Access
   ```bash
   ./uvcs --revoke-access johndoe/myproject --user janedoe
   # or shorter form:
   ./uvcs --ra johndoe/myproject --user janedoe
   ```
   Revokes repository access from a user.

5. List Access
   ```bash
   ./uvcs --list-access johndoe/myproject  # or shorter form: --la johndoe/myproject
   ```
   Shows all users who have access to the repository, including their access levels.

### User Management

1. Create User
   ```bash
   # Create user with username
   ./uvcs --create-user johndoe --password secretpass --firstname John --lastname Doe
   # or shorter form:
   ./uvcs --cu johndoe -p secretpass -f John -l Doe

   # Create user with email
   ./uvcs --email john@example.com --password secretpass --firstname John --lastname Doe
   # or shorter form:
   ./uvcs -e john@example.com -p secretpass -f John -l Doe

   # Create user with both (recommended)
   ./uvcs --create-user johndoe --email john@example.com --password secretpass --firstname John --lastname Doe
   ```
   Creates a new user with the specified credentials. Either username or email must be provided.

2. List Users
   ```bash
   ./uvcs --list-users  # or shorter form: --lu
   ```
   Shows all users with their identifier (username or email), names, creation dates, and status.

3. Enable User
   ```bash
   # Enable by username
   ./uvcs --enable-user johndoe
   # or by email
   ./uvcs --enable-user john@example.com
   ```
   Enables a disabled user, allowing them to use the system.

4. Disable User
   ```bash
   # Disable by username
   ./uvcs --disable-user johndoe
   # or by email
   ./uvcs --disable-user john@example.com
   ```
   Disables an enabled user, preventing them from using the system.

### Branch Management

1. List Branches
   ```bash
   ./uvcs --list-branches  # or shorter form: --lb
   ```
   Shows all active branches with their IDs, names, descriptions, creation dates, and commit counts.

2. Create Branch
   ```bash
   ./uvcs --create-branch feature/auth  # or shorter form: --cb feature/auth
   ```
   Creates a new branch with the specified name.

3. Delete Branch
   ```bash
   ./uvcs --delete-branch feature/old  # or shorter form: --db feature/old
   ```
   Soft deletes a branch (marks it as inactive).

4. List Commits
   ```bash
   ./uvcs --list-commits develop  # or shorter form: --lc develop
   ```
   Shows all commits in the specified branch, including commit ID, date, author, message, and hash.

## Web Server

When run without command-line arguments, UVCS operates as a web server (default port 80).

### Authentication

The server provides two authentication endpoints:

1. Register (`POST /register`)
   ```bash
   curl -X POST http://localhost:80/register \
     -d "username=johndoe" \
     -d "password=secretpass" \
     -d "firstname=John" \
     -d "lastname=Doe"
   ```
   Response:
   ```json
   {
     "message": "User registered successfully",
     "user": {
       "id": 1,
       "username": "johndoe",
       "firstname": "John",
       "lastname": "Doe",
       "skey1": "generated_key_1",
       "skey2": "generated_key_2"
     }
   }
   ```

2. Login (`POST /login`)
   ```bash
   curl -X POST http://localhost:80/login \
     -d "username=johndoe" \
     -d "password=secretpass"
   ```
   Response:
   ```json
   {
     "message": "Login successful",
     "user": {
       "id": 1,
       "username": "johndoe",
       "firstname": "John",
       "lastname": "Doe",
       "skey1": "your_skey1",
       "skey2": "your_skey2"
     }
   }
   ```

Note: The `skey1` and `skey2` values returned by these endpoints are required for authenticating all other API requests.

### Branch Management API

All branch management endpoints require authentication using `skey1` and `skey2` in the POST form data.

1. List Branches
   ```http
   POST /api/branches/list
   Form data: skey1, skey2
   ```
   Response:
   ```json
   {
     "branches": [
       {
         "ID": 1,
         "Name": "develop",
         "Description": "Main development branch",
         "CreatedAt": "2024-03-20T10:00:00Z",
         "CommitIDs": [1, 2, 3],
         "HeadCommit": 3,
         "IsActive": true
       }
     ]
   }
   ```

2. Create Branch
   ```http
   POST /api/branches/create
   Form data: skey1, skey2, name
   ```
   Response:
   ```json
   {
     "message": "Branch created successfully"
   }
   ```

3. Delete Branch
   ```http
   POST /api/branches/delete/:name
   Form data: skey1, skey2
   ```
   Response:
   ```json
   {
     "message": "Branch deleted successfully"
   }
   ```

4. List Commits
   ```http
   POST /api/branches/:name/commits
   Form data: skey1, skey2
   ```
   Response:
   ```json
   {
     "commits": [
       {
         "id": 1,
         "hash": "abc123def456",
         "message": "Initial commit",
         "datetime": "2024-03-20T10:00:00Z",
         "author": "John Doe"
       }
     ]
   }
   ```

### Repository API

1. Get Repository Information
   ```http
   POST /api/repository
   Form data: 
   - skey1: string (required)
   - skey2: string (required)
   - owner: string (required) - repository owner's username
   - name: string (required) - repository name
   ```
   Response:
   ```json
   {
     "branches": [
       {
         "id": 1,
         "name": "develop",
         "description": "Main development branch",
         "created_at": "2024-03-20T10:00:00Z",
         "commit_ids": [1, 2, 3],
         "head_commit": 3,
         "is_active": true
       }
     ],
     "commits": [
       {
         "id": 1,
         "hash": "abc123def456",
         "message": "Initial commit",
         "datetime": "2024-03-20T10:00:00Z",
         "tags": ["v1.0.0", "stable"],
         "author": {
           "id": 1,
           "full_name": "John Doe"
         },
         "changes": [
           {
             "file_path": "src/main.go",
             "change_type": "A",
             "content_change": {
               "content": "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n",
               "ast_changes": [
                 {
                   "type": "package_declaration",
                   "name": "main"
                 },
                 {
                   "type": "function_declaration",
                   "name": "main",
                   "body": [
                     {
                       "type": "function_call",
                       "package": "fmt",
                       "function": "Println",
                       "arguments": ["Hello, World!"]
                     }
                   ]
                 }
               ]
             }
           }
         ]
       }
     ],
     "content": {
       "files": {
         "src/main.go": {
           "content": "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n",
           "commit_id": 1,
           "timestamp": "2024-03-20T10:00:00Z"
         }
       }
     },
     "access": {
       "level": "write",
       "users": [
         {
           "id": 1,
           "full_name": "John Doe",
           "access": "owner"
         },
         {
           "id": 2,
           "full_name": "Jane Doe",
           "access": "write"
         }
       ]
     }
   }
   ```

   The response includes:
   - All branches with their metadata
   - Complete commit history
   - Current repository content
   - Access control information

### Example API Usage with cURL

```bash
# List branches
curl -X POST http://localhost:80/api/branches/list \
  -d "skey1=your_skey1&skey2=your_skey2"

# Create branch
curl -X POST http://localhost:80/api/branches/create \
  -d "skey1=your_skey1&skey2=your_skey2&name=feature/auth"

# Delete branch
curl -X POST http://localhost:80/api/branches/delete/feature/old \
  -d "skey1=your_skey1&skey2=your_skey2"

# List commits
curl -X POST http://localhost:80/api/branches/develop/commits \
  -d "skey1=your_skey1&skey2=your_skey2"

# Get repository information
curl -X POST http://localhost:80/api/repository \
  -d "skey1=your_skey1&skey2=your_skey2"
```

## Environment Variables

- `PORT`: Web server port (default: 80)

## Testing

### Prerequisites
- PostgreSQL server running locally
- Test database named `uvcs_test` created
- Test user with username `postgres` and password `postgres`

### Running Tests
To run all tests:
```bash
cd backend
go test ./... -v
```

To run tests for a specific package:
```bash
# Test database package
go test ./modules/db -v

# Test utils package
go test ./modules/utils -v

# Test handlers
go test ./handlers -v

# Test commands
go test ./commands -v
```

### Test Coverage
To run tests with coverage:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Structure
The test suite includes:

1. **Unit Tests**
   - Authentication handlers (`handlers/auth_test.go`)
   - Repository handlers (`handlers/repository_test.go`)
   - Commit handlers (`handlers/commit_test.go`)
   - Utility functions (`modules/utils/utils_test.go`)
   - Database operations (`modules/db/db_test.go`)
   - CLI commands (`commands/commands_test.go`)

2. **Integration Tests**
   - Database schema and constraints
   - Repository operations
   - User authentication flow
   - Commit operations with diffs

3. **Test Utilities**
   - Test database setup and cleanup
   - Test user creation
   - Test repository creation
   - Test data generation

### Test Database
The test suite uses a separate database (`uvcs_test`) to avoid interfering with the production database. The test database is automatically set up with the required schema before running tests and cleaned up afterward.

## Running with Docker Compose

The application can be run in either development or production mode using Docker Compose.

### Development Mode

Development mode features:
- Hot reloading for both frontend (Vite) and backend (Air)
- Source code mounted as volumes for instant updates
- Development-specific nginx configuration for HMR
- Debug logging enabled

To start in development mode:
```bash
cd server-setup
docker-compose -f docker-compose.dev.yml up --build
```

The development environment will be available at:
- Frontend: http://localhost/ (with HMR)
- Backend API: http://localhost/api/
- Direct Vite access: http://localhost:5173
- Direct Backend access: http://localhost:8080

### Production Mode

Production mode features:
- Optimized builds for both frontend and backend
- Static file serving with caching
- Production-grade nginx configuration
- Release mode for better performance

To start in production mode:
```bash
cd server-setup
docker-compose up --build
```

The production environment will be available at:
- Frontend: http://localhost/
- Backend API: http://localhost/api/

### Stopping the Application

To stop either environment:
```bash
# For development
docker-compose -f docker-compose.dev.yml down

# For production
docker-compose down
```

To completely clean up (including volumes):
```bash
# For development
docker-compose -f docker-compose.dev.yml down -v

# For production
docker-compose down -v
```