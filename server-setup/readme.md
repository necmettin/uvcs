# UVCS Project Setup and Run Instructions

## Prerequisites
- Docker and Docker Compose installed
- Git (optional, for cloning)

## Project Structure
```
.
├── backend/           # Go application code
│   ├── handlers/     # API endpoint handlers
│   ├── models/       # Data models
│   ├── main.go      # Main application entry
│   └── go.mod       # Go module file
└── server-setup/     # Docker and deployment configs
    ├── Dockerfile   # Go application container
    ├── docker-compose.yml  # Service orchestration
    ├── .env         # Environment variables
    └── postgres-data/     # PostgreSQL data directory
```

## API Documentation

### Authentication Endpoints

1. Register User
   ```http
   POST /register
   Content-Type: application/x-www-form-urlencoded

   Form data:
   - username: string (required)
   - password: string (required)
   - firstname: string (required)
   - lastname: string (required)
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
   Note: Save the `skey1` and `skey2` values securely. They are required for authenticating API requests.

2. Login
   ```http
   POST /login
   Content-Type: application/x-www-form-urlencoded

   Form data:
   - username: string (required)
   - password: string (required)
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

### Example Usage with cURL

```bash
# Register a new user
curl -X POST http://localhost:8080/register \
  -d "username=johndoe" \
  -d "password=secretpass" \
  -d "firstname=John" \
  -d "lastname=Doe"

# Login
curl -X POST http://localhost:8080/login \
  -d "username=johndoe" \
  -d "password=secretpass"
```

### Branch Management Endpoints

All branch management endpoints require authentication using `skey1` and `skey2` obtained during registration or login.

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

### Example Branch API Usage with cURL

```bash
# List branches
curl -X POST http://localhost:8080/api/branches/list \
  -d "skey1=your_skey1&skey2=your_skey2"

# Create branch
curl -X POST http://localhost:8080/api/branches/create \
  -d "skey1=your_skey1&skey2=your_skey2&name=feature/auth"

# Delete branch
curl -X POST http://localhost:8080/api/branches/delete/feature/old \
  -d "skey1=your_skey1&skey2=your_skey2"

# List commits
curl -X POST http://localhost:8080/api/branches/develop/commits \
  -d "skey1=your_skey1&skey2=your_skey2"
```

## Environment Variables
The application uses these environment variables (configured in .env file):
- PORT=80 (container port)
- POSTGRES_USER=postgres
- POSTGRES_PASSWORD=postgres
- POSTGRES_DB=uvcs
- DB_HOST=postgres
- DB_PORT=5432

To modify any of these values, edit the .env file before starting the services.

## Running the Project

1. First time setup:
```bash
# Create PostgreSQL data directory
mkdir -p server-setup/postgres-data
```

2. Start the services:
```bash
# Navigate to server-setup directory
cd server-setup

# Build and start services
docker-compose up --build
```

The services will be available at:
- UVCS API: http://localhost:8080
- PostgreSQL: localhost:5432

## Stopping the Project
```bash
# In the server-setup directory
docker-compose down
```

## Development
- The API runs on port 80 inside the container, mapped to 8080 on your host
- PostgreSQL data persists in ./postgres-data directory
- Logs are visible in the docker-compose output
- Environment variables can be modified in the .env file 