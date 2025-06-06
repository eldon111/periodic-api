package main

// This is an example of how to use testcontainers for integration testing
// Run with: go run test_container_example.go

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}
	defer postgresContainer.Terminate(ctx)

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("Failed to get container port: %v", err)
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), "testuser", "testpassword", "testdb")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create a test table
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create test table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO test_table (name) VALUES ($1)", "test_name")
	if err != nil {
		log.Fatalf("Failed to insert test data: %v", err)
	}

	// Query test data
	var id int
	var name string
	err = db.QueryRow("SELECT id, name FROM test_table WHERE name = $1", "test_name").Scan(&id, &name)
	if err != nil {
		log.Fatalf("Failed to query test data: %v", err)
	}

	fmt.Printf("✅ Testcontainer example successful!\n")
	fmt.Printf("Connected to PostgreSQL container at %s:%s\n", host, port.Port())
	fmt.Printf("Retrieved data: ID=%d, Name=%s\n", id, name)
	fmt.Printf("Container will be automatically cleaned up when this program exits.\n")
}