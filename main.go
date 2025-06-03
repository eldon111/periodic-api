package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"awesomeProject/db"
	"awesomeProject/handlers"
	"awesomeProject/store"
)

func init() {
	// Set the application's default timezone to UTC
	// This ensures all time operations default to UTC
	time.Local = time.UTC
}

func main() {
	var itemStore store.ScheduledItemStore

	// Check environment variable to determine which store to use
	usePostgres := os.Getenv("USE_POSTGRES_DB")
	
	if strings.ToLower(usePostgres) == "true" {
		// Initialize database connection for PostgreSQL
		database, err := db.InitDB()
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer database.Close()
		
		// Create PostgreSQL store instance
		itemStore = store.NewPostgresScheduledItemStore(database)
		log.Println("Using PostgreSQL database for storage")
	} else {
		// Create in-memory store instance
		itemStore = store.NewMemoryScheduledItemStore()
		log.Println("Using in-memory database for storage")
	}

	// Add sample data if needed
	itemStore.AddSampleData()

	// Create handler instance
	itemHandler := handlers.NewScheduledItemHandler(itemStore)

	// Set up routes
	itemHandler.SetupRoutes()

	// Start the server
	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
