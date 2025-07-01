package main

import (
	"testing"
	"time"

	"periodic-api/internal/models"
	"periodic-api/internal/store"
)

// Test the createTodoText function
func TestCreateTodoText(t *testing.T) {
	tests := []struct {
		name        string
		item        models.ScheduledItem
		expectedText string
	}{
		{
			name: "Item with title and description",
			item: models.ScheduledItem{
				Title:       "Daily Standup",
				Description: "Team daily standup meeting",
			},
			expectedText: "Daily Standup - Team daily standup meeting",
		},
		{
			name: "Item with title only",
			item: models.ScheduledItem{
				Title:       "Weekly Report",
				Description: "",
			},
			expectedText: "Weekly Report",
		},
		{
			name: "Item with empty description",
			item: models.ScheduledItem{
				Title:       "Code Review",
				Description: "",
			},
			expectedText: "Code Review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createTodoText(tt.item)
			if result != tt.expectedText {
				t.Errorf("Expected '%s', got '%s'", tt.expectedText, result)
			}
		})
	}
}

// Test the updateProcessedScheduledItem function with in-memory stores
func TestUpdateProcessedScheduledItem(t *testing.T) {
	// Create in-memory store for testing
	store := store.NewMemoryScheduledItemStore()

	t.Run("Delete non-repeating item", func(t *testing.T) {
		// Create a non-repeating item
		item := models.ScheduledItem{
			Title:           "One-time task",
			Description:     "Test task",
			StartsAt:        time.Now().Add(-time.Hour),
			Repeats:         false,
			NextExecutionAt: time.Now().Add(-time.Hour),
		}

		createdItem := store.CreateScheduledItem(item)
		if createdItem.ID == 0 {
			t.Fatal("Failed to create scheduled item")
		}

		// Verify item exists
		_, exists := store.GetScheduledItem(createdItem.ID)
		if !exists {
			t.Fatal("Item should exist before processing")
		}

		// Process the item
		updateProcessedScheduledItem(store, createdItem)

		// Verify item was deleted
		_, exists = store.GetScheduledItem(createdItem.ID)
		if exists {
			t.Error("Non-repeating item should be deleted after processing")
		}
	})

	t.Run("Update repeating item with valid next execution", func(t *testing.T) {
		// Create a repeating item with future executions
		cronExpr := "0 */6 * * *" // Every 6 hours
		futureExpiration := time.Now().Add(24 * time.Hour)
		
		item := models.ScheduledItem{
			Title:           "Repeating task",
			Description:     "Test repeating task",
			StartsAt:        time.Now().Add(-time.Hour),
			Repeats:         true,
			CronExpression:  &cronExpr,
			Expiration:      &futureExpiration,
			NextExecutionAt: time.Now().Add(-time.Hour),
		}

		createdItem := store.CreateScheduledItem(item)
		if createdItem.ID == 0 {
			t.Fatal("Failed to create scheduled item")
		}

		originalNextExecution := createdItem.NextExecutionAt

		// Process the item
		updateProcessedScheduledItem(store, createdItem)

		// Verify item still exists
		updatedItem, exists := store.GetScheduledItem(createdItem.ID)
		if !exists {
			t.Fatal("Repeating item should still exist after processing")
		}

		// Verify next execution time was updated
		if !updatedItem.NextExecutionAt.After(originalNextExecution) {
			t.Error("Next execution time should have been updated to a future time")
		}

		// Clean up
		store.DeleteScheduledItem(createdItem.ID)
	})

	t.Run("Delete expired repeating item", func(t *testing.T) {
		// Create an expired repeating item
		cronExpr := "0 */6 * * *" // Every 6 hours
		pastExpiration := time.Now().Add(-time.Hour) // Expired 1 hour ago
		
		item := models.ScheduledItem{
			Title:           "Expired repeating task",
			Description:     "Test expired task",
			StartsAt:        time.Now().Add(-2 * time.Hour),
			Repeats:         true,
			CronExpression:  &cronExpr,
			Expiration:      &pastExpiration,
			NextExecutionAt: time.Now().Add(-30 * time.Minute),
		}

		createdItem := store.CreateScheduledItem(item)
		if createdItem.ID == 0 {
			t.Fatal("Failed to create scheduled item")
		}

		// Verify item exists
		_, exists := store.GetScheduledItem(createdItem.ID)
		if !exists {
			t.Fatal("Item should exist before processing")
		}

		// Process the item
		updateProcessedScheduledItem(store, createdItem)

		// Verify item was deleted due to expiration
		_, exists = store.GetScheduledItem(createdItem.ID)
		if exists {
			t.Error("Expired repeating item should be deleted after processing")
		}
	})
}

