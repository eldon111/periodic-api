package models

import (
	"time"
)

// ScheduledItem represents the data model for our CRUD operations
type ScheduledItem struct {
	ID             int64      `json:"id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	StartsAt       time.Time  `json:"startsAt"`
	Repeats        bool       `json:"repeats"`
	CronExpression *string    `json:"cronExpression,omitempty"`
	Expiration     *time.Time `json:"expiration,omitempty"`
}