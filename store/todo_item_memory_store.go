package store

import (
	"awesomeProject/models"
	"sync"
)

// MemoryTodoItemStore provides in-memory storage operations for todo items
type MemoryTodoItemStore struct {
	sync.RWMutex
	items  map[int64]models.TodoItem
	nextID int64
}

// NewMemoryTodoItemStore creates a new in-memory store
func NewMemoryTodoItemStore() *MemoryTodoItemStore {
	return &MemoryTodoItemStore{
		items:  make(map[int64]models.TodoItem),
		nextID: 1,
	}
}

// CreateTodoItem adds a new todo item to the in-memory store
func (s *MemoryTodoItemStore) CreateTodoItem(item models.TodoItem) models.TodoItem {
	s.Lock()
	defer s.Unlock()

	// Assign a new ID
	item.ID = s.nextID
	s.nextID++

	// Store the item
	s.items[item.ID] = item
	return item
}

// GetTodoItem retrieves a todo item by ID from the in-memory store
func (s *MemoryTodoItemStore) GetTodoItem(id int64) (models.TodoItem, bool) {
	s.RLock()
	defer s.RUnlock()

	item, exists := s.items[id]
	return item, exists
}

// GetAllTodoItems returns all todo items from the in-memory store
func (s *MemoryTodoItemStore) GetAllTodoItems() []models.TodoItem {
	s.RLock()
	defer s.RUnlock()

	items := make([]models.TodoItem, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items
}

// UpdateTodoItem updates an existing todo item in the in-memory store
func (s *MemoryTodoItemStore) UpdateTodoItem(id int64, updatedItem models.TodoItem) (models.TodoItem, bool) {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.items[id]; !exists {
		return models.TodoItem{}, false
	}

	updatedItem.ID = id
	s.items[id] = updatedItem
	return updatedItem, true
}

// DeleteTodoItem removes a todo item from the in-memory store
func (s *MemoryTodoItemStore) DeleteTodoItem(id int64) bool {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.items[id]; !exists {
		return false
	}

	delete(s.items, id)
	return true
}

// AddSampleData adds sample data to the in-memory store
func (s *MemoryTodoItemStore) AddSampleData() {
	// Only add sample data if the store is empty
	if len(s.GetAllTodoItems()) == 0 {
		s.CreateTodoItem(models.TodoItem{
			Text:    "Buy groceries",
			Checked: false,
		})

		s.CreateTodoItem(models.TodoItem{
			Text:    "Clean the house",
			Checked: true,
		})

		s.CreateTodoItem(models.TodoItem{
			Text:    "Finish project",
			Checked: false,
		})
	}
}