package integration

import (
	"periodic-api/internal/models"
	"periodic-api/internal/store"
	"testing"
	"time"
)

func TestScheduledItemIntegration(t *testing.T) {
	skipIfDBNotAvailable(t)

	// Clean up before and after the test
	cleanupScheduledItems(t)
	defer cleanupScheduledItems(t)

	// Create store
	scheduleStore := store.NewPostgresScheduledItemStore(getActiveDB())

	// Test data
	now := time.Now()
	cronExpr := "0 0 * * *"
	expiration := now.AddDate(0, 1, 0)

	testItem := models.ScheduledItem{
		Title:          "Integration Test Item",
		Description:    "This is an integration test scheduled item",
		StartsAt:       now,
		Repeats:        true,
		CronExpression: &cronExpr,
		Expiration:     &expiration,
	}

	t.Run("Full CRUD Workflow", func(t *testing.T) {
		// Create
		created := scheduleStore.CreateScheduledItem(testItem)
		if created.ID == 0 {
			t.Fatal("Created item should have non-zero ID")
		}

		// Read
		retrieved, found := scheduleStore.GetScheduledItem(created.ID)
		if !found {
			t.Fatal("Should find the created item")
		}
		if retrieved.Title != testItem.Title {
			t.Errorf("Expected title %s, got %s", testItem.Title, retrieved.Title)
		}

		// Delete
		deleted := scheduleStore.DeleteScheduledItem(created.ID)
		if !deleted {
			t.Fatal("Delete should succeed")
		}

		// Verify deletion
		_, found = scheduleStore.GetScheduledItem(created.ID)
		if found {
			t.Error("Item should not be found after deletion")
		}
	})

	t.Run("Multiple Items Operations", func(t *testing.T) {
		// Create multiple items
		items := []models.ScheduledItem{
			{
				Title:       "Item 1",
				Description: "First item",
				StartsAt:    now,
				Repeats:     false,
			},
			{
				Title:          "Item 2",
				Description:    "Second item",
				StartsAt:       now.Add(time.Hour),
				Repeats:        true,
				CronExpression: &cronExpr,
			},
		}

		var createdIDs []int64
		for _, item := range items {
			created := scheduleStore.CreateScheduledItem(item)
			createdIDs = append(createdIDs, created.ID)
		}

		// Get all items
		allItems := scheduleStore.GetAllScheduledItems()
		if len(allItems) < 2 {
			t.Errorf("Expected at least 2 items, got %d", len(allItems))
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
		if foundCount != 2 {
			t.Errorf("Expected to find 2 created items, found %d", foundCount)
		}

		// Clean up
		for _, id := range createdIDs {
			scheduleStore.DeleteScheduledItem(id)
		}
	})

	t.Run("Edge Cases", func(t *testing.T) {
		// Test with non-existent ID
		_, found := scheduleStore.GetScheduledItem(99999)
		if found {
			t.Error("Should not find non-existent item")
		}

		// Test delete non-existent item
		deleted := scheduleStore.DeleteScheduledItem(99999)
		if deleted {
			t.Error("Delete of non-existent item should fail")
		}
	})
}

func cleanupScheduledItems(t *testing.T) {
	_, err := getActiveDB().Exec("DELETE FROM scheduled_items")
	if err != nil {
		t.Logf("Failed to cleanup scheduled_items: %v", err)
	}
}
