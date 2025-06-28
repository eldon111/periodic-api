package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// getEnvOrDefault returns the environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitDB initializes the database connection without running migrations
func InitDB() (*sql.DB, error) {
	// Get database connection details from environment variables or use defaults
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPortStr := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "eldon")
	dbPass := getEnvOrDefault("DB_PASSWORD", "moron")
	dbName := getEnvOrDefault("DB_NAME", "periodic_db")

	// Convert port to integer
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	// Build connection string with SSL mode based on environment
	sslMode := getEnvOrDefault("DB_SSL_MODE", "disable")
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, sslMode)

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return db, nil
}
