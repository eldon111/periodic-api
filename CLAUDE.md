# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go REST API server providing CRUD operations for scheduled items. The application features a dual-storage architecture that can switch between PostgreSQL (for production) and in-memory storage (for local development) via the `USE_POSTGRES_DB` environment variable.

## Common Commands

### Running the application
```bash
go run main.go
```

### Testing the API
```bash
./scripts/test_api.sh
./scripts/test_generate_api.sh
```

### Integration Testing
```bash
# Smart runner (auto-detects best option)
./scripts/test_integration_smart.sh

# Testcontainers (recommended - uses throwaway Docker containers)
./scripts/test_integration_testcontainers.sh

# Docker Compose approach
./scripts/test_integration_docker.sh

# Local PostgreSQL approach
./scripts/test_integration.sh
```

### Build
```bash
go build
```

### Get dependencies
```bash
go mod tidy
```

## Architecture

### Storage Layer
The application uses a storage interface pattern (`store.ScheduledItemStore`) with two implementations:
- `MemoryScheduledItemStore`: Thread-safe in-memory storage for local development
- `PostgresScheduledItemStore`: PostgreSQL storage for production

Storage selection is controlled by the `USE_POSTGRES_DB` environment variable:
- `USE_POSTGRES_DB=true`: Uses PostgreSQL (requires database setup)
- `USE_POSTGRES_DB` unset or any other value: Uses in-memory storage

### Package Structure
- `models/`: Data models (ScheduledItem struct)
- `store/`: Storage interface and implementations
- `handlers/`: HTTP request handlers and routing
- `db/`: PostgreSQL database initialization and configuration
- `middleware/`: HTTP middleware components

### Data Model
The core entity is `ScheduledItem` with fields:
- ID, Title, Description, StartsAt (required)
- Repeats (boolean), CronExpression, Expiration (optional)

### API Endpoints
- `GET /scheduled-items` - List all items
- `POST /scheduled-items` - Create new item
- `GET /scheduled-items/{id}` - Get specific item
- `PUT /scheduled-items/{id}` - Update item
- `DELETE /scheduled-items/{id}` - Delete item
- `POST /generate-scheduled-item` - Generate item from text prompt using AWS LLM

## Database Configuration

PostgreSQL connection details are hardcoded in `db/db.go`. For local PostgreSQL development, update the constants:
```go
const (
    dbHost = "localhost"
    dbPort = 5432
    dbUser = "postgres"
    dbPass = "your-password"
    dbName = "scheduled_items_db"
)
```

The application automatically creates the `scheduled_items` table on startup.