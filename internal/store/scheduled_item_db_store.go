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
func (s *PostgresScheduledItemStore) GetNextScheduledItems(limit int, offset int64) ([]models.ScheduledItem, error) {
	s.RLock()
	defer s.RUnlock()

	now := time.Now()

	query := `
		SELECT id, title, description, starts_at, repeats, cron_expression, expiration, next_execution_at 
		FROM scheduled_items 
		WHERE next_execution_at <= $1 
		ORDER BY next_execution_at 
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, now, limit, offset)
	if err != nil {
		return []models.ScheduledItem{}, err
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
			return []models.ScheduledItem{}, err
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
		return []models.ScheduledItem{}, err
	}

	return items, nil
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
	}
}
