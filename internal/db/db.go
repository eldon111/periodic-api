package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Database connection details
const (
	// Update these with your actual PostgreSQL instance details
	dbHost = "localhost" // Use your Google Cloud SQL IP when deploying
	dbPort = 5432
	dbUser = "postgres"
	dbPass = "your-password"
	dbName = "scheduled_items_db"
)

// InitDB initializes the database connection and creates the necessary tables
func InitDB() (*sql.DB, error) {
	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	// Read the SQL script from file
	sqlScript, err := os.ReadFile("db_init.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to read db_init.sql: %w", err)
	}

	// Execute the SQL script
	_, err = db.Exec(string(sqlScript))
	if err != nil {
		return nil, fmt.Errorf("failed to execute db_init.sql: %w", err)
	}

	return db, nil
}
