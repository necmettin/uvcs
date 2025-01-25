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

## API Endpoints

1. Register a new user:
```bash
curl -X POST http://localhost:8080/register \
  -d "firstname=John" \
  -d "lastname=Doe" \
  -d "email=john@example.com" \
  -d "password=secret123"
```

2. Login:
```bash
curl -X POST http://localhost:8080/login \
  -d "email=john@example.com" \
  -d "password=secret123"
```

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