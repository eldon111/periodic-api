package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"periodic-api/internal/db"
	"periodic-api/internal/migrations"
)

func main() {
	var (
		action      = flag.String("action", "up", "Migration action: up, down, status, version, force")
		steps       = flag.Int("steps", 1, "Number of steps for down migration")
		version     = flag.Uint("version", 0, "Target version for migrate to specific version")
		forceVer    = flag.Int("force", -1, "Force version (use with caution)")
		migrationsDir = flag.String("path", "migrations", "Path to migrations directory")
	)
	flag.Parse()

	// Get absolute path to migrations directory
	absPath, err := filepath.Abs(*migrationsDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Check if migrations directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Fatalf("Migrations directory does not exist: %s", absPath)
	}

	// Initialize database connection
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	switch *action {
	case "up":
		if err := runMigrationsUp(database, absPath); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := runMigrationsDown(database, absPath, *steps); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Printf("Rolled back %d migration(s) successfully\n", *steps)

	case "status":
		if err := showMigrationStatus(database, absPath); err != nil {
			log.Fatalf("Failed to show migration status: %v", err)
		}

	case "version":
		if *version == 0 {
			fmt.Println("Please specify a target version with -version flag")
			os.Exit(1)
		}
		if err := migrateTo(database, absPath, *version); err != nil {
			log.Fatalf("Migration to version %d failed: %v", *version, err)
		}
		fmt.Printf("Migrated to version %d successfully\n", *version)

	case "force":
		if *forceVer < 0 {
			fmt.Println("Please specify a version to force with -force flag")
			os.Exit(1)
		}
		if err := forceVersion(database, absPath, *forceVer); err != nil {
			log.Fatalf("Force version %d failed: %v", *forceVer, err)
		}
		fmt.Printf("Forced version to %d successfully\n", *forceVer)

	default:
		fmt.Printf("Unknown action: %s. Use: up, down, status, version, or force\n", *action)
		os.Exit(1)
	}
}

func runMigrationsUp(database *sql.DB, migrationsPath string) error {
	fmt.Println("Running pending migrations...")
	return migrations.MigrateUp(database, migrationsPath)
}

func runMigrationsDown(database *sql.DB, migrationsPath string, steps int) error {
	if steps <= 0 {
		return fmt.Errorf("steps must be greater than 0 for rollback")
	}
	
	fmt.Printf("Rolling back %d migration(s)...\n", steps)
	return migrations.MigrateDown(database, migrationsPath, steps)
}

func showMigrationStatus(database *sql.DB, migrationsPath string) error {
	version, dirty, err := migrations.MigrateStatus(database, migrationsPath)
	if err != nil {
		return err
	}

	fmt.Println("Migration Status:")
	fmt.Println("================")
	
	if version == 0 {
		fmt.Println("Current version: No migrations applied")
	} else {
		fmt.Printf("Current version: %d\n", version)
	}
	
	if dirty {
		fmt.Println("Status: DIRTY (migration failed, needs manual intervention)")
	} else {
		fmt.Println("Status: CLEAN")
	}

	return nil
}

func migrateTo(database *sql.DB, migrationsPath string, version uint) error {
	fmt.Printf("Migrating to version %d...\n", version)
	return migrations.MigrateTo(database, migrationsPath, version)
}

func forceVersion(database *sql.DB, migrationsPath string, version int) error {
	fmt.Printf("Forcing version to %d...\n", version)
	return migrations.ForceVersion(database, migrationsPath, version)
}