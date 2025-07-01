package models

import (
	"time"
)

// ExecutionLog represents a log entry for scheduled item processing
type ExecutionLog struct {
	ID              int64      `json:"id"`
	ScheduledItemID int64      `json:"scheduledItemId"`
	ExecutedAt      time.Time  `json:"executedAt"`
	Status          string     `json:"status"`
	ErrorMessage    *string    `json:"errorMessage,omitempty"`
	TodoItemID      *int64     `json:"todoItemId,omitempty"`
}