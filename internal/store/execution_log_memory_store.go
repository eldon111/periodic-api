package store

import (
	"periodic-api/internal/models"
	"sync"
	"time"
)

// MemoryExecutionLogStore provides in-memory storage operations for execution logs
type MemoryExecutionLogStore struct {
	sync.RWMutex
	logs   map[int64]models.ExecutionLog
	nextID int64
}

// NewMemoryExecutionLogStore creates a new in-memory execution log store
func NewMemoryExecutionLogStore() *MemoryExecutionLogStore {
	return &MemoryExecutionLogStore{
		logs:   make(map[int64]models.ExecutionLog),
		nextID: 1,
	}
}

// CreateExecutionLog adds a new execution log to the in-memory store
func (s *MemoryExecutionLogStore) CreateExecutionLog(log models.ExecutionLog) models.ExecutionLog {
	s.Lock()
	defer s.Unlock()

	// Assign a new ID and set executed time if not provided
	log.ID = s.nextID
	s.nextID++
	
	if log.ExecutedAt.IsZero() {
		log.ExecutedAt = time.Now()
	}

	// Store the log
	s.logs[log.ID] = log
	return log
}

// GetExecutionLog retrieves an execution log by ID from the in-memory store
func (s *MemoryExecutionLogStore) GetExecutionLog(id int64) (models.ExecutionLog, bool) {
	s.RLock()
	defer s.RUnlock()

	log, exists := s.logs[id]
	return log, exists
}

// GetAllExecutionLogs returns all execution logs from the in-memory store
func (s *MemoryExecutionLogStore) GetAllExecutionLogs() []models.ExecutionLog {
	s.RLock()
	defer s.RUnlock()

	logs := make([]models.ExecutionLog, 0, len(s.logs))
	for _, log := range s.logs {
		logs = append(logs, log)
	}
	return logs
}

// GetExecutionLogsByScheduledItemID returns all execution logs for a specific scheduled item
func (s *MemoryExecutionLogStore) GetExecutionLogsByScheduledItemID(scheduledItemID int64) []models.ExecutionLog {
	s.RLock()
	defer s.RUnlock()

	var logs []models.ExecutionLog
	for _, log := range s.logs {
		if log.ScheduledItemID == scheduledItemID {
			logs = append(logs, log)
		}
	}
	return logs
}