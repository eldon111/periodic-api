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
			expiration TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}