// Test the logExecution function
func TestLogExecution(t *testing.T) {
	// Create in-memory execution log store for testing
	logStore := store.NewMemoryExecutionLogStore()

	t.Run("Log successful execution", func(t *testing.T) {
		scheduledItemID := int64(123)
		todoItemID := int64(456)
		
		initialLogCount := len(logStore.GetAllExecutionLogs())

		// Log successful execution
		logExecution(logStore, scheduledItemID, "success", nil, &todoItemID)

		// Verify log was created
		finalLogs := logStore.GetAllExecutionLogs()
		if len(finalLogs) != initialLogCount+1 {
			t.Fatalf("Expected %d logs, got %d", initialLogCount+1, len(finalLogs))
		}

		// Find our log entry
		var ourLog *models.ExecutionLog
		for _, log := range finalLogs {
			if log.ScheduledItemID == scheduledItemID {
				ourLog = &log
				break
			}
		}

		if ourLog == nil {
			t.Fatal("Log entry not found")
		}

		// Verify log details
		if ourLog.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", ourLog.Status)
		}

		if ourLog.TodoItemID == nil || *ourLog.TodoItemID != todoItemID {
			t.Errorf("Expected todo item ID %d, got %v", todoItemID, ourLog.TodoItemID)
		}

		if ourLog.ErrorMessage != nil {
			t.Errorf("Expected no error message, got '%s'", *ourLog.ErrorMessage)
		}
	})

	t.Run("Log failed execution", func(t *testing.T) {
		scheduledItemID := int64(789)
		errorMsg := "Test error message"
		
		initialLogCount := len(logStore.GetAllExecutionLogs())

		// Log failed execution
		logExecution(logStore, scheduledItemID, "error", &errorMsg, nil)

		// Verify log was created
		finalLogs := logStore.GetAllExecutionLogs()
		if len(finalLogs) != initialLogCount+1 {
			t.Fatalf("Expected %d logs, got %d", initialLogCount+1, len(finalLogs))
		}

		// Find our log entry
		var ourLog *models.ExecutionLog
		for _, log := range finalLogs {
			if log.ScheduledItemID == scheduledItemID {
				ourLog = &log
				break
			}
		}

		if ourLog == nil {
			t.Fatal("Log entry not found")
		}

		// Verify log details
		if ourLog.Status != "error" {
			t.Errorf("Expected status 'error', got '%s'", ourLog.Status)
		}

		if ourLog.ErrorMessage == nil || *ourLog.ErrorMessage != errorMsg {
			t.Errorf("Expected error message '%s', got %v", errorMsg, ourLog.ErrorMessage)
		}

		if ourLog.TodoItemID != nil {
			t.Errorf("Expected no todo item ID, got %v", *ourLog.TodoItemID)
		}
	})

	t.Run("Reject invalid parameters", func(t *testing.T) {
		initialLogCount := len(logStore.GetAllExecutionLogs())

		// Test invalid scheduled item ID
		logExecution(logStore, 0, "success", nil, nil)
		
		// Test invalid status
		logExecution(logStore, 123, "invalid_status", nil, nil)

		// Verify no logs were created
		finalLogs := logStore.GetAllExecutionLogs()
		if len(finalLogs) != initialLogCount {
			t.Errorf("Expected %d logs (no new logs), got %d", initialLogCount, len(finalLogs))
		}
	})
}