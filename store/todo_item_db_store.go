package store

import (
	"awesomeProject/models"
	"database/sql"
	"log"
	"sync"
)

// PostgresTodoItemStore provides PostgreSQL storage operations for todo items
type PostgresTodoItemStore struct {
	sync.RWMutex
	db *sql.DB
}

// NewPostgresTodoItemStore creates a new PostgreSQL store with the given database connection
func NewPostgresTodoItemStore(db *sql.DB) *PostgresTodoItemStore {
	return &PostgresTodoItemStore{
		db: db,
	}
}

// CreateTodoItem adds a new todo item to the database
func (s *PostgresTodoItemStore) CreateTodoItem(item models.TodoItem) models.TodoItem {
	s.Lock()
	defer s.Unlock()

	query := `
		INSERT INTO todo_items 
		(text, checked) 
		VALUES ($1, $2) 
		RETURNING id
	`

	err := s.db.QueryRow(
		query,
		item.Text,
		item.Checked,
	).Scan(&item.ID)

	if err != nil {
		log.Printf("Error creating todo item: %v", err)
		return models.TodoItem{} // Return empty item on error
	}

	return item
}

// GetTodoItem retrieves a todo item by ID from the database
func (s *PostgresTodoItemStore) GetTodoItem(id int64) (models.TodoItem, bool) {
	s.RLock()
	defer s.RUnlock()

	var item models.TodoItem
	query := `
		SELECT id, text, checked 
		FROM todo_items 
		WHERE id = $1
	`

	err := s.db.QueryRow(query, id).Scan(
		&item.ID,
		&item.Text,
		&item.Checked,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.TodoItem{}, false
		}
		log.Printf("Error getting todo item: %v", err)
		return models.TodoItem{}, false
	}

	return item, true
}

// GetAllTodoItems returns all todo items from the database
func (s *PostgresTodoItemStore) GetAllTodoItems() []models.TodoItem {
	s.RLock()
	defer s.RUnlock()

	query := `
		SELECT id, text, checked 
		FROM todo_items
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error querying todo items: %v", err)
		return []models.TodoItem{}
	}
	defer rows.Close()

	var items []models.TodoItem
	for rows.Next() {
		var item models.TodoItem

		err := rows.Scan(
			&item.ID,
			&item.Text,
			&item.Checked,
		)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}

	return items
}

// UpdateTodoItem updates an existing todo item in the database
func (s *PostgresTodoItemStore) UpdateTodoItem(id int64, updatedItem models.TodoItem) (models.TodoItem, bool) {
	s.Lock()
	defer s.Unlock()

	query := `
		UPDATE todo_items 
		SET text = $1, checked = $2 
		WHERE id = $3
	`

	result, err := s.db.Exec(
		query,
		updatedItem.Text,
		updatedItem.Checked,
		id,
	)

	if err != nil {
		log.Printf("Error updating todo item: %v", err)
		return models.TodoItem{}, false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return models.TodoItem{}, false
	}

	if rowsAffected == 0 {
		return models.TodoItem{}, false
	}

	updatedItem.ID = id
	return updatedItem, true
}

// DeleteTodoItem removes a todo item from the database
func (s *PostgresTodoItemStore) DeleteTodoItem(id int64) bool {
	s.Lock()
	defer s.Unlock()

	query := `DELETE FROM todo_items WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting todo item: %v", err)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return false
	}

	return rowsAffected > 0
}

// AddSampleData adds sample data to the database if it's empty
func (s *PostgresTodoItemStore) AddSampleData() {
	count := 0
	err := s.db.QueryRow("SELECT COUNT(*) FROM todo_items").Scan(&count)
	if err != nil {
		log.Printf("Error checking for existing data: %v", err)
		return
	}

	// Add sample data if the table is empty
	if count == 0 {
		log.Println("Adding sample todo items...")

		// Add some sample data
		s.CreateTodoItem(models.TodoItem{
			Text:    "Buy groceries",
			Checked: false,
		})

		s.CreateTodoItem(models.TodoItem{
			Text:    "Clean the house",
			Checked: true,
		})

		s.CreateTodoItem(models.TodoItem{
			Text:    "Finish project",
			Checked: false,
		})
	}
}