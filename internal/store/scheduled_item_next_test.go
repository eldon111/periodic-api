package store

import (
	"awesomeProject/internal/models"
	"testing"
	"time"
)

func TestMemoryStoreGetNextScheduledItems(t *testing.T) {
	store := NewMemoryScheduledItemStore()
	now := time.Now()

	// Create test items with different next execution times
	item1 := models.ScheduledItem{
		Title:           "Future Item 1",
		Description:     "Description 1",
		StartsAt:        now.Add(1 * time.Hour),
		Repeats:         false,
		NextExecutionAt: timePtr(now.Add(1 * time.Hour)),
	}

	item2 := models.ScheduledItem{
		Title:           "Future Item 2",
		Description:     "Description 2",
		StartsAt:        now.Add(2 * time.Hour),
		Repeats:         false,
		NextExecutionAt: timePtr(now.Add(2 * time.Hour)),
	}

	item3 := models.ScheduledItem{
		Title:           "Past Item",
		Description:     "Description 3",
		StartsAt:        now.Add(-1 * time.Hour),
		Repeats:         false,
		NextExecutionAt: nil, // No next execution
	}

	item4 := models.ScheduledItem{
		Title:           "Earliest Item",
		Description:     "Description 4",
		StartsAt:        now.Add(30 * time.Minute),
		Repeats:         false,
		NextExecutionAt: timePtr(now.Add(30 * time.Minute)),
	}

	// Add items to store
	store.CreateScheduledItem(item1)
	store.CreateScheduledItem(item2)
	store.CreateScheduledItem(item3)
	store.CreateScheduledItem(item4)

	// Test GetNextScheduledItems
	nextItems, err := store.GetNextScheduledItems(5, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return 3 items (excluding the one with nil NextExecutionAt)
	if len(nextItems) != 3 {
		t.Errorf("Expected 3 items but got %d", len(nextItems))
	}

	// Should be sorted by next execution time (earliest first)
	if len(nextItems) > 0 && nextItems[0].Title != "Earliest Item" {
		t.Errorf("Expected first item to be 'Earliest Item' but got '%s'", nextItems[0].Title)
	}

	// Test with limit
	limitedItems, err := store.GetNextScheduledItems(2, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(limitedItems) != 2 {
		t.Errorf("Expected 2 items with limit but got %d", len(limitedItems))
	}

	// Test with limit larger than available items
	allItems, err := store.GetNextScheduledItems(10, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(allItems) != 3 {
		t.Errorf("Expected 3 items with large limit but got %d", len(allItems))
	}
}

func TestMemoryStoreNextExecutionCalculation(t *testing.T) {
	store := NewMemoryScheduledItemStore()
	now := time.Now()

	// Test non-repeating future item
	futureItem := models.ScheduledItem{
		Title:       "Future Item",
		Description: "Future description",
		StartsAt:    now.Add(1 * time.Hour),
		Repeats:     false,
	}

	created := store.CreateScheduledItem(futureItem)
	if created.NextExecutionAt == nil {
		t.Error("Expected NextExecutionAt to be set for future non-repeating item")
	} else if created.NextExecutionAt.Unix() != futureItem.StartsAt.Unix() {
		t.Errorf("Expected NextExecutionAt to equal StartsAt for non-repeating item")
	}

	// Test non-repeating past item
	pastItem := models.ScheduledItem{
		Title:       "Past Item",
		Description: "Past description",
		StartsAt:    now.Add(-1 * time.Hour),
		Repeats:     false,
	}

	createdPast := store.CreateScheduledItem(pastItem)
	if createdPast.NextExecutionAt != nil {
		t.Error("Expected NextExecutionAt to be nil for past non-repeating item")
	}

	// Test repeating item with cron expression
	cronExpr := "0 9 * * MON-FRI" // 9 AM weekdays
	repeatingItem := models.ScheduledItem{
		Title:          "Repeating Item",
		Description:    "Repeating description",
		StartsAt:       now.Add(-1 * time.Hour), // Started in the past
		Repeats:        true,
		CronExpression: &cronExpr,
	}

	createdRepeating := store.CreateScheduledItem(repeatingItem)
	if createdRepeating.NextExecutionAt == nil {
		t.Error("Expected NextExecutionAt to be set for repeating item")
	} else if !createdRepeating.NextExecutionAt.After(now) {
		t.Error("Expected NextExecutionAt to be in the future for repeating item")
	}
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}
