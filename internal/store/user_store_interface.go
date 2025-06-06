package store

import (
	"awesomeProject/internal/models"
)

// UserStore defines the interface for user storage operations
type UserStore interface {
	CreateUser(user models.User) models.User
	GetUser(id int64) (models.User, bool)
	GetAllUsers() []models.User
	UpdateUser(id int64, updatedUser models.User) (models.User, bool)
	DeleteUser(id int64) bool
	AddSampleData()
}
