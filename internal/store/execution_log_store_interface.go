package store

import (
	"periodic-api/internal/models"
)

// ExecutionLogStore defines the interface for execution log storage operations
type ExecutionLogStore interface {
	CreateExecutionLog(log models.ExecutionLog) models.ExecutionLog
	GetExecutionLog(id int64) (models.ExecutionLog, bool)
	GetAllExecutionLogs() []models.ExecutionLog
	GetExecutionLogsByScheduledItemID(scheduledItemID int64) []models.ExecutionLog
}