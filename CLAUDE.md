# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **Periodic API** - a Go REST API server that serves as the backend for the Periodic app. It provides CRUD operations for scheduled items and features a dual-storage architecture that can switch between PostgreSQL (for production) and in-memory storage (for local development) via the `USE_POSTGRES_DB` environment variable.

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
# Testcontainers (recommended - uses throwaway Docker containers)
./scripts/test_integration_testcontainers.sh

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

### OpenAPI/Swagger Documentation
```bash
# Generate OpenAPI specification from code annotations
~/go/bin/swag init -g cmd/app/main.go -o docs

# Alternative: use go run if swag is not in PATH
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/app/main.go -o docs
```

### Database Migrations
```bash
# Run all pending migrations
go run cmd/migrate/main.go -action=up

# Check migration status
go run cmd/migrate/main.go -action=status

# Rollback migrations (rollback 1 step)
go run cmd/migrate/main.go -action=down -steps=1

# Migrate to specific version
go run cmd/migrate/main.go -action=version -version=1

# Force migration version (use with caution)
go run cmd/migrate/main.go -action=force -force=1

# Use custom migrations directory
go run cmd/migrate/main.go -action=up -path=custom/migrations/path
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

PostgreSQL connection details are configured via environment variables in `internal/db/db.go`:
- `DB_HOST` (default: "localhost")
- `DB_PORT` (default: "5432") 
- `DB_USER` (default: "postgres")
- `DB_PASSWORD` (default: "your-password")
- `DB_NAME` (default: "scheduled_items_db")
- `DB_SSL_MODE` (default: "disable")

## Database Migrations

The application uses [golang-migrate/migrate](https://github.com/golang-migrate/migrate) for database schema management:

### Migration System Features
- **Tracking**: Uses `schema_migrations` table to track applied migrations
- **Auto-migration**: Runs pending migrations on app startup (controlled by `AUTO_MIGRATE` env var)
- **CLI Tool**: Standalone migration tool at `cmd/migrate/main.go`
- **Rollback Support**: Can rollback migrations with down SQL files
- **Version Control**: Migrate to specific versions or force version
- **Dirty State Detection**: Detects and handles failed migrations

### Migration Files
- Migration files are stored in `migrations/` directory
- Format: `YYYYMMDDHHMMSS_description.up.sql` and `YYYYMMDDHHMMSS_description.down.sql`
- Example: `000001_initial_schema.up.sql` and `000001_initial_schema.down.sql`
- Each migration requires both up and down files

### Environment Variables
- `AUTO_MIGRATE=true` (default): Run migrations on app startup
- `AUTO_MIGRATE=false`: Skip automatic migrations (use CLI tool)  
- `MIGRATIONS_PATH=migrations` (default): Path to migrations directory

### Creating New Migrations
1. Create sequential numbered migration files:
   ```
   migrations/000003_add_new_table.up.sql
   migrations/000003_add_new_table.down.sql
   ```
2. Write forward migration SQL in `.up.sql` file
3. Write rollback migration SQL in `.down.sql` file