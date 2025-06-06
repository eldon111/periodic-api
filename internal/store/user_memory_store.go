package store

import (
	"awesomeProject/internal/models"
	"sync"
)

// MemoryUserStore provides in-memory storage operations for users
type MemoryUserStore struct {
	sync.RWMutex
	users  map[int64]models.User
	nextID int64
}

// NewMemoryUserStore creates a new in-memory store
func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users:  make(map[int64]models.User),
		nextID: 1,
	}
}

// CreateUser adds a new user to the in-memory store
func (s *MemoryUserStore) CreateUser(user models.User) models.User {
	s.Lock()
	defer s.Unlock()

	// Assign a new ID
	user.ID = s.nextID
	s.nextID++

	// Store the user
	s.users[user.ID] = user
	return user
}

// GetUser retrieves a user by ID from the in-memory store
func (s *MemoryUserStore) GetUser(id int64) (models.User, bool) {
	s.RLock()
	defer s.RUnlock()

	user, exists := s.users[id]
	return user, exists
}

// GetAllUsers returns all users from the in-memory store
func (s *MemoryUserStore) GetAllUsers() []models.User {
	s.RLock()
	defer s.RUnlock()

	users := make([]models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

// UpdateUser updates an existing user in the in-memory store
func (s *MemoryUserStore) UpdateUser(id int64, updatedUser models.User) (models.User, bool) {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.users[id]; !exists {
		return models.User{}, false
	}

	updatedUser.ID = id
	s.users[id] = updatedUser
	return updatedUser, true
}

// DeleteUser removes a user from the in-memory store
func (s *MemoryUserStore) DeleteUser(id int64) bool {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.users[id]; !exists {
		return false
	}

	delete(s.users, id)
	return true
}

// AddSampleData adds sample data to the in-memory store
func (s *MemoryUserStore) AddSampleData() {
	// Only add sample data if the store is empty
	if len(s.GetAllUsers()) == 0 {
		s.CreateUser(models.User{
			Username: "admin",
			PasswordHash: []byte("admin123"),
		})

		s.CreateUser(models.User{
			Username: "user1",
			PasswordHash: []byte("password123"),
		})
	}
}
