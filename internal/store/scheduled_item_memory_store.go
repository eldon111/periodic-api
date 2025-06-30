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

// UpdateNextExecutionAt updates the next execution time for a scheduled item
func (s *MemoryScheduledItemStore) UpdateNextExecutionAt(id int64, nextExecutionAt time.Time) bool {
	s.Lock()
	defer s.Unlock()

	item, exists := s.items[id]
	if !exists {
		return false
	}

	item.NextExecutionAt = nextExecutionAt
	s.items[id] = item
	return true
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

	now := time.Now()
	
	// Filter items that are due for execution and not expired
	var itemsDue []models.ScheduledItem
	for _, item := range s.items {
		// Skip items that are not yet due
		if item.NextExecutionAt.After(now) {
			continue
		}
		
		// Skip expired items
		if item.Expiration != nil && now.After(*item.Expiration) {
			continue
		}
		
		itemsDue = append(itemsDue, item)
	}

	// Sort by next execution time (earliest first)
	sort.Slice(itemsDue, func(i, j int) bool {
		return itemsDue[i].NextExecutionAt.Before(itemsDue[j].NextExecutionAt)
	})

	// Apply pagination
	startIndex := int(offset)
	endIndex := startIndex + limit

	if endIndex > len(itemsDue) {
		endIndex = len(itemsDue)
	}

	return itemsDue[startIndex:endIndex], nil
}

// AddSampleData adds sample data to the in-memory store
func (s *MemoryScheduledItemStore) AddSampleData() {
	// Only add sample data if the store is empty
	if len(s.GetAllScheduledItems()) == 0 {
		startsAt1, _ := time.Parse(time.RFC3339, "2023-05-15T10:00:00Z")
		s.CreateScheduledItem(models.ScheduledItem{
			Title:           "Sample Scheduled Item 1",
			Description:     "Description for item 1",
			StartsAt:        startsAt1,
			Repeats:         false,
			NextExecutionAt: startsAt1, // One-time item, set to start time
		})

		cronExpr := "0 9 * * MON-FRI"
		startsAt2, _ := time.Parse(time.RFC3339, "2023-05-16T14:30:00Z")
		expirationTime, _ := time.Parse(time.RFC3339, "2023-12-31T23:59:59Z")
		nextExec2 := time.Now().Add(time.Hour) // Set to 1 hour from now for testing
		s.CreateScheduledItem(models.ScheduledItem{
			Title:           "Sample Scheduled Item 2",
			Description:     "Description for item 2",
			StartsAt:        startsAt2,
			Repeats:         true,
			CronExpression:  &cronExpr,
			Expiration:      &expirationTime,
			NextExecutionAt: nextExec2,
		})
	}
}
