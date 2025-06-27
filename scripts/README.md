# Scripts

This directory contains shell scripts for testing and running various operations in the Periodic API.

## Available Scripts

### API Testing Scripts

- **`test_api.sh`** - Tests the main REST API endpoints (CRUD operations for scheduled items)
- **`test_generate_api.sh`** - Tests the AI-powered schedule generation endpoint

### Integration Testing Scripts

- **`test_integration.sh`** - Runs integration tests against a local PostgreSQL database
- **`test_integration_docker.sh`** - Runs integration tests using Docker Compose PostgreSQL container
- **`test_integration_testcontainers.sh`** - Runs integration tests using testcontainers (automatic PostgreSQL containers)
- **`test_integration_smart.sh`** - Smart test runner that auto-detects the best available testing option

## Usage

All scripts are executable and can be run from the project root directory:

```bash
# API testing
./scripts/test_api.sh
./scripts/test_generate_api.sh

# Integration testing (recommended: smart runner)
./scripts/test_integration_smart.sh

# Or specific integration test approaches
./scripts/test_integration_testcontainers.sh  # Testcontainers (recommended)
./scripts/test_integration.sh                 # Local PostgreSQL
./scripts/test_integration_docker.sh          # Docker Compose
```

## Prerequisites

### For API Testing Scripts
- Go application running on `localhost:8080`
- `curl` command available

### For Integration Testing Scripts
- **Local PostgreSQL approach**: PostgreSQL server running with test database
- **Docker approach**: Docker and Docker Compose installed
- **Smart runner**: No specific requirements (automatically chooses best option)

## Environment Variables

Integration test scripts support these environment variables:

- `TEST_DB_HOST` - Database host (default: localhost)
- `TEST_DB_PORT` - Database port (default: 5432)
- `TEST_DB_USER` - Database username (default: postgres)
- `TEST_DB_PASS` - Database password (default: your-password)
- `TEST_DB_NAME` - Database name (default: test_db)

Example:
```bash
TEST_DB_HOST=mydb.example.com TEST_DB_PORT=5433 ./scripts/test_integration.sh
```

## Script Details

### `test_integration_smart.sh` (Recommended)

This script automatically detects available testing options and chooses the best one:

1. **First choice**: Testcontainers (if Docker available)
2. **Second choice**: Docker Compose (if Docker available)
3. **Third choice**: Local PostgreSQL (if PostgreSQL detected on port 5432)
4. **Fallback**: Error message with setup instructions

This approach provides the most reliable testing experience across different environments.