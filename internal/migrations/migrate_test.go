package migrations

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	ctx := testcontainers.WithReaper(testcontainers.NewSessionReaper())
	
	container, err := postgres.Run(ctx,
		"postgres:13",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	cleanup := func() {
		db.Close()
		container.Terminate(ctx)
	}

	return db, cleanup
}

func TestMigrateUp(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Get test migrations path
	migrationsPath, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatalf("Failed to get migrations path: %v", err)
	}

	// Run migrations up
	err = MigrateUp(db, migrationsPath)
	if err != nil {
		t.Fatalf("Failed to run migrations up: %v", err)
	}

	// Check if schema_migrations table exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'schema_migrations'
		)
	`).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check schema_migrations table: %v", err)
	}

	if !exists {
		t.Error("schema_migrations table was not created")
	}

	// Check if application tables exist
	tables := []string{"scheduled_items", "todo_items", "users"}
	for _, table := range tables {
		var tableExists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = $1
			)
		`, table).Scan(&tableExists)
		if err != nil {
			t.Fatalf("Failed to check table %s: %v", table, err)
		}

		if !tableExists {
			t.Errorf("Table %s was not created", table)
		}
	}
}

func TestMigrateStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Get test migrations path
	migrationsPath, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatalf("Failed to get migrations path: %v", err)
	}

	// Check status before any migrations
	version, dirty, err := MigrateStatus(db, migrationsPath)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if version != 0 {
		t.Errorf("Expected version 0 before migrations, got %d", version)
	}

	if dirty {
		t.Error("Expected clean state before migrations")
	}

	// Run migrations
	err = MigrateUp(db, migrationsPath)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Check status after migrations
	version, dirty, err = MigrateStatus(db, migrationsPath)
	if err != nil {
		t.Fatalf("Failed to get migration status after migrations: %v", err)
	}

	if version == 0 {
		t.Error("Expected version > 0 after migrations")
	}

	if dirty {
		t.Error("Expected clean state after successful migrations")
	}
}

func TestMigrateDown(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Get test migrations path
	migrationsPath, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatalf("Failed to get migrations path: %v", err)
	}

	// Run migrations up first
	err = MigrateUp(db, migrationsPath)
	if err != nil {
		t.Fatalf("Failed to run migrations up: %v", err)
	}

	// Check that tables exist
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'scheduled_items'
		)
	`).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check scheduled_items table: %v", err)
	}

	if !exists {
		t.Error("scheduled_items table should exist after migration up")
	}

	// Run migration down
	err = MigrateDown(db, migrationsPath, 1)
	if err != nil {
		t.Fatalf("Failed to run migration down: %v", err)
	}

	// Check status after rollback
	version, dirty, err := MigrateStatus(db, migrationsPath)
	if err != nil {
		t.Fatalf("Failed to get migration status after rollback: %v", err)
	}

	if dirty {
		t.Error("Expected clean state after rollback")
	}

	// Version should be decremented
	if version >= 2 {
		t.Errorf("Expected lower version after rollback, got %d", version)
	}
}