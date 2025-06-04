# REST API Server

A simple REST API server for CRUD operations built with Go's standard library.

## Overview

This project implements a basic REST web server that provides CRUD (Create, Read, Update, Delete) operations for a "ScheduledItem" resource. The server uses a PostgreSQL database for persistent storage and provides JSON-based API endpoints.

## Features

- RESTful API design
- PostgreSQL database for persistent storage
- Thread-safe database operations
- Full CRUD operations:
  - Create new scheduled items
  - Read individual scheduled items or list all scheduled items
  - Update existing scheduled items
  - Delete scheduled items
- JSON request/response format
- Google Cloud SQL compatibility

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | /scheduled-items   | List all scheduled items |
| POST   | /scheduled-items   | Create a new scheduled item |
| GET    | /scheduled-items/{id} | Get a specific scheduled item by ID |
| PUT    | /scheduled-items/{id} | Update a specific scheduled item |
| DELETE | /scheduled-items/{id} | Delete a specific scheduled item |

## Data Model

Each ScheduledItem has the following structure:

```json
{
  "id": 1,
  "title": "Scheduled Item Name",
  "description": "Detailed description of the scheduled item",
  "startsAt": "2023-05-15T10:00:00Z",
  "repeats": false,
  "cronExpression": "0 0 9 * * MON-FRI",
  "expiration": "2023-12-31T23:59:59Z"
}
```

Note: `cronExpression` and `expiration` fields are optional and will only appear when relevant.

## Running the Server

To run the server:

```bash
go run main.go
```

The server will start on port 8080 and initialize with two sample scheduled items.

## Testing the API

A test script is included to verify the API functionality. To run the tests:

```bash
.\test_api.bat
```

This script will:
1. Get all scheduled items (showing initial sample data)
2. Create a new scheduled item
3. Get all scheduled items again (including the new item)
4. Get a specific scheduled item by ID
5. Update a scheduled item
6. Get the updated scheduled item
7. Delete a scheduled item
8. Get all scheduled items again (showing the deletion worked)

## Example Requests

### Create a Scheduled Item

```bash
curl -X POST http://localhost:8080/scheduled-items -H "Content-Type: application/json" -d "{\"title\":\"New Scheduled Item\",\"description\":\"Description of the new item\",\"startsAt\":\"2023-05-20T15:00:00Z\",\"repeats\":false}"
```

### Get All Scheduled Items

```bash
curl -X GET http://localhost:8080/scheduled-items
```

### Get Scheduled Item by ID

```bash
curl -X GET http://localhost:8080/scheduled-items/1
```

### Update a Scheduled Item

```bash
curl -X PUT http://localhost:8080/scheduled-items/1 -H "Content-Type: application/json" -d "{\"title\":\"Updated Scheduled Item\",\"description\":\"Updated description\",\"startsAt\":\"2023-05-21T16:30:00Z\",\"repeats\":true,\"cronExpression\":\"0 0 12 * * *\",\"expiration\":\"2023-12-31T23:59:59Z\"}"
```

### Delete a Scheduled Item

```bash
curl -X DELETE http://localhost:8080/scheduled-items/1
```

## Database Configuration

The application uses PostgreSQL for persistent storage. You need to configure the database connection in the `main.go` file:

```go
// Database connection details
const (
    // Update these with your actual PostgreSQL instance details
    dbHost     = "localhost" // Use your Google Cloud SQL IP when deploying
    dbPort     = 5432
    dbUser     = "postgres"
    dbPass     = "your-password"
    dbName     = "scheduled_items_db"
)
```

### Local Development Setup

1. Install PostgreSQL on your local machine
2. Create a database named `scheduled_items_db`
3. Update the connection details in `main.go` if necessary
4. Run the application with `go run main.go`

The application will automatically create the necessary tables on startup.

### Google Cloud Deployment

To deploy on Google Cloud with Cloud SQL for PostgreSQL:

1. Create a PostgreSQL instance in Google Cloud SQL
2. Update the connection details in `main.go` with your Cloud SQL instance information
3. Deploy your application to Google Cloud Run or Google Compute Engine

## Dependencies

- Go 1.24 or later
- PostgreSQL 12 or later
- github.com/lib/pq - PostgreSQL driver for Go

## Future Improvements

- Implement authentication and authorization
- Add input validation
- Implement pagination for listing scheduled items
- Add filtering and sorting options
- Add database migration support
