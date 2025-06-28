package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// buildFileURL creates a proper file URL for the migrations path
func buildFileURL(migrationsPath string) string {
	// Convert to absolute path and normalize separators
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		// Fallback to the original path if Abs fails
		absPath = migrationsPath
	}

	if runtime.GOOS == "windows" {
		// On Windows, use a different approach - convert to Unix-style path
		absPath = strings.ReplaceAll(absPath, "\\", "/")
		// Handle drive letters properly for Windows file URLs
		if len(absPath) >= 2 && absPath[1] == ':' {
			// For Windows: C:/path -> file:///C:/path
			return "file:///" + absPath
		}
	}

	// For Unix-like systems
	return "file://" + absPath
}

// MigrateUp runs all pending migrations
func MigrateUp(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	sourceURL := buildFileURL(migrationsPath)
	log.Printf("Debug: Using migrations path: %s", migrationsPath)
	log.Printf("Debug: Generated source URL: %s", sourceURL)

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	return nil
}

// MigrateDown rolls back migrations
func MigrateDown(db *sql.DB, migrationsPath string, steps int) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	sourceURL := buildFileURL(migrationsPath)
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run down migrations: %w", err)
	}

	return nil
}

// MigrateStatus returns the current migration version and status
func MigrateStatus(db *sql.DB, migrationsPath string) (uint, bool, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("could not create postgres driver: %w", err)
	}

	sourceURL := buildFileURL(migrationsPath)
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("could not get migration version: %w", err)
	}

	return version, dirty, nil
}

// MigrateTo migrates to a specific version
func MigrateTo(db *sql.DB, migrationsPath string, version uint) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	sourceURL := buildFileURL(migrationsPath)
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Migrate(version); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not migrate to version %d: %w", version, err)
	}

	return nil
}

// ForceVersion sets the migration version without running migrations
func ForceVersion(db *sql.DB, migrationsPath string, version int) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	sourceURL := buildFileURL(migrationsPath)
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		return fmt.Errorf("could not force version %d: %w", version, err)
	}

	return nil
}
