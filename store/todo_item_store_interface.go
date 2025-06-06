package store

import (
	"awesomeProject/models"
)

// TodoItemStore defines the interface for todo item storage operations
type TodoItemStore interface {
	CreateTodoItem(item models.TodoItem) models.TodoItem
	GetTodoItem(id int64) (models.TodoItem, bool)
	GetAllTodoItems() []models.TodoItem
	UpdateTodoItem(id int64, updatedItem models.TodoItem) (models.TodoItem, bool)
	DeleteTodoItem(id int64) bool
	AddSampleData()
}