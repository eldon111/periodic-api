package store

import (
	"awesomeProject/internal/models"
	"awesomeProject/internal/utils"
	"database/sql"
	"log"
	"sync"
	"time"
)

// PostgresScheduledItemStore provides PostgreSQL storage operations for scheduled items
type PostgresScheduledItemStore struct {
	sync.RWMutex
	db *sql.DB
}

// NewPostgresScheduledItemStore creates a new PostgreSQL store with the given database connection
func NewPostgresScheduledItemStore(db *sql.DB) *PostgresScheduledItemStore {
	return &PostgresScheduledItemStore{
		db: db,
	}
}

// CreateScheduledItem adds a new scheduled item to the database
func (s *PostgresScheduledItemStore) CreateScheduledItem(item models.ScheduledItem) models.ScheduledItem {
	s.Lock()
	defer s.Unlock()

	// Calculate next execution time
	item.NextExecutionAt = utils.CalculateNextExecution(
		item.StartsAt,
		item.Repeats,
		item.CronExpression,
		item.Expiration,
	)

	query := `
		INSERT INTO scheduled_items 
		(title, description, starts_at, repeats, cron_expression, expiration, next_execution_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id
	`

	err := s.db.QueryRow(
		query,
		item.Title,
		item.Description,
		item.StartsAt,
		item.Repeats,
		item.CronExpression,
		item.Expiration,
		item.NextExecutionAt,
	).Scan(&item.ID)

	if err != nil {
		log.Printf("Error creating scheduled item: %v", err)
		return models.ScheduledItem{} // Return empty item on error
	}

	return item
}

// GetScheduledItem retrieves a scheduled item by ID from the database
func (s *PostgresScheduledItemStore) GetScheduledItem(id int64) (models.ScheduledItem, bool) {
	s.RLock()
	defer s.RUnlock()

	var item models.ScheduledItem
	query := `
		SELECT id, title, description, starts_at, repeats, cron_expression, expiration, next_execution_at 
		FROM scheduled_items 
		WHERE id = $1
	`

	var cronExpression sql.NullString
	var expiration sql.NullTime
	var nextExecutionAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.StartsAt,
		&item.Repeats,
		&cronExpression,
		&expiration,
		&nextExecutionAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.ScheduledItem{}, false
		}
		log.Printf("Error getting scheduled item: %v", err)
		return models.ScheduledItem{}, false
	}

	// Handle nullable fields
	if cronExpression.Valid {
		item.CronExpression = &cronExpression.String
	}
	if expiration.Valid {
		item.Expiration = &expiration.Time
	}
	if nextExecutionAt.Valid {
		item.NextExecutionAt = &nextExecutionAt.Time
	}

	return item, true
}

// GetAllScheduledItems returns all scheduled items from the database
func (s *PostgresScheduledItemStore) GetAllScheduledItems() []models.ScheduledItem {
	s.RLock()
	defer s.RUnlock()

	query := `
		SELECT id, title, description, starts_at, repeats, cron_expression, expiration, next_execution_at 
		FROM scheduled_items
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error querying scheduled items: %v", err)
		return []models.ScheduledItem{}
	}
	defer rows.Close()

	var items []models.ScheduledItem
	for rows.Next() {
		var item models.ScheduledItem
		var cronExpression sql.NullString
		var expiration sql.NullTime
		var nextExecutionAt sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.StartsAt,
			&item.Repeats,
			&cronExpression,
			&expiration,
			&nextExecutionAt,
		)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Handle nullable fields
		if cronExpression.Valid {
			item.CronExpression = &cronExpression.String
		}
		if expiration.Valid {
			item.Expiration = &expiration.Time
		}
		if nextExecutionAt.Valid {
			item.NextExecutionAt = &nextExecutionAt.Time
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}

	return items
}

// UpdateScheduledItem updates an existing scheduled item in the database
func (s *PostgresScheduledItemStore) UpdateScheduledItem(id int64, updatedItem models.ScheduledItem) (models.ScheduledItem, bool) {
	s.Lock()
	defer s.Unlock()

	// Calculate next execution time for updated item
	updatedItem.NextExecutionAt = utils.CalculateNextExecution(
		updatedItem.StartsAt,
		updatedItem.Repeats,
		updatedItem.CronExpression,
		updatedItem.Expiration,
	)

	query := `
		UPDATE scheduled_items 
		SET title = $1, description = $2, starts_at = $3, repeats = $4, cron_expression = $5, expiration = $6, next_execution_at = $7 
		WHERE id = $8
	`

	result, err := s.db.Exec(
		query,
		updatedItem.Title,
		updatedItem.Description,
		updatedItem.StartsAt,
		updatedItem.Repeats,
		updatedItem.CronExpression,
		updatedItem.Expiration,
		updatedItem.NextExecutionAt,
		id,
	)

	if err != nil {
		log.Printf("Error updating scheduled item: %v", err)
		return models.ScheduledItem{}, false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return models.ScheduledItem{}, false
	}

	if rowsAffected == 0 {
		return models.ScheduledItem{}, false
	}

	updatedItem.ID = id
	return updatedItem, true
}

// DeleteScheduledItem removes a scheduled item from the database
func (s *PostgresScheduledItemStore) DeleteScheduledItem(id int64) bool {
	s.Lock()
	defer s.Unlock()

	query := `DELETE FROM scheduled_items WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting scheduled item: %v", err)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return false
	}

	return rowsAffected > 0
}

// GetNextScheduledItems returns scheduled items ordered by next execution time
func (s *PostgresScheduledItemStore) GetNextScheduledItems(limit int) []models.ScheduledItem {
	s.RLock()
	defer s.RUnlock()

	query := `
		SELECT id, title, description, starts_at, repeats, cron_expression, expiration, next_execution_at 
		FROM scheduled_items 
		WHERE next_execution_at IS NOT NULL 
		ORDER BY next_execution_at 
		LIMIT $1
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		log.Printf("Error querying next scheduled items: %v", err)
		return []models.ScheduledItem{}
	}
	defer rows.Close()

	var items []models.ScheduledItem
	for rows.Next() {
		var item models.ScheduledItem
		var cronExpression sql.NullString
		var expiration sql.NullTime
		var nextExecutionAt sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.StartsAt,
			&item.Repeats,
			&cronExpression,
			&expiration,
			&nextExecutionAt,
		)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Handle nullable fields
		if cronExpression.Valid {
			item.CronExpression = &cronExpression.String
		}
		if expiration.Valid {
			item.Expiration = &expiration.Time
		}
		if nextExecutionAt.Valid {
			item.NextExecutionAt = &nextExecutionAt.Time
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}

	return items
}

// UpdateExistingItemsNextExecution updates next_execution_at for all existing items that have null values
func (s *PostgresScheduledItemStore) UpdateExistingItemsNextExecution() {
	s.Lock()
	defer s.Unlock()

	log.Println("Updating next_execution_at for existing scheduled items...")

	// Get all items that don't have next_execution_at set
	query := `
		SELECT id, title, description, starts_at, repeats, cron_expression, expiration 
		FROM scheduled_items 
		WHERE next_execution_at IS NULL
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error querying items for migration: %v", err)
		return
	}
	defer rows.Close()

	updateCount := 0
	for rows.Next() {
		var item models.ScheduledItem
		var cronExpression sql.NullString
		var expiration sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.StartsAt,
			&item.Repeats,
			&cronExpression,
			&expiration,
		)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Handle nullable fields
		if cronExpression.Valid {
			item.CronExpression = &cronExpression.String
		}
		if expiration.Valid {
			item.Expiration = &expiration.Time
		}

		// Calculate next execution time
		nextExecution := utils.CalculateNextExecution(
			item.StartsAt,
			item.Repeats,
			item.CronExpression,
			item.Expiration,
		)

		// Update the item with calculated next execution time
		updateQuery := `UPDATE scheduled_items SET next_execution_at = $1 WHERE id = $2`
		_, err = s.db.Exec(updateQuery, nextExecution, item.ID)
		if err != nil {
			log.Printf("Error updating item %d: %v", item.ID, err)
			continue
		}

		updateCount++
	}

	log.Printf("Updated next_execution_at for %d existing items", updateCount)
}

// AddSampleData adds sample data to the database if it's empty
func (s *PostgresScheduledItemStore) AddSampleData() {
	count := 0
	err := s.db.QueryRow("SELECT COUNT(*) FROM scheduled_items").Scan(&count)
	if err != nil {
		log.Printf("Error checking for existing data: %v", err)
		return
	}

	// Add sample data if the table is empty
	if count == 0 {
		log.Println("Adding sample data...")

		// Add some sample data
		startsAt1, _ := time.Parse(time.RFC3339, "2023-05-15T10:00:00Z")
		s.CreateScheduledItem(models.ScheduledItem{
			Title:       "Sample Scheduled Item 1",
			Description: "Description for item 1",
			StartsAt:    startsAt1,
			Repeats:     false,
		})

		cronExpr := "0 0 9 * * MON-FRI"
		startsAt2, _ := time.Parse(time.RFC3339, "2023-05-16T14:30:00Z")
		expirationTime, _ := time.Parse(time.RFC3339, "2023-12-31T23:59:59Z")
		s.CreateScheduledItem(models.ScheduledItem{
			Title:          "Sample Scheduled Item 2",
			Description:    "Description for item 2",
			StartsAt:       startsAt2,
			Repeats:        true,
			CronExpression: &cronExpr,
			Expiration:     &expirationTime,
		})
	} else {
		// If there are existing items, update their next_execution_at values
		s.UpdateExistingItemsNextExecution()
	}
}
