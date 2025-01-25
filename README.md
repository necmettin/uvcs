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
```

## Environment Variables

- `PORT`: Web server port (default: 80) 