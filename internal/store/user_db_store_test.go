package store

import (
	"periodic-api/internal/models"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// TestDB holds the connection to the test database
var testDB *sql.DB

// TestMain sets up the test database and runs all tests
func TestMain(m *testing.M) {
	// Set up the test database connection
	var err error
	testDB, err = setupTestDB()
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	defer testDB.Close()

	// Run the tests
	exitCode := m.Run()

	// Clean up the test database
	cleanupTestDB(testDB)

	os.Exit(exitCode)
}

// setupTestDB creates a connection to the test database
// In a real implementation, you would use an in-memory PostgreSQL database
// or a Docker container with PostgreSQL for testing
func setupTestDB() (*sql.DB, error) {
	// For demonstration purposes, we'll connect to a local PostgreSQL database
	// In a real implementation, you would use environment variables for these values
	dbHost := "localhost"
	dbPort := 5432
	dbUser := "postgres"
	dbPass := "your-password"
	dbName := "test_db"

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

	// Create the users table
	_, err = db.Exec(`
		DROP TABLE IF EXISTS users;
		CREATE TABLE users (
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

// cleanupTestDB cleans up the test database
func cleanupTestDB(db *sql.DB) {
	// Drop the users table
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		log.Printf("Failed to drop users table: %v", err)
	}
}

// TestUserStore_CRUD tests the CRUD operations of the user store
func TestUserStore_CRUD(t *testing.T) {
	// Skip the test if the database connection is not available
	if testDB == nil {
		t.Skip("Test database connection not available")
	}

	// Create a new user store with the test database connection
	store := NewPostgresUserStore(testDB)

	// Test user
	testUser := models.User{
		Username:     "testuser",
		PasswordHash: []byte("hashedpassword"),
	}

	// Test CreateUser
	t.Run("CreateUser", func(t *testing.T) {
		createdUser := store.CreateUser(testUser)
		if createdUser.ID == 0 {
			t.Errorf("Expected user ID to be non-zero, got %d", createdUser.ID)
		}
		if createdUser.Username != testUser.Username {
			t.Errorf("Expected username to be %s, got %s", testUser.Username, createdUser.Username)
		}

		// Save the ID for later tests
		testUser.ID = createdUser.ID
	})

	// Test GetUser
	t.Run("GetUser", func(t *testing.T) {
		retrievedUser, found := store.GetUser(testUser.ID)
		if !found {
			t.Errorf("Expected to find user with ID %d, but not found", testUser.ID)
		}
		if retrievedUser.ID != testUser.ID {
			t.Errorf("Expected user ID to be %d, got %d", testUser.ID, retrievedUser.ID)
		}
		if retrievedUser.Username != testUser.Username {
			t.Errorf("Expected username to be %s, got %s", testUser.Username, retrievedUser.Username)
		}
	})

	// Test GetAllUsers
	t.Run("GetAllUsers", func(t *testing.T) {
		users := store.GetAllUsers()
		found := false
		for _, user := range users {
			if user.ID == testUser.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find user with ID %d in GetAllUsers result, but not found", testUser.ID)
		}
	})

	// Test UpdateUser
	t.Run("UpdateUser", func(t *testing.T) {
		updatedUser := models.User{
			Username:     "updateduser",
			PasswordHash: []byte("updatedpassword"),
		}

		result, success := store.UpdateUser(testUser.ID, updatedUser)
		if !success {
			t.Errorf("Failed to update user with ID %d", testUser.ID)
		}
		if result.ID != testUser.ID {
			t.Errorf("Expected user ID to be %d, got %d", testUser.ID, result.ID)
		}
		if result.Username != updatedUser.Username {
			t.Errorf("Expected username to be %s, got %s", updatedUser.Username, result.Username)
		}

		// Verify the update
		retrievedUser, found := store.GetUser(testUser.ID)
		if !found {
			t.Errorf("Expected to find user with ID %d after update, but not found", testUser.ID)
		}
		if retrievedUser.Username != updatedUser.Username {
			t.Errorf("Expected username to be %s after update, got %s", updatedUser.Username, retrievedUser.Username)
		}
	})

	// Test DeleteUser
	t.Run("DeleteUser", func(t *testing.T) {
		success := store.DeleteUser(testUser.ID)
		if !success {
			t.Errorf("Failed to delete user with ID %d", testUser.ID)
		}

		// Verify the deletion
		_, found := store.GetUser(testUser.ID)
		if found {
			t.Errorf("Expected user with ID %d to be deleted, but it was found", testUser.ID)
		}
	})
}
