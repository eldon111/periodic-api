package integration

import (
	"log"
	"testing"
	"time"

	"periodic-api/internal/models"
	"periodic-api/internal/store"
	"periodic-api/internal/utils"
)

// Test the complete scheduler workflow
func TestSchedulerProcessWorkflow(t *testing.T) {
	skipIfDBNotAvailable(t)
	db := getActiveDB()

	// Create store instances
	itemStore := store.NewPostgresScheduledItemStore(db)
	todoStore := store.NewPostgresTodoItemStore(db)
	logStore := store.NewPostgresExecutionLogStore(db)

	// Clean up tables before test
	cleanupSchedulerTables(t, db)

	t.Run("ProcessOneTimeItem", func(t *testing.T) {
		testProcessOneTimeItem(t, itemStore, todoStore, logStore)
	})

	t.Run("ProcessRepeatingItem", func(t *testing.T) {
		testProcessRepeatingItem(t, itemStore, todoStore, logStore)
	})

	t.Run("ProcessExpiredItem", func(t *testing.T) {
		testProcessExpiredItem(t, itemStore, todoStore, logStore)
	})

	t.Run("ProcessMultipleItems", func(t *testing.T) {
		testProcessMultipleItems(t, itemStore, todoStore, logStore)
	})
}

// testProcessOneTimeItem tests processing a one-time scheduled item
func testProcessOneTimeItem(t *testing.T, itemStore store.ScheduledItemStore, todoStore store.TodoItemStore, logStore store.ExecutionLogStore) {
	// Create a one-time item due for execution
	now := time.Now()
	pastTime := now.Add(-time.Hour) // 1 hour ago

	item := models.ScheduledItem{
		Title:           "One-time task",
		Description:     "Test one-time scheduled item",
		StartsAt:        pastTime,
		Repeats:         false,
		NextExecutionAt: pastTime, // Due for execution
	}

	createdItem := itemStore.CreateScheduledItem(item)
	if createdItem.ID == 0 {
		t.Fatal("Failed to create scheduled item")
	}

	// Count initial state
	initialTodos := len(todoStore.GetAllTodoItems())
	initialLogs := len(logStore.GetAllExecutionLogs())

	// Simulate scheduler processing
	itemsDue, err := itemStore.GetNextScheduledItems(10, 0)
	if err != nil {
		t.Fatalf("Failed to get items due: %v", err)
	}

	if len(itemsDue) != 1 {
		t.Fatalf("Expected 1 item due, got %d", len(itemsDue))
	}

	dueItem := itemsDue[0]
	if dueItem.ID != createdItem.ID {
		t.Errorf("Expected item ID %d, got %d", createdItem.ID, dueItem.ID)
	}

	// Process the item (simulate what the scheduler does)
	todoText := createTodoTextForTest(dueItem)
	todoItem := models.TodoItem{
		Text:    todoText,
		Checked: false,
	}

	createdTodo := todoStore.CreateTodoItem(todoItem)
	if createdTodo.ID == 0 {
		t.Fatal("Failed to create todo item")
	}

	// Log execution
	logExecution := models.ExecutionLog{
		ScheduledItemID: dueItem.ID,
		ExecutedAt:      time.Now(),
		Status:          "success",
		TodoItemID:      &createdTodo.ID,
	}
	logStore.CreateExecutionLog(logExecution)

	// For one-time item, delete it after processing
	if !itemStore.DeleteScheduledItem(dueItem.ID) {
		t.Error("Failed to delete completed one-time item")
	}

	// Verify results
	finalTodos := todoStore.GetAllTodoItems()
	if len(finalTodos) != initialTodos+1 {
		t.Errorf("Expected %d todos, got %d", initialTodos+1, len(finalTodos))
	}

	finalLogs := logStore.GetAllExecutionLogs()
	if len(finalLogs) != initialLogs+1 {
		t.Errorf("Expected %d logs, got %d", initialLogs+1, len(finalLogs))
	}

	// Verify the scheduled item was deleted
	_, exists := itemStore.GetScheduledItem(createdItem.ID)
	if exists {
		t.Error("One-time item should have been deleted after processing")
	}

	// Verify todo content
	if createdTodo.Text != "One-time task - Test one-time scheduled item" {
		t.Errorf("Unexpected todo text: %s", createdTodo.Text)
	}
}

