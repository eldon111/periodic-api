package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB holds the connection to the test database
var TestDB *sql.DB

// TestContainerDB holds the connection to the testcontainer database
var TestContainerDB *sql.DB

func init() {
	// Set the application's default timezone to UTC
	time.Local = time.UTC
	log.Printf("default time set to UTC")
}

// TestMain sets up the test database and runs all tests
func TestMain(m *testing.M) {
	// Check if we should use testcontainers
	if os.Getenv("USE_TESTCONTAINERS") == "true" {
		TestMainWithContainer(m)
		return
	}

	// Set up the test database connection (regular approach)
	var err error
	TestDB, err = setupTestDB()
	if err != nil {
		log.Printf("Failed to set up test database: %v", err)
		log.Printf("Skipping integration tests - database not available")
		os.Exit(0)
	}
	defer TestDB.Close()

	// Run the tests
	exitCode := m.Run()

	os.Exit(exitCode)
}

// regularTestMain runs the regular test setup without checking for testcontainers
func regularTestMain(m *testing.M) {
	// Set up the test database connection (regular approach)
	var err error
	TestDB, err = setupTestDB()
	if err != nil {
		log.Printf("Failed to set up test database: %v", err)
		log.Printf("Skipping integration tests - database not available")
		os.Exit(0)
	}
	defer TestDB.Close()

	// Run the tests
	exitCode := m.Run()

	os.Exit(exitCode)
}

