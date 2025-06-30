package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"periodic-api/internal/db"
	"periodic-api/internal/models"
	"periodic-api/internal/store"
)

func init() {
	// Set the application's default timezone to UTC
	time.Local = time.UTC
}

func main() {
	var itemStore store.ScheduledItemStore
	var todoStore store.TodoItemStore
	var executionLogStore store.ExecutionLogStore

	// Check environment variable to determine which store to use
	usePostgres := os.Getenv("USE_POSTGRES_DB")

	if strings.ToLower(usePostgres) == "true" {
		// Initialize database connection for PostgreSQL
		database, err := db.InitDB()
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer database.Close()

		// Create PostgreSQL store instances
		itemStore = store.NewPostgresScheduledItemStore(database)
		todoStore = store.NewPostgresTodoItemStore(database)
		executionLogStore = store.NewPostgresExecutionLogStore(database)
		log.Println("Scheduler using PostgreSQL database for storage")
	} else {
		// Create in-memory store instances
		itemStore = store.NewMemoryScheduledItemStore()
		todoStore = store.NewMemoryTodoItemStore()
		executionLogStore = store.NewMemoryExecutionLogStore()
		log.Println("Scheduler using in-memory database for storage")
	}

	// Get interval from environment variable, default to 30 seconds
	interval := 30 * time.Second
	if intervalStr := os.Getenv("SCHEDULER_INTERVAL"); intervalStr != "" {
		if parsedInterval, err := time.ParseDuration(intervalStr); err == nil {
			interval = parsedInterval
		} else {
			log.Printf("Invalid SCHEDULER_INTERVAL format, using default: %v", interval)
		}
	}

	log.Printf("Starting scheduler service with interval: %v", interval)

	// Create a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create ticker for periodic execution
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run initial check
	processScheduledItems(itemStore, todoStore, executionLogStore)

	// Main service loop
	for {
		select {
		case <-ticker.C:
			processScheduledItems(itemStore, todoStore, executionLogStore)
		case <-sigChan:
			log.Println("Received shutdown signal, stopping scheduler...")
			return
		}
	}
}

func processScheduledItems(store store.ScheduledItemStore, todoStore store.TodoItemStore, logStore store.ExecutionLogStore) {
	log.Println("Processing scheduled items...")

	// Get items that are due for execution using the optimized query
	// Use a reasonable limit for batch processing
	itemsDue, err := store.GetNextScheduledItems(100, 0)
	if err != nil {
		log.Printf("Error getting scheduled items due for execution: %v", err)
		return
	}

	log.Printf("Found %d items due for execution", len(itemsDue))

	// Process each item due for execution
	var successCount, errorCount int
	for _, item := range itemsDue {
		log.Printf("Processing item: ID=%d, Title='%s', NextExecutionAt=%v", 
			item.ID, item.Title, item.NextExecutionAt)

		// Create todo item from scheduled item
		todoText := createTodoText(item)
		todoItem := models.TodoItem{
			Text:    todoText,
			Checked: false,
		}

		createdTodo := todoStore.CreateTodoItem(todoItem)
		if createdTodo.ID > 0 {
			successCount++
			log.Printf("Created todo item ID=%d: '%s' for scheduled item ID=%d", 
				createdTodo.ID, createdTodo.Text, item.ID)
		} else {
			errorCount++
			log.Printf("Failed to create todo item for scheduled item ID=%d", item.ID)
		}

		// TODO: Update next execution time and log execution (next chunks)
	}

	if successCount > 0 || errorCount > 0 {
		log.Printf("Processed %d items: %d successful, %d errors", 
			len(itemsDue), successCount, errorCount)
	}

	log.Println("Finished processing scheduled items")
}

// createTodoText generates a descriptive todo item text from a scheduled item
func createTodoText(item models.ScheduledItem) string {
	// Create a meaningful todo text based on the scheduled item
	if item.Description != "" {
		// If there's a description, use both title and description
		return fmt.Sprintf("%s - %s", item.Title, item.Description)
	}
	
	// If no description, just use the title
	return item.Title
}