// testProcessRepeatingItem tests processing a repeating scheduled item
func testProcessRepeatingItem(t *testing.T, itemStore store.ScheduledItemStore, todoStore store.TodoItemStore, logStore store.ExecutionLogStore) {
	// Create a repeating item due for execution
	now := time.Now()
	pastTime := now.Add(-time.Hour) // 1 hour ago
	cronExpr := "0 * * * *"         // Every hour

	item := models.ScheduledItem{
		Title:           "Repeating task",
		Description:     "Test repeating scheduled item",
		StartsAt:        pastTime,
		Repeats:         true,
		CronExpression:  &cronExpr,
		NextExecutionAt: pastTime, // Due for execution
	}

	createdItem := itemStore.CreateScheduledItem(item)
	if createdItem.ID == 0 {
		t.Fatal("Failed to create scheduled item")
	}

	// Count initial state
	initialTodos := len(todoStore.GetAllTodoItems())
	initialLogs := len(logStore.GetAllExecutionLogs())

	// Get items due for execution
	itemsDue, err := itemStore.GetNextScheduledItems(10, 0)
	if err != nil {
		t.Fatalf("Failed to get items due: %v", err)
	}

	if len(itemsDue) == 0 {
		t.Fatal("Expected at least 1 item due")
	}

	var dueItem *models.ScheduledItem
	for _, item := range itemsDue {
		if item.ID == createdItem.ID {
			dueItem = &item
			break
		}
	}

	if dueItem == nil {
		t.Fatal("Created item not found in due items")
	}

	// Process the item
	todoText := createTodoTextForTest(*dueItem)
	todoItem := models.TodoItem{
		Text:    todoText,
		Checked: false,
	}

	createdTodo := todoStore.CreateTodoItem(todoItem)
	if createdTodo.ID == 0 {
		t.Fatal("Failed to create todo item")
	}

	// Log execution
	logExecution := models.ExecutionLog{
		ScheduledItemID: dueItem.ID,
		ExecutedAt:      time.Now(),
		Status:          "success",
		TodoItemID:      &createdTodo.ID,
	}
	logStore.CreateExecutionLog(logExecution)

	// For repeating item, update next execution time using the same logic as the scheduler
	nextExec := utils.CalculateNextExecution(dueItem.NextExecutionAt, dueItem.Repeats, dueItem.CronExpression, dueItem.Expiration)
	if nextExec == nil {
		t.Error("Failed to calculate next execution time")
		return
	}
	log.Printf("dueItem: %v\n", *dueItem)
	log.Printf("nextExec: %s\n", *nextExec)

	if !itemStore.UpdateNextExecutionAt(dueItem.ID, *nextExec) {
		t.Error("Failed to update next execution time")
	}

	// Verify results
	finalTodos := todoStore.GetAllTodoItems()
	if len(finalTodos) != initialTodos+1 {
		t.Errorf("Expected %d todos, got %d", initialTodos+1, len(finalTodos))
	}

	finalLogs := logStore.GetAllExecutionLogs()
	if len(finalLogs) != initialLogs+1 {
		t.Errorf("Expected %d logs, got %d", initialLogs+1, len(finalLogs))
	}

	// Verify the scheduled item still exists with updated next execution
	updatedItem, exists := itemStore.GetScheduledItem(createdItem.ID)
	if !exists {
		t.Error("Repeating item should still exist after processing")
	}
	log.Printf("updatedItem.NextExecutionAt: %v\n", updatedItem.NextExecutionAt)
	log.Printf("now: %v\n", now)
	log.Printf("updatedItem.NextExecutionAt.Before(now): %v\n", updatedItem.NextExecutionAt.Before(now))

	if updatedItem.NextExecutionAt.Before(now) {
		t.Error("Next execution time should be in the future")
	}
}

// testProcessExpiredItem tests processing an expired item
func testProcessExpiredItem(t *testing.T, itemStore store.ScheduledItemStore, todoStore store.TodoItemStore, logStore store.ExecutionLogStore) {
	// Create an expired repeating item
	now := time.Now()
	pastTime := now.Add(-2 * time.Hour) // 2 hours ago
	expiration := now.Add(-time.Hour)   // Expired 1 hour ago
	cronExpr := "0 * * * *"             // Every hour

	item := models.ScheduledItem{
		Title:           "Expired task",
		Description:     "Test expired scheduled item",
		StartsAt:        pastTime,
		Repeats:         true,
		CronExpression:  &cronExpr,
		Expiration:      &expiration,
		NextExecutionAt: pastTime, // Would be due, but item is expired
	}

	createdItem := itemStore.CreateScheduledItem(item)
	if createdItem.ID == 0 {
		t.Fatal("Failed to create scheduled item")
	}

	// Get items due for execution - expired item should NOT be included
	itemsDue, err := itemStore.GetNextScheduledItems(10, 0)
	if err != nil {
		t.Fatalf("Failed to get items due: %v", err)
	}

	// Verify expired item is not in the due list
	for _, item := range itemsDue {
		if item.ID == createdItem.ID {
			t.Error("Expired item should not be returned in due items")
		}
	}

	// Clean up
	itemStore.DeleteScheduledItem(createdItem.ID)
}

