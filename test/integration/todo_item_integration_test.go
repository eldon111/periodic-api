package integration

import (
	"awesomeProject/internal/models"
	"awesomeProject/internal/store"
	"testing"
)

func TestTodoItemIntegration(t *testing.T) {
	skipIfDBNotAvailable(t)

	// Clean up before and after the test
	cleanupTodoItems(t)
	defer cleanupTodoItems(t)

	// Create store
	todoStore := store.NewPostgresTodoItemStore(getActiveDB())

	// Test data
	testItem := models.TodoItem{
		Text:    "Integration Test Todo",
		Checked: false,
	}

	t.Run("Full CRUD Workflow", func(t *testing.T) {
		// Create
		created := todoStore.CreateTodoItem(testItem)
		if created.ID == 0 {
			t.Fatal("Created item should have non-zero ID")
		}
		if created.Text != testItem.Text {
			t.Errorf("Expected text %s, got %s", testItem.Text, created.Text)
		}
		if created.Checked != testItem.Checked {
			t.Errorf("Expected checked %v, got %v", testItem.Checked, created.Checked)
		}

		// Read
		retrieved, found := todoStore.GetTodoItem(created.ID)
		if !found {
			t.Fatal("Should find the created item")
		}
		if retrieved.Text != testItem.Text {
			t.Errorf("Expected text %s, got %s", testItem.Text, retrieved.Text)
		}
		if retrieved.Checked != testItem.Checked {
			t.Errorf("Expected checked %v, got %v", testItem.Checked, retrieved.Checked)
		}

		// Update
		updated := models.TodoItem{
			Text:    "Updated Integration Test Todo",
			Checked: true,
		}

		result, success := todoStore.UpdateTodoItem(created.ID, updated)
		if !success {
			t.Fatal("Update should succeed")
		}
		if result.Text != updated.Text {
			t.Errorf("Expected updated text %s, got %s", updated.Text, result.Text)
		}
		if result.Checked != updated.Checked {
			t.Errorf("Expected updated checked %v, got %v", updated.Checked, result.Checked)
		}

		// Verify update persisted
		verified, found := todoStore.GetTodoItem(created.ID)
		if !found {
			t.Fatal("Should still find the item after update")
		}
		if verified.Text != updated.Text {
			t.Errorf("Update should persist: expected text %s, got %s", updated.Text, verified.Text)
		}
		if verified.Checked != updated.Checked {
			t.Errorf("Update should persist: expected checked %v, got %v", updated.Checked, verified.Checked)
		}

		// Delete
		deleted := todoStore.DeleteTodoItem(created.ID)
		if !deleted {
			t.Fatal("Delete should succeed")
		}

		// Verify deletion
		_, found = todoStore.GetTodoItem(created.ID)
		if found {
			t.Error("Item should not be found after deletion")
		}
	})

	t.Run("Multiple Items Operations", func(t *testing.T) {
		// Create multiple items
		items := []models.TodoItem{
			{Text: "Todo 1", Checked: false},
			{Text: "Todo 2", Checked: true},
			{Text: "Todo 3", Checked: false},
		}

		var createdIDs []int64
		for _, item := range items {
			created := todoStore.CreateTodoItem(item)
			createdIDs = append(createdIDs, created.ID)
		}

		// Get all items
		allItems := todoStore.GetAllTodoItems()
		if len(allItems) < 3 {
			t.Errorf("Expected at least 3 items, got %d", len(allItems))
		}

		// Verify our items are in the list
		foundCount := 0
		for _, item := range allItems {
			for _, id := range createdIDs {
				if item.ID == id {
					foundCount++
					break
				}
			}
		}
		if foundCount != 3 {
			t.Errorf("Expected to find 3 created items, found %d", foundCount)
		}

		// Test filtering by checked status
		checkedCount := 0
		uncheckedCount := 0
		for _, item := range allItems {
			// Only count our test items
			isOurItem := false
			for _, id := range createdIDs {
				if item.ID == id {
					isOurItem = true
					break
				}
			}
			if isOurItem {
				if item.Checked {
					checkedCount++
				} else {
					uncheckedCount++
				}
			}
		}
		if checkedCount != 1 {
			t.Errorf("Expected 1 checked item, got %d", checkedCount)
		}
		if uncheckedCount != 2 {
			t.Errorf("Expected 2 unchecked items, got %d", uncheckedCount)
		}

		// Clean up
		for _, id := range createdIDs {
			todoStore.DeleteTodoItem(id)
		}
	})

	t.Run("Toggle Checked Status", func(t *testing.T) {
		// Create item
		item := models.TodoItem{
			Text:    "Toggle Test Item",
			Checked: false,
		}
		created := todoStore.CreateTodoItem(item)

		// Toggle to checked
		updated := models.TodoItem{
			Text:    created.Text,
			Checked: true,
		}
		result, success := todoStore.UpdateTodoItem(created.ID, updated)
		if !success {
			t.Fatal("Should be able to update checked status")
		}
		if !result.Checked {
			t.Error("Item should be checked after update")
		}

		// Toggle back to unchecked
		updated.Checked = false
		result, success = todoStore.UpdateTodoItem(created.ID, updated)
		if !success {
			t.Fatal("Should be able to update checked status again")
		}
		if result.Checked {
			t.Error("Item should be unchecked after second update")
		}

		// Clean up
		todoStore.DeleteTodoItem(created.ID)
	})

	t.Run("Edge Cases", func(t *testing.T) {
		// Test with non-existent ID
		_, found := todoStore.GetTodoItem(99999)
		if found {
			t.Error("Should not find non-existent item")
		}

		// Test update non-existent item
		_, success := todoStore.UpdateTodoItem(99999, testItem)
		if success {
			t.Error("Update of non-existent item should fail")
		}

		// Test delete non-existent item
		deleted := todoStore.DeleteTodoItem(99999)
		if deleted {
			t.Error("Delete of non-existent item should fail")
		}

		// Test with empty text (should still work)
		emptyItem := models.TodoItem{
			Text:    "",
			Checked: false,
		}
		created := todoStore.CreateTodoItem(emptyItem)
		if created.ID == 0 {
			t.Error("Should be able to create item with empty text")
		}
		todoStore.DeleteTodoItem(created.ID)
	})
}

func cleanupTodoItems(t *testing.T) {
	_, err := getActiveDB().Exec("DELETE FROM todo_items")
	if err != nil {
		t.Logf("Failed to cleanup todo_items: %v", err)
	}
}
