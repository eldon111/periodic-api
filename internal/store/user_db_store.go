package store

import (
	"periodic-api/internal/models"
	"database/sql"
	"log"
	"sync"
)

// PostgresUserStore provides PostgreSQL storage operations for users
type PostgresUserStore struct {
	sync.RWMutex
	db *sql.DB
}

// NewPostgresUserStore creates a new PostgreSQL store with the given database connection
func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{
		db: db,
	}
}

// CreateUser adds a new user to the database
func (s *PostgresUserStore) CreateUser(user models.User) models.User {
	s.Lock()
	defer s.Unlock()

	query := `
		INSERT INTO users 
		(username, password_hash) 
		VALUES ($1, $2) 
		RETURNING id
	`

	err := s.db.QueryRow(
		query,
		user.Username,
		user.PasswordHash,
	).Scan(&user.ID)

	if err != nil {
		log.Printf("Error creating user: %v", err)
		return models.User{} // Return empty user on error
	}

	return user
}

// GetUser retrieves a user by ID from the database
func (s *PostgresUserStore) GetUser(id int64) (models.User, bool) {
	s.RLock()
	defer s.RUnlock()

	var user models.User
	query := `
		SELECT id, username, password_hash 
		FROM users 
		WHERE id = $1
	`

	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, false
		}
		log.Printf("Error getting user: %v", err)
		return models.User{}, false
	}

	return user, true
}

// GetAllUsers returns all users from the database
func (s *PostgresUserStore) GetAllUsers() []models.User {
	s.RLock()
	defer s.RUnlock()

	query := `
		SELECT id, username, password_hash 
		FROM users
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error querying users: %v", err)
		return []models.User{}
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
		)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}

	return users
}

// UpdateUser updates an existing user in the database
func (s *PostgresUserStore) UpdateUser(id int64, updatedUser models.User) (models.User, bool) {
	s.Lock()
	defer s.Unlock()

	query := `
		UPDATE users 
		SET username = $1, password_hash = $2 
		WHERE id = $3
	`

	result, err := s.db.Exec(
		query,
		updatedUser.Username,
		updatedUser.PasswordHash,
		id,
	)

	if err != nil {
		log.Printf("Error updating user: %v", err)
		return models.User{}, false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return models.User{}, false
	}

	if rowsAffected == 0 {
		return models.User{}, false
	}

	updatedUser.ID = id
	return updatedUser, true
}

// DeleteUser removes a user from the database
func (s *PostgresUserStore) DeleteUser(id int64) bool {
	s.Lock()
	defer s.Unlock()

	query := `DELETE FROM users WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
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
func (s *PostgresUserStore) AddSampleData() {
	count := 0
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Printf("Error checking for existing data: %v", err)
		return
	}

	// Add sample data if the table is empty
	if count == 0 {
		log.Println("Adding sample user data...")

		// Add some sample data
		s.CreateUser(models.User{
			Username: "admin",
			PasswordHash: []byte("admin123"),
		})

		s.CreateUser(models.User{
			Username: "user1",
			PasswordHash: []byte("password123"),
		})
	}
}
