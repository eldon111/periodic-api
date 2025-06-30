package models

import (
	"time"
)

// ScheduledItem represents the data model for our CRUD operations
type ScheduledItem struct {
	ID              int64      `json:"id" example:"1"`
	Title           string     `json:"title" example:"Daily standup meeting"`
	Description     string     `json:"description" example:"Team daily standup meeting to discuss progress"`
	StartsAt        time.Time  `json:"startsAt" example:"2024-01-01T09:00:00Z"`
	Repeats         bool       `json:"repeats" example:"true"`
	CronExpression  *string    `json:"cronExpression,omitempty" example:"0 9 * * 1-5"`
	Expiration      *time.Time `json:"expiration,omitempty" example:"2024-12-31T23:59:59Z"`
	NextExecutionAt *time.Time `json:"nextExecutionAt,omitempty" example:"2024-01-02T09:00:00Z"`
}
