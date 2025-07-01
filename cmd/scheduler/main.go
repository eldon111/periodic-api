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
	"periodic-api/internal/utils"
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

	// Early return if no items to process
	if len(itemsDue) == 0 {
		log.Println("No items due for execution")
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

			// Update next execution time after successful todo creation
			updateProcessedScheduledItem(store, item)

			// Log successful execution
			logExecution(logStore, item.ID, "success", nil, &createdTodo.ID)
		} else {
			errorCount++
			errorMsg := "Failed to create todo item"
			log.Printf("%s for scheduled item ID=%d", errorMsg, item.ID)

			// Log failed execution
			logExecution(logStore, item.ID, "error", &errorMsg, nil)
		}
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

// updateProcessedScheduledItem calculates and updates the next execution time for a scheduled item
func updateProcessedScheduledItem(store store.ScheduledItemStore, item models.ScheduledItem) {
	if !item.Repeats {
		if store.DeleteScheduledItem(item.ID) {
			log.Printf("Deleted completed non-repeating item ID=%d", item.ID)
		} else {
			log.Printf("Failed to delete completed item ID=%d", item.ID)
		}
	}

	// For repeating items, calculate the next execution based on cron expression
	nextExec := utils.CalculateNextExecution(item.StartsAt, item.Repeats, item.CronExpression, item.Expiration)
	if nextExec != nil {
		success := store.UpdateNextExecutionAt(item.ID, *nextExec)
		if success {
			log.Printf("Updated next execution for repeating item ID=%d to %v", item.ID, *nextExec)
		} else {
			log.Printf("Failed to update next execution for item ID=%d", item.ID)
		}
	} else {
		// Repeating item has expired or no valid next execution
		if store.DeleteScheduledItem(item.ID) {
			log.Printf("Deleted expired repeating item ID=%d", item.ID)
		} else {
			log.Printf("Failed to delete expired item ID=%d", item.ID)
		}
	}
}

// logExecution creates an execution log entry for a scheduled item processing attempt
func logExecution(logStore store.ExecutionLogStore, scheduledItemID int64, status string, errorMessage *string, todoItemID *int64) {
	// Validate input parameters
	if scheduledItemID <= 0 {
		log.Printf("Invalid scheduled item ID for execution log: %d", scheduledItemID)
		return
	}

	if status != "success" && status != "error" && status != "skipped" {
		log.Printf("Invalid status for execution log: %s", status)
		return
	}

	executionLog := models.ExecutionLog{
		ScheduledItemID: scheduledItemID,
		ExecutedAt:      time.Now(),
		Status:          status,
		ErrorMessage:    errorMessage,
		TodoItemID:      todoItemID,
	}

	createdLog := logStore.CreateExecutionLog(executionLog)
	if createdLog.ID > 0 {
		if status == "success" && todoItemID != nil {
			log.Printf("Logged successful execution: log ID=%d, scheduled item ID=%d, todo item ID=%d",
				createdLog.ID, scheduledItemID, *todoItemID)
		} else if status == "error" && errorMessage != nil {
			log.Printf("Logged failed execution: log ID=%d, scheduled item ID=%d, error: %s",
				createdLog.ID, scheduledItemID, *errorMessage)
		} else {
			log.Printf("Logged execution: log ID=%d, scheduled item ID=%d, status: %s",
				createdLog.ID, scheduledItemID, status)
		}
	} else {
		log.Printf("Failed to create execution log for scheduled item ID=%d", scheduledItemID)
	}
}
