package main

import (
	"testing"
	"time"

	"periodic-api/internal/models"
	"periodic-api/internal/store"
)

// TestCompleteSchedulerWorkflow tests the entire scheduler workflow using in-memory stores
func TestCompleteSchedulerWorkflow(t *testing.T) {
	// Create in-memory stores
	itemStore := store.NewMemoryScheduledItemStore()
	todoStore := store.NewMemoryTodoItemStore()
	logStore := store.NewMemoryExecutionLogStore()

	t.Run("Complete workflow with multiple item types", func(t *testing.T) {
		// Set up test data
		now := time.Now()
		pastTime := now.Add(-time.Hour)
		futureTime := now.Add(time.Hour)
		cronExpr := "0 */2 * * *" // Every 2 hours
		expiration := now.Add(24 * time.Hour)

		// Create different types of scheduled items
		items := []models.ScheduledItem{
			{
				Title:           "One-time urgent task",
				Description:     "Needs immediate attention",
				StartsAt:        pastTime,
				Repeats:         false,
				NextExecutionAt: pastTime, // Due now
			},
			{
				Title:           "Regular cleanup",
				Description:     "Clean temporary files",
				StartsAt:        pastTime,
				Repeats:         true,
				CronExpression:  &cronExpr,
				Expiration:      &expiration,
				NextExecutionAt: pastTime, // Due now
			},
			{
				Title:           "Future task",
				Description:     "Not yet due",
				StartsAt:        futureTime,
				Repeats:         false,
				NextExecutionAt: futureTime, // Not due yet
			},
		}

		// Create all items
		var createdItems []models.ScheduledItem
		for _, item := range items {
			created := itemStore.CreateScheduledItem(item)
			if created.ID == 0 {
				t.Fatalf("Failed to create scheduled item: %s", item.Title)
			}
			createdItems = append(createdItems, created)
		}

		// Record initial state
		initialTodos := len(todoStore.GetAllTodoItems())
		initialLogs := len(logStore.GetAllExecutionLogs())
		initialItems := len(itemStore.GetAllScheduledItems())

		// Execute the main scheduler processing function
		processScheduledItems(itemStore, todoStore, logStore)

		// Verify results
		finalTodos := todoStore.GetAllTodoItems()
		finalLogs := logStore.GetAllExecutionLogs()
		finalItems := itemStore.GetAllScheduledItems()

		// Should have created 2 todo items (one-time + repeating item that were due)
		expectedTodos := initialTodos + 2
		if len(finalTodos) != expectedTodos {
			t.Errorf("Expected %d todos, got %d", expectedTodos, len(finalTodos))
		}

		// Should have 2 execution logs
		expectedLogs := initialLogs + 2
		if len(finalLogs) != expectedLogs {
			t.Errorf("Expected %d logs, got %d", expectedLogs, len(finalLogs))
		}

		// Should have one less item (one-time item deleted, repeating item updated, future item unchanged)
		expectedItems := initialItems - 1
		if len(finalItems) != expectedItems {
			t.Errorf("Expected %d scheduled items, got %d", expectedItems, len(finalItems))
		}

		// Verify todo content
		todoTexts := make(map[string]bool)
		for _, todo := range finalTodos {
			todoTexts[todo.Text] = true
		}

		expectedTodoTexts := []string{
			"One-time urgent task - Needs immediate attention",
			"Regular cleanup - Clean temporary files",
		}

		for _, expectedText := range expectedTodoTexts {
			if !todoTexts[expectedText] {
				t.Errorf("Expected todo text not found: %s", expectedText)
			}
		}

		// Verify execution logs
		var successLogs, errorLogs int
		for _, log := range finalLogs {
			switch log.Status {
			case "success":
				successLogs++
			case "error":
				errorLogs++
			}
		}

		if successLogs < 2 {
			t.Errorf("Expected at least 2 success logs, got %d", successLogs)
		}

		if errorLogs > 0 {
			t.Errorf("Expected 0 error logs, got %d", errorLogs)
		}

		// Verify the one-time item was deleted
		_, exists := itemStore.GetScheduledItem(createdItems[0].ID)
		if exists {
			t.Error("One-time item should have been deleted")
		}

		// Verify the repeating item still exists and was updated
		repeatingItem, exists := itemStore.GetScheduledItem(createdItems[1].ID)
		if !exists {
			t.Error("Repeating item should still exist")
		} else if !repeatingItem.NextExecutionAt.After(now) {
			t.Error("Repeating item should have future next execution time")
		}

		// Verify the future item was not processed
		futureItem, exists := itemStore.GetScheduledItem(createdItems[2].ID)
		if !exists {
			t.Error("Future item should still exist")
		} else if !futureItem.NextExecutionAt.Equal(futureTime) {
			t.Error("Future item's next execution time should be unchanged")
		}
	})

	t.Run("Handle empty queue gracefully", func(t *testing.T) {
		// Clear all scheduled items
		allItems := itemStore.GetAllScheduledItems()
		for _, item := range allItems {
			itemStore.DeleteScheduledItem(item.ID)
		}

		// Record initial state
		initialTodos := len(todoStore.GetAllTodoItems())
		initialLogs := len(logStore.GetAllExecutionLogs())

		// Process with empty queue
		processScheduledItems(itemStore, todoStore, logStore)

		// Verify no changes
		finalTodos := len(todoStore.GetAllTodoItems())
		finalLogs := len(logStore.GetAllExecutionLogs())

		if finalTodos != initialTodos {
			t.Errorf("Expected %d todos (no change), got %d", initialTodos, finalTodos)
		}

		if finalLogs != initialLogs {
			t.Errorf("Expected %d logs (no change), got %d", initialLogs, finalLogs)
		}
	})

	t.Run("Handle processing errors gracefully", func(t *testing.T) {
		// This test would require mocking the todo store to simulate failures
		// For now, we've tested the error logging functionality in unit tests
		t.Skip("Error simulation requires mocking - covered by unit tests")
	})
}