// testProcessMultipleItems tests processing multiple items in one run
func testProcessMultipleItems(t *testing.T, itemStore store.ScheduledItemStore, todoStore store.TodoItemStore, logStore store.ExecutionLogStore) {
	now := time.Now()
	pastTime := now.Add(-time.Hour)

	// Create multiple items due for execution
	items := []models.ScheduledItem{
		{
			Title:           "Task 1",
			Description:     "First task",
			StartsAt:        pastTime,
			Repeats:         false,
			NextExecutionAt: pastTime,
		},
		{
			Title:           "Task 2",
			Description:     "Second task",
			StartsAt:        pastTime,
			Repeats:         false,
			NextExecutionAt: pastTime,
		},
		{
			Title:           "Task 3",
			Description:     "Third task",
			StartsAt:        pastTime,
			Repeats:         false,
			NextExecutionAt: pastTime,
		},
	}

	var createdIDs []int64
	for _, item := range items {
		created := itemStore.CreateScheduledItem(item)
		if created.ID == 0 {
			t.Fatal("Failed to create scheduled item")
		}
		createdIDs = append(createdIDs, created.ID)
	}

	// Count initial state
	initialTodos := len(todoStore.GetAllTodoItems())
	initialLogs := len(logStore.GetAllExecutionLogs())

	// Get all items due for execution
	itemsDue, err := itemStore.GetNextScheduledItems(10, 0)
	if err != nil {
		t.Fatalf("Failed to get items due: %v", err)
	}

	// Should have at least our 3 items
	if len(itemsDue) < 3 {
		t.Fatalf("Expected at least 3 items due, got %d", len(itemsDue))
	}

	// Process each item
	var processedCount int
	for _, dueItem := range itemsDue {
		// Check if this is one of our test items
		isTestItem := false
		for _, id := range createdIDs {
			if dueItem.ID == id {
				isTestItem = true
				break
			}
		}

		if !isTestItem {
			continue // Skip items from other tests
		}

		// Create todo item
		todoText := createTodoTextForTest(dueItem)
		todoItem := models.TodoItem{
			Text:    todoText,
			Checked: false,
		}

		createdTodo := todoStore.CreateTodoItem(todoItem)
		if createdTodo.ID == 0 {
			t.Errorf("Failed to create todo item for scheduled item %d", dueItem.ID)
			continue
		}

		// Log execution
		logExecution := models.ExecutionLog{
			ScheduledItemID: dueItem.ID,
			ExecutedAt:      time.Now(),
			Status:          "success",
			TodoItemID:      &createdTodo.ID,
		}
		logStore.CreateExecutionLog(logExecution)

		// Delete one-time item
		itemStore.DeleteScheduledItem(dueItem.ID)
		processedCount++
	}

	if processedCount != 3 {
		t.Errorf("Expected to process 3 items, processed %d", processedCount)
	}

	// Verify results
	finalTodos := todoStore.GetAllTodoItems()
	finalLogs := logStore.GetAllExecutionLogs()

	if len(finalTodos) < initialTodos+3 {
		t.Errorf("Expected at least %d todos, got %d", initialTodos+3, len(finalTodos))
	}

	if len(finalLogs) < initialLogs+3 {
		t.Errorf("Expected at least %d logs, got %d", initialLogs+3, len(finalLogs))
	}
}

// createTodoTextForTest mimics the createTodoText function from the scheduler
func createTodoTextForTest(item models.ScheduledItem) string {
	if item.Description != "" {
		return item.Title + " - " + item.Description
	}
	return item.Title
}

// cleanupSchedulerTables cleans up test data
func cleanupSchedulerTables(t *testing.T, db interface{}) {
	// Note: In a real scenario, you might want to clean up test data
	// For now, we'll let the tables accumulate test data
	// This could be enhanced to clean up between tests if needed
}