// setupTestDB creates a connection to the test database
func setupTestDB() (*sql.DB, error) {
	// Use environment variables for database configuration
	dbHost := getEnvOrDefault("TEST_DB_HOST", "localhost")
	dbPort := getEnvOrDefault("TEST_DB_PORT", "5432")
	dbUser := getEnvOrDefault("TEST_DB_USER", "postgres")
	dbPass := getEnvOrDefault("TEST_DB_PASS", "your-password")
	dbName := getEnvOrDefault("TEST_DB_NAME", "test_db")

	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	// Create all required tables
	if err = createAllTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

// createAllTables creates all required tables for testing by running migrations
// This ensures test database schema matches the actual application schema
func createAllTables(db *sql.DB) error {
	// Get the path to migrations directory relative to the project root
	// Go up from test/integration to project root
	projectRoot := filepath.Join("..", "..")
	migrationsPath := filepath.Join(projectRoot, "migrations")

	// Run migrations to create the database schema
	// This replaces the old approach of using db_init.sql to ensure
	// tests always use the current, up-to-date schema
	err := runMigrations(db, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations executes all .up.sql migration files in order
func runMigrations(db *sql.DB, migrationsPath string) error {
	// First, create the schema_migrations table if it doesn't exist
	createMigrationsTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		);
	`
	_, err := db.Exec(createMigrationsTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Read migration files
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort .up.sql files
	var upFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			upFiles = append(upFiles, file.Name())
		}
	}
	sort.Strings(upFiles)

	// Execute each migration file
	for _, fileName := range upFiles {
		// Extract version number from filename (e.g., "000001_initial_schema.up.sql" -> 1)
		version := extractVersionFromFilename(fileName)

		// Check if this migration has already been applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status for version %d: %w", version, err)
		}

		if count > 0 {
			log.Printf("Migration %s already applied, skipping", fileName)
			continue
		}

		// Read and execute the migration file
		filePath := filepath.Join(migrationsPath, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
		}

		// Execute the SQL
		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
		}

		// Record that this migration has been applied
		_, err = db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES ($1, FALSE)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", fileName, err)
		}

		log.Printf("Applied migration: %s", fileName)
	}

	return nil
}

// extractVersionFromFilename extracts the version number from a migration filename
func extractVersionFromFilename(filename string) int64 {
	// Migration files are named like "000001_description.up.sql"
	parts := strings.Split(filename, "_")
	if len(parts) == 0 {
		return 0
	}

	versionStr := parts[0]
	// Convert version string to int64 (e.g., "000001" -> 1)
	var version int64
	for _, char := range versionStr {
		if char >= '0' && char <= '9' {
			version = version*10 + int64(char-'0')
		}
	}

	return version
}

// getEnvOrDefault returns the environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// skipIfDBNotAvailable skips the test if the database is not available
func skipIfDBNotAvailable(t *testing.T) {
	// If neither database is set up, try to set up based on environment
	if TestDB == nil && TestContainerDB == nil {
		if os.Getenv("USE_TESTCONTAINERS") == "true" {
			if err := setupTestContainerDB(); err != nil {
				t.Skipf("Failed to set up testcontainer database: %v", err)
			}
		} else {
			var err error
			TestDB, err = setupTestDB()
			if err != nil {
				t.Skipf("Failed to set up regular database: %v", err)
			}
		}
	}

	if TestDB == nil && TestContainerDB == nil {
		t.Skip("Test database connection not available")
	}
}

// getActiveDB returns the active database connection (either TestDB or TestContainerDB)
func getActiveDB() *sql.DB {
	if TestContainerDB != nil {
		return TestContainerDB
	}
	return TestDB
}

// TestContainerInstance holds the container instance
var TestContainerInstance testcontainers.Container

// TestMainWithContainer sets up a PostgreSQL testcontainer and runs all tests
func TestMainWithContainer(m *testing.M) {
	// Check if Docker is available
	if !isDockerAvailable() {
		log.Println("Docker not available, skipping container-based tests")
		log.Println("Falling back to regular database connection if available")
		// Fall back to regular test setup - but avoid infinite recursion
		regularTestMain(m)
		return
	}

	ctx := context.Background()

	// Start PostgreSQL container
	container, err := startPostgresContainer(ctx)
	if err != nil {
		log.Printf("Failed to start PostgreSQL container: %v", err)
		log.Println("Falling back to regular database connection if available")
		regularTestMain(m)
		return
	}
	TestContainerInstance = container

	// Get database connection
	db, err := connectToContainer(ctx, container)
	if err != nil {
		log.Printf("Failed to connect to container database: %v", err)
		container.Terminate(ctx)
		os.Exit(1)
	}
	TestContainerDB = db

	// Run the tests
	exitCode := m.Run()

	// Clean up
	if db != nil {
		db.Close()
	}
	if container != nil {
		container.Terminate(ctx)
	}

	os.Exit(exitCode)
}

// startPostgresContainer starts a PostgreSQL testcontainer
func startPostgresContainer(ctx context.Context) (testcontainers.Container, error) {
	dbName := "testdb"
	dbUser := "testuser"
	dbPassword := "testpassword"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return postgresContainer, nil
}

// connectToContainer creates a database connection to the PostgreSQL container
func connectToContainer(ctx context.Context, container testcontainers.Container) (*sql.DB, error) {
	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), "testuser", "testpassword", "testdb")

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// Test the connection with retries
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("db.Ping after %d retries: %w", maxRetries, err)
	}

	// Create all required tables
	if err = createAllTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

// isDockerAvailable checks if Docker is available
func isDockerAvailable() bool {
	ctx := context.Background()

	// Try to get Docker client
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return false
	}
	defer provider.Close()

	// Try to ping Docker
	_, err = provider.DaemonHost(ctx)
	return err == nil
}

// setupTestContainerDB sets up a testcontainer database for testing
func setupTestContainerDB() error {
	// Check if Docker is available
	if !isDockerAvailable() {
		return fmt.Errorf("Docker not available")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	container, err := startPostgresContainer(ctx)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	TestContainerInstance = container

	// Get database connection
	db, err := connectToContainer(ctx, container)
	if err != nil {
		container.Terminate(ctx)
		return fmt.Errorf("failed to connect to container: %w", err)
	}
	TestContainerDB = db

	return nil
}
