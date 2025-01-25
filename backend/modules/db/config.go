package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Type     string // "postgres" or "sqlite3"
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Path     string // For SQLite3
}

func loadConfig(isCLI bool) (*DBConfig, error) {
	// Load .env file if it exists
	godotenv.Load()

	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		if isCLI {
			dbType = "sqlite3"
		} else {
			dbType = "postgres"
		}
	}

	if dbType == "sqlite3" {
		dbPath := os.Getenv("SQLITE_PATH")
		if dbPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("error getting home directory: %v", err)
			}
			dbPath = filepath.Join(homeDir, ".uvcs")
		}

		// Ensure directory exists
		err := os.MkdirAll(filepath.Dir(dbPath), 0755)
		if err != nil {
			return nil, fmt.Errorf("error creating database directory: %v", err)
		}

		return &DBConfig{
			Type: "sqlite3",
			Path: dbPath,
		}, nil
	}

	// PostgreSQL configuration
	return &DBConfig{
		Type:     "postgres",
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("POSTGRES_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("POSTGRES_DB", "uvcs"),
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *DBConfig) GetDSN() string {
	if c.Type == "sqlite3" {
		return c.Path
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)
}
