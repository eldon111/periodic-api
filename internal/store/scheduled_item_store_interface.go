package store

import (
	"awesomeProject/internal/models"
)

// ScheduledItemStore defines the interface for scheduled item storage operations
type ScheduledItemStore interface {
	CreateScheduledItem(item models.ScheduledItem) models.ScheduledItem
	GetScheduledItem(id int64) (models.ScheduledItem, bool)
	GetAllScheduledItems() []models.ScheduledItem
	GetNextScheduledItems(limit int, offset int64) ([]models.ScheduledItem, error)
	UpdateScheduledItem(id int64, updatedItem models.ScheduledItem) (models.ScheduledItem, bool)
	DeleteScheduledItem(id int64) bool
	AddSampleData()
}
