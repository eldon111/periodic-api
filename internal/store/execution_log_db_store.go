package store

import (
	"database/sql"
	"log"
	"periodic-api/internal/models"
	"sync"
	"time"
)

// PostgresExecutionLogStore provides PostgreSQL storage operations for execution logs
type PostgresExecutionLogStore struct {
	sync.RWMutex
	db *sql.DB
}

// NewPostgresExecutionLogStore creates a new PostgreSQL execution log store with the given database connection
func NewPostgresExecutionLogStore(db *sql.DB) *PostgresExecutionLogStore {
	return &PostgresExecutionLogStore{
		db: db,
	}
}

// CreateExecutionLog adds a new execution log to the database
func (s *PostgresExecutionLogStore) CreateExecutionLog(logEntry models.ExecutionLog) models.ExecutionLog {
	s.Lock()
	defer s.Unlock()

	// Set executed time if not provided
	if logEntry.ExecutedAt.IsZero() {
		logEntry.ExecutedAt = time.Now()
	}

	query := `
		INSERT INTO execution_logs 
		(scheduled_item_id, executed_at, status, error_message, todo_item_id) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`

	err := s.db.QueryRow(
		query,
		logEntry.ScheduledItemID,
		logEntry.ExecutedAt,
		logEntry.Status,
		logEntry.ErrorMessage,
		logEntry.TodoItemID,
	).Scan(&logEntry.ID)

	if err != nil {
		log.Printf("Error creating execution log: %v", err)
		return models.ExecutionLog{} // Return empty log on error
	}

	return logEntry
}

// GetExecutionLog retrieves an execution log by ID from the database
func (s *PostgresExecutionLogStore) GetExecutionLog(id int64) (models.ExecutionLog, bool) {
	s.RLock()
	defer s.RUnlock()

	var logEntry models.ExecutionLog
	query := `
		SELECT id, scheduled_item_id, executed_at, status, error_message, todo_item_id 
		FROM execution_logs 
		WHERE id = $1
	`

	err := s.db.QueryRow(query, id).Scan(
		&logEntry.ID,
		&logEntry.ScheduledItemID,
		&logEntry.ExecutedAt,
		&logEntry.Status,
		&logEntry.ErrorMessage,
		&logEntry.TodoItemID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.ExecutionLog{}, false
		}
		log.Printf("Error getting execution log: %v", err)
		return models.ExecutionLog{}, false
	}

	return logEntry, true
}

// GetAllExecutionLogs returns all execution logs from the database
func (s *PostgresExecutionLogStore) GetAllExecutionLogs() []models.ExecutionLog {
	s.RLock()
	defer s.RUnlock()

	query := `
		SELECT id, scheduled_item_id, executed_at, status, error_message, todo_item_id 
		FROM execution_logs
		ORDER BY executed_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error querying execution logs: %v", err)
		return []models.ExecutionLog{}
	}
	defer rows.Close()

	var logs []models.ExecutionLog
	for rows.Next() {
		var logEntry models.ExecutionLog

		err := rows.Scan(
			&logEntry.ID,
			&logEntry.ScheduledItemID,
			&logEntry.ExecutedAt,
			&logEntry.Status,
			&logEntry.ErrorMessage,
			&logEntry.TodoItemID,
		)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		logs = append(logs, logEntry)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}

	return logs
}

// GetExecutionLogsByScheduledItemID returns all execution logs for a specific scheduled item
func (s *PostgresExecutionLogStore) GetExecutionLogsByScheduledItemID(scheduledItemID int64) []models.ExecutionLog {
	s.RLock()
	defer s.RUnlock()

	query := `
		SELECT id, scheduled_item_id, executed_at, status, error_message, todo_item_id 
		FROM execution_logs
		WHERE scheduled_item_id = $1
		ORDER BY executed_at DESC
	`

	rows, err := s.db.Query(query, scheduledItemID)
	if err != nil {
		log.Printf("Error querying execution logs by scheduled item ID: %v", err)
		return []models.ExecutionLog{}
	}
	defer rows.Close()

	var logs []models.ExecutionLog
	for rows.Next() {
		var logEntry models.ExecutionLog

		err := rows.Scan(
			&logEntry.ID,
			&logEntry.ScheduledItemID,
			&logEntry.ExecutedAt,
			&logEntry.Status,
			&logEntry.ErrorMessage,
			&logEntry.TodoItemID,
		)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		logs = append(logs, logEntry)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}

	return logs
}