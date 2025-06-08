package db

import (
	"database/sql"
	"fmt"

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

	// Create the scheduled_items table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS scheduled_items (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			starts_at TIMESTAMP NOT NULL,
			repeats BOOLEAN NOT NULL,
			cron_expression TEXT,
			expiration TIMESTAMP,
			next_execution_at TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduled_items table: %w", err)
	}

	// Add next_execution_at column if it doesn't exist (for existing databases)
	_, err = db.Exec(`
		ALTER TABLE scheduled_items 
		ADD COLUMN IF NOT EXISTS next_execution_at TIMESTAMP
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to add next_execution_at column: %w", err)
	}

	// Create index on next_execution_at for efficient querying
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_scheduled_items_next_execution 
		ON scheduled_items (next_execution_at) 
		WHERE next_execution_at IS NOT NULL
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduled_items table: %w", err)
	}

	// Create the todo_items table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todo_items (
			id SERIAL PRIMARY KEY,
			text TEXT NOT NULL,
			checked BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo_items table: %w", err)
	}

	// Create the users table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL,
			password_hash BYTEA NOT NULL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create users table: %w", err)
	}

	return db, nil
}
