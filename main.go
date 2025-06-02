package main

import (
	"fmt"
	"log"
	"net/http"

	"awesomeProject/db"
	"awesomeProject/handlers"
	"awesomeProject/store"
)

func main() {
	// Initialize database connection
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create store instance
	itemStore := store.NewScheduledItemStore(database)

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
