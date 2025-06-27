package store

import (
	"periodic-api/internal/models"
	"sort"
	"sync"
	"time"
)

// MemoryScheduledItemStore provides in-memory storage operations for scheduled items
type MemoryScheduledItemStore struct {
	sync.RWMutex
	items  map[int64]models.ScheduledItem
	nextID int64
}

// NewMemoryScheduledItemStore creates a new in-memory store
func NewMemoryScheduledItemStore() *MemoryScheduledItemStore {
	return &MemoryScheduledItemStore{
		items:  make(map[int64]models.ScheduledItem),
		nextID: 1,
	}
}

// CreateScheduledItem adds a new scheduled item to the in-memory store
func (s *MemoryScheduledItemStore) CreateScheduledItem(item models.ScheduledItem) models.ScheduledItem {
	s.Lock()
	defer s.Unlock()

	// Assign a new ID
	item.ID = s.nextID
	s.nextID++

	// Store the item
	s.items[item.ID] = item
	return item
}

// GetScheduledItem retrieves a scheduled item by ID from the in-memory store
func (s *MemoryScheduledItemStore) GetScheduledItem(id int64) (models.ScheduledItem, bool) {
	s.RLock()
	defer s.RUnlock()

	item, exists := s.items[id]
	return item, exists
}

// GetAllScheduledItems returns all scheduled items from the in-memory store
func (s *MemoryScheduledItemStore) GetAllScheduledItems() []models.ScheduledItem {
	s.RLock()
	defer s.RUnlock()

	items := make([]models.ScheduledItem, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items
}

// DeleteScheduledItem removes a scheduled item from the in-memory store
func (s *MemoryScheduledItemStore) DeleteScheduledItem(id int64) bool {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.items[id]; !exists {
		return false
	}

	delete(s.items, id)
	return true
}

// GetNextScheduledItems returns scheduled items ordered by next execution time with pagination
func (s *MemoryScheduledItemStore) GetNextScheduledItems(limit int, offset int64) ([]models.ScheduledItem, error) {
	s.RLock()
	defer s.RUnlock()

	// Filter items that have a next execution time
	var itemsWithNextExecution []models.ScheduledItem
	for _, item := range s.items {
		if item.NextExecutionAt != nil {
			itemsWithNextExecution = append(itemsWithNextExecution, item)
		}
	}

	// Sort by next execution time (earliest first)
	sort.Slice(itemsWithNextExecution, func(i, j int) bool {
		return itemsWithNextExecution[i].NextExecutionAt.Before(*itemsWithNextExecution[j].NextExecutionAt)
	})

	// Apply pagination
	startIndex := int(offset)
	endIndex := startIndex + limit

	if endIndex > len(itemsWithNextExecution) {
		endIndex = len(itemsWithNextExecution)
	}

	return itemsWithNextExecution[startIndex:endIndex], nil
}

// AddSampleData adds sample data to the in-memory store
func (s *MemoryScheduledItemStore) AddSampleData() {
	// Only add sample data if the store is empty
	if len(s.GetAllScheduledItems()) == 0 {
		startsAt1, _ := time.Parse(time.RFC3339, "2023-05-15T10:00:00Z")
		s.CreateScheduledItem(models.ScheduledItem{
			Title:       "Sample Scheduled Item 1",
			Description: "Description for item 1",
			StartsAt:    startsAt1,
			Repeats:     false,
		})

		cronExpr := "0 9 * * MON-FRI"
		startsAt2, _ := time.Parse(time.RFC3339, "2023-05-16T14:30:00Z")
		expirationTime, _ := time.Parse(time.RFC3339, "2023-12-31T23:59:59Z")
		s.CreateScheduledItem(models.ScheduledItem{
			Title:          "Sample Scheduled Item 2",
			Description:    "Description for item 2",
			StartsAt:       startsAt2,
			Repeats:        true,
			CronExpression: &cronExpr,
			Expiration:     &expirationTime,
		})
	}
}
