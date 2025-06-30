package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"periodic-api/internal/db"
	"periodic-api/internal/store"
)

func init() {
	// Set the application's default timezone to UTC
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
		log.Println("Scheduler using PostgreSQL database for storage")
	} else {
		// Create in-memory store instance
		itemStore = store.NewMemoryScheduledItemStore()
		log.Println("Scheduler using in-memory database for storage")
	}

	// Get interval from environment variable, default to 30 seconds
	interval := 30 * time.Second
	if intervalStr := os.Getenv("SCHEDULER_INTERVAL"); intervalStr != "" {
		if parsedInterval, err := time.ParseDuration(intervalStr); err == nil {
			interval = parsedInterval
		} else {
			log.Printf("Invalid SCHEDULER_INTERVAL format, using default: %v", interval)
		}
	}

	log.Printf("Starting scheduler service with interval: %v", interval)

	// Create a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create ticker for periodic execution
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run initial check
	processScheduledItems(itemStore)

	// Main service loop
	for {
		select {
		case <-ticker.C:
			processScheduledItems(itemStore)
		case <-sigChan:
			log.Println("Received shutdown signal, stopping scheduler...")
			return
		}
	}
}

func processScheduledItems(store store.ScheduledItemStore) {
	log.Println("Processing scheduled items...")

	items := store.GetAllScheduledItems()

	log.Printf("Found %d scheduled items", len(items))

	// TODO: Add your processing logic here
	// For now, just log the items
	for _, item := range items {
		log.Printf("Item: ID=%d, Title='%s', StartsAt=%v",
			item.ID, item.Title, item.StartsAt)
	}

	log.Println("Finished processing scheduled items")
}
