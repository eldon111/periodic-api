# Integration Tests

This directory contains integration tests for the database stores in the Periodic API.

## Overview

The integration tests verify that the database store implementations work correctly with a real PostgreSQL database. These tests complement the unit tests by testing the actual database interactions.

## Test Structure

- `db_test_setup.go` - Common database setup and teardown for regular tests
- `testcontainer_setup.go` - Testcontainer setup for throwaway PostgreSQL instances
- `container_integration_test.go` - Integration tests using testcontainers
- `scheduled_item_integration_test.go` - Integration tests for ScheduledItem store
- `todo_item_integration_test.go` - Integration tests for TodoItem store  
- `user_integration_test.go` - Integration tests for User store

## Throwaway Database Options

### 1. Testcontainers (Recommended)

**Pros**: Fully isolated, no local setup required, automatic cleanup
**Requirements**: Docker installed

```bash
# From project root
./scripts/test_integration_testcontainers.sh
```

### 2. Docker Compose

**Pros**: Consistent environment, easy to reproduce
**Requirements**: Docker and Docker Compose

```bash
# From project root
./scripts/test_integration_docker.sh
```

### 3. Smart Test Runner

**Pros**: Automatically chooses the best available option
**Requirements**: None (falls back gracefully)

```bash
# From project root  
./scripts/test_integration_smart.sh
```

## Database Configuration

The tests use environment variables for database configuration:

- `TEST_DB_HOST` - Database host (default: localhost)
- `TEST_DB_PORT` - Database port (default: 5432)
- `TEST_DB_USER` - Database username (default: postgres)
- `TEST_DB_PASS` - Database password (default: your-password)
- `TEST_DB_NAME` - Database name (default: test_db)

## Running the Tests

### Option 1: Testcontainers (Recommended)

```bash
# From project root
./scripts/test_integration_testcontainers.sh
```

This approach:
- Automatically downloads and starts a PostgreSQL container
- Creates a fresh database for each test run
- Cleans up automatically when tests complete
- Requires Docker but no local PostgreSQL setup

### Option 2: Docker Compose

```bash
# Start test database
cd test
docker-compose -f docker-compose.test.yml up -d

# Run tests
cd ../test/integration
TEST_DB_HOST=localhost TEST_DB_PORT=5433 TEST_DB_USER=testuser TEST_DB_PASS=testpassword TEST_DB_NAME=test_db go test -v

# Cleanup
cd ../test
docker-compose -f docker-compose.test.yml down -v
```

### Option 3: Local PostgreSQL

```bash
cd test/integration
go test -v
```

### Option 4: Smart Runner (Auto-detect best option)

```bash
# From project root
./scripts/test_integration_smart.sh
```

### Run specific test file:

```bash
cd test/integration  
go test -v -run TestScheduledItemIntegration
go test -v -run TestTodoItemIntegration
go test -v -run TestUserIntegration
```

## Test Features

Each integration test suite includes:

- **Full CRUD Workflow** - Tests create, read, update, delete operations
- **Multiple Items Operations** - Tests operations with multiple records
- **Edge Cases** - Tests error conditions and boundary cases
- **Data Integrity** - Verifies data persists correctly across operations

## Notes

- Tests automatically skip if database connection is not available
- Each test cleans up its data before and after execution
- Tests use separate cleanup functions to avoid interference
- Database tables are recreated for each test run to ensure clean state

## Database Schema

The tests create the following tables:

### scheduled_items
```sql
CREATE TABLE scheduled_items (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    starts_at TIMESTAMP NOT NULL,
    repeats BOOLEAN NOT NULL,
    cron_expression TEXT,
    expiration TIMESTAMP
)
```

### todo_items
```sql
CREATE TABLE todo_items (
    id SERIAL PRIMARY KEY,
    text TEXT NOT NULL,
    checked BOOLEAN NOT NULL DEFAULT FALSE
)
```

### users
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash BYTEA NOT NULL
)
```