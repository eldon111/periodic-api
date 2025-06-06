package store

import (
	"awesomeProject/models"
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/lib/pq"
)

// TestTodoItemDB holds the connection to the test database
var testTodoItemDB *sql.DB

// setupTodoItemDBForTest sets up the test database for todo item tests
func setupTodoItemDBForTest(t *testing.T) {
	// Skip if we already have a test DB from another test
	if testTodoItemDB != nil {
		return
	}
	
	var err error
	testTodoItemDB, err = setupTodoItemTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	
	// Register cleanup function
	t.Cleanup(func() {
		cleanupTodoItemTestDB(testTodoItemDB)
		testTodoItemDB.Close()
		testTodoItemDB = nil
	})
}

// setupTodoItemTestDB creates a connection to the test database
// In a real implementation, you would use an in-memory PostgreSQL database
// or a Docker container with PostgreSQL for testing
func setupTodoItemTestDB() (*sql.DB, error) {
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

	// Create the todo_items table
	_, err = db.Exec(`
		DROP TABLE IF EXISTS todo_items;
		CREATE TABLE todo_items (
			id SERIAL PRIMARY KEY,
			text TEXT NOT NULL,
			checked BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo_items table: %w", err)
	}

	return db, nil
}

// cleanupTodoItemTestDB cleans up the test database
func cleanupTodoItemTestDB(db *sql.DB) {
	// Drop the todo_items table
	_, err := db.Exec("DROP TABLE IF EXISTS todo_items")
	if err != nil {
		log.Printf("Failed to drop todo_items table: %v", err)
	}
}

// TestTodoItemStore_CRUD tests the CRUD operations of the todo item store
func TestTodoItemStore_CRUD(t *testing.T) {
	// Set up the test database
	setupTodoItemDBForTest(t)
	
	// Skip the test if the database connection is not available
	if testTodoItemDB == nil {
		t.Skip("Test database connection not available")
	}

	// Create a new todo item store with the test database connection
	store := NewPostgresTodoItemStore(testTodoItemDB)

	// Test todo item
	testItem := models.TodoItem{
		Text:    "Test Todo Item",
		Checked: false,
	}

	// Test CreateTodoItem
	t.Run("CreateTodoItem", func(t *testing.T) {
		createdItem := store.CreateTodoItem(testItem)
		if createdItem.ID == 0 {
			t.Errorf("Expected item ID to be non-zero, got %d", createdItem.ID)
		}
		if createdItem.Text != testItem.Text {
			t.Errorf("Expected text to be %s, got %s", testItem.Text, createdItem.Text)
		}
		if createdItem.Checked != testItem.Checked {
			t.Errorf("Expected checked to be %v, got %v", testItem.Checked, createdItem.Checked)
		}

		// Save the ID for later tests
		testItem.ID = createdItem.ID
	})

	// Test GetTodoItem
	t.Run("GetTodoItem", func(t *testing.T) {
		retrievedItem, found := store.GetTodoItem(testItem.ID)
		if !found {
			t.Errorf("Expected to find item with ID %d, but not found", testItem.ID)
		}
		if retrievedItem.ID != testItem.ID {
			t.Errorf("Expected item ID to be %d, got %d", testItem.ID, retrievedItem.ID)
		}
		if retrievedItem.Text != testItem.Text {
			t.Errorf("Expected text to be %s, got %s", testItem.Text, retrievedItem.Text)
		}
		if retrievedItem.Checked != testItem.Checked {
			t.Errorf("Expected checked to be %v, got %v", testItem.Checked, retrievedItem.Checked)
		}
	})

	// Test GetAllTodoItems
	t.Run("GetAllTodoItems", func(t *testing.T) {
		items := store.GetAllTodoItems()
		found := false
		for _, item := range items {
			if item.ID == testItem.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find item with ID %d in GetAllTodoItems result, but not found", testItem.ID)
		}
	})

	// Test UpdateTodoItem
	t.Run("UpdateTodoItem", func(t *testing.T) {
		updatedItem := models.TodoItem{
			Text:    "Updated Todo Item",
			Checked: true,
		}

		result, success := store.UpdateTodoItem(testItem.ID, updatedItem)
		if !success {
			t.Errorf("Failed to update item with ID %d", testItem.ID)
		}
		if result.ID != testItem.ID {
			t.Errorf("Expected item ID to be %d, got %d", testItem.ID, result.ID)
		}
		if result.Text != updatedItem.Text {
			t.Errorf("Expected text to be %s, got %s", updatedItem.Text, result.Text)
		}
		if result.Checked != updatedItem.Checked {
			t.Errorf("Expected checked to be %v, got %v", updatedItem.Checked, result.Checked)
		}

		// Verify the update
		retrievedItem, found := store.GetTodoItem(testItem.ID)
		if !found {
			t.Errorf("Expected to find item with ID %d after update, but not found", testItem.ID)
		}
		if retrievedItem.Text != updatedItem.Text {
			t.Errorf("Expected text to be %s after update, got %s", updatedItem.Text, retrievedItem.Text)
		}
		if retrievedItem.Checked != updatedItem.Checked {
			t.Errorf("Expected checked to be %v after update, got %v", updatedItem.Checked, retrievedItem.Checked)
		}
	})

	// Test DeleteTodoItem
	t.Run("DeleteTodoItem", func(t *testing.T) {
		success := store.DeleteTodoItem(testItem.ID)
		if !success {
			t.Errorf("Failed to delete item with ID %d", testItem.ID)
		}

		// Verify the deletion
		_, found := store.GetTodoItem(testItem.ID)
		if found {
			t.Errorf("Expected item with ID %d to be deleted, but it was found", testItem.ID)
		}
	})
}