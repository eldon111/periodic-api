package store

import (
	"periodic-api/internal/models"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestScheduledItemDB holds the connection to the test database
var testScheduledItemDB *sql.DB

// setupScheduledItemDBForTest sets up the test database for scheduled item tests
func setupScheduledItemDBForTest(t *testing.T) {
	// Skip if we already have a test DB from another test
	if testScheduledItemDB != nil {
		return
	}

	var err error
	testScheduledItemDB, err = setupScheduledItemTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}

	// Register cleanup function
	t.Cleanup(func() {
		cleanupScheduledItemTestDB(testScheduledItemDB)
		testScheduledItemDB.Close()
		testScheduledItemDB = nil
	})
}

// setupScheduledItemTestDB creates a connection to the test database
// In a real implementation, you would use an in-memory PostgreSQL database
// or a Docker container with PostgreSQL for testing
func setupScheduledItemTestDB() (*sql.DB, error) {
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

	// Create the scheduled_items table
	_, err = db.Exec(`
		DROP TABLE IF EXISTS scheduled_items;
		CREATE TABLE scheduled_items (
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
		return nil, fmt.Errorf("failed to create scheduled_items table: %w", err)
	}

	return db, nil
}

// cleanupScheduledItemTestDB cleans up the test database
func cleanupScheduledItemTestDB(db *sql.DB) {
	// Drop the scheduled_items table
	_, err := db.Exec("DROP TABLE IF EXISTS scheduled_items")
	if err != nil {
		log.Printf("Failed to drop scheduled_items table: %v", err)
	}
}

// TestScheduledItemStore_CRUD tests the CRUD operations of the scheduled item store
func TestScheduledItemStore_CRUD(t *testing.T) {
	// Set up the test database
	setupScheduledItemDBForTest(t)

	// Skip the test if the database connection is not available
	if testScheduledItemDB == nil {
		t.Skip("Test database connection not available")
	}

	// Create a new scheduled item store with the test database connection
	store := NewPostgresScheduledItemStore(testScheduledItemDB)

	// Test scheduled item
	now := time.Now()
	cronExpr := "0 0 * * *"            // Run at midnight every day
	expiration := now.AddDate(0, 1, 0) // 1 month from now

	testItem := models.ScheduledItem{
		Title:          "Test Item",
		Description:    "This is a test scheduled item",
		StartsAt:       now,
		Repeats:        true,
		CronExpression: &cronExpr,
		Expiration:     &expiration,
	}

	// Test CreateScheduledItem
	t.Run("CreateScheduledItem", func(t *testing.T) {
		createdItem := store.CreateScheduledItem(testItem)
		if createdItem.ID == 0 {
			t.Errorf("Expected item ID to be non-zero, got %d", createdItem.ID)
		}
		if createdItem.Title != testItem.Title {
			t.Errorf("Expected title to be %s, got %s", testItem.Title, createdItem.Title)
		}
		if createdItem.Description != testItem.Description {
			t.Errorf("Expected description to be %s, got %s", testItem.Description, createdItem.Description)
		}
		if !createdItem.StartsAt.Equal(testItem.StartsAt) {
			t.Errorf("Expected startsAt to be %v, got %v", testItem.StartsAt, createdItem.StartsAt)
		}
		if createdItem.Repeats != testItem.Repeats {
			t.Errorf("Expected repeats to be %v, got %v", testItem.Repeats, createdItem.Repeats)
		}
		if *createdItem.CronExpression != *testItem.CronExpression {
			t.Errorf("Expected cronExpression to be %s, got %s", *testItem.CronExpression, *createdItem.CronExpression)
		}
		if !createdItem.Expiration.Equal(*testItem.Expiration) {
			t.Errorf("Expected expiration to be %v, got %v", *testItem.Expiration, *createdItem.Expiration)
		}

		// Save the ID for later tests
		testItem.ID = createdItem.ID
	})

	// Test GetScheduledItem
	t.Run("GetScheduledItem", func(t *testing.T) {
		retrievedItem, found := store.GetScheduledItem(testItem.ID)
		if !found {
			t.Errorf("Expected to find item with ID %d, but not found", testItem.ID)
		}
		if retrievedItem.ID != testItem.ID {
			t.Errorf("Expected item ID to be %d, got %d", testItem.ID, retrievedItem.ID)
		}
		if retrievedItem.Title != testItem.Title {
			t.Errorf("Expected title to be %s, got %s", testItem.Title, retrievedItem.Title)
		}
		if retrievedItem.Description != testItem.Description {
			t.Errorf("Expected description to be %s, got %s", testItem.Description, retrievedItem.Description)
		}
	})

	// Test GetAllScheduledItems
	t.Run("GetAllScheduledItems", func(t *testing.T) {
		items := store.GetAllScheduledItems()
		found := false
		for _, item := range items {
			if item.ID == testItem.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find item with ID %d in GetAllScheduledItems result, but not found", testItem.ID)
		}
	})

	// Test DeleteScheduledItem
	t.Run("DeleteScheduledItem", func(t *testing.T) {
		success := store.DeleteScheduledItem(testItem.ID)
		if !success {
			t.Errorf("Failed to delete item with ID %d", testItem.ID)
		}

		// Verify the deletion
		_, found := store.GetScheduledItem(testItem.ID)
		if found {
			t.Errorf("Expected item with ID %d to be deleted, but it was found", testItem.ID)
		}
	})
}
