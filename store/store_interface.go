package store

import (
	"awesomeProject/models"
)

// ScheduledItemStorer defines the interface for scheduled item storage operations
type ScheduledItemStorer interface {
	CreateScheduledItem(item models.ScheduledItem) models.ScheduledItem
	GetScheduledItem(id int64) (models.ScheduledItem, bool)
	GetAllScheduledItems() []models.ScheduledItem
	UpdateScheduledItem(id int64, updatedItem models.ScheduledItem) (models.ScheduledItem, bool)
	DeleteScheduledItem(id int64) bool
	AddSampleData()
}
