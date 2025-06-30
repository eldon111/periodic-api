// Package main implements the Periodic API server
// @title Periodic API
// @version 1.0
// @description A REST API server for managing Periodic items with support for PostgreSQL and in-memory storage.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email eldon+periodic@emathias.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "periodic-api/docs"
	"periodic-api/internal/db"
	"periodic-api/internal/handlers"
	"periodic-api/internal/migrations"
	"periodic-api/internal/store"

	httpSwagger "github.com/swaggo/http-swagger"
)

func init() {
	// Set the application's default timezone to UTC
	// This ensures all time operations default to UTC
	time.Local = time.UTC
}

func main() {
	var itemStore store.ScheduledItemStore
	var todoStore store.TodoItemStore
	var userStore store.UserStore
	// var executionLogStore store.ExecutionLogStore // Will be used in future chunks

	// Check environment variable to determine which store to use
	usePostgres := os.Getenv("USE_POSTGRES_DB")

	if strings.ToLower(usePostgres) == "true" {
		// Initialize database connection for PostgreSQL
		database, err := db.InitDB()
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer database.Close()

		// Run migrations if auto-migration is enabled
		autoMigrate := os.Getenv("AUTO_MIGRATE")
		if autoMigrate == "" || strings.ToLower(autoMigrate) == "true" {
			log.Println("Running database migrations...")

			// Get migrations directory path
			migrationsPath := "migrations"
			if customPath := os.Getenv("MIGRATIONS_PATH"); customPath != "" {
				migrationsPath = customPath
			}

			absPath, err := filepath.Abs(migrationsPath)
			if err != nil {
				log.Fatalf("Failed to get absolute path for migrations: %v", err)
			}

			// Check if migrations directory exists
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				log.Printf("Migrations directory does not exist: %s. Skipping auto-migration.", absPath)
			} else {
				if err := migrations.MigrateUp(database, absPath); err != nil {
					log.Fatalf("Failed to run migrations: %v", err)
				}
				log.Println("Database migrations completed successfully")
			}
		}

		// Create PostgreSQL store instances
		itemStore = store.NewPostgresScheduledItemStore(database)
		todoStore = store.NewPostgresTodoItemStore(database)
		userStore = store.NewPostgresUserStore(database)
		// executionLogStore = store.NewPostgresExecutionLogStore(database) // Will be used in future chunks
		log.Println("Using PostgreSQL database for storage")
	} else {
		// Create in-memory store instances
		itemStore = store.NewMemoryScheduledItemStore()
		todoStore = store.NewMemoryTodoItemStore()
		userStore = store.NewMemoryUserStore()
		// executionLogStore = store.NewMemoryExecutionLogStore() // Will be used in future chunks
		log.Println("Using in-memory database for storage")
	}

	// Add sample data for stores
	todoStore.AddSampleData()
	userStore.AddSampleData()

	// Create handler instances
	itemHandler := handlers.NewScheduledItemHandler(itemStore)
	todoHandler := handlers.NewTodoItemHandler(todoStore)
	userHandler := handlers.NewUserHandler(userStore)

	// Set up routes
	itemHandler.SetupRoutes()
	todoHandler.SetupRoutes()
	userHandler.SetupRoutes()

	// Add Swagger documentation endpoint
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Start the server
	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	fmt.Printf("API documentation available at: http://localhost%s/swagger/\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
