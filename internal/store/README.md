# Database Store Testing

This directory contains the implementation of database stores for the application's models, along with unit tests for these stores.

## Testing Approach

The unit tests for the database stores use a real PostgreSQL database for testing. This approach ensures that the SQL queries and database interactions work correctly with a real PostgreSQL database.

### Test Setup

Each test file includes functions to set up a test database connection and create the necessary tables for testing. The tests are designed to be independent and can be run individually or together.

### Running the Tests

To run the tests, you need to have a PostgreSQL database available. You can either:

1. Use a local PostgreSQL database
2. Use a Docker container with PostgreSQL
3. Use an in-memory PostgreSQL database like [pg_tmp](https://eradman.com/ephemeralpg/) or [Zonky's embedded PostgreSQL](https://github.com/zonkyio/embedded-postgres)

#### Configuration

Before running the tests, update the database connection details in the test files:

- `user_db_store_test.go`
- `scheduled_item_db_store_test.go`
- `todo_item_db_store_test.go`

Update the following variables in each file:

```go
dbHost := "localhost"
dbPort := 5432
dbUser := "postgres"
dbPass := "your-password"
dbName := "test_db"
```

#### Running Tests

To run all tests:

```bash
go test ./store
```

To run tests for a specific store:

```bash
go test -run TestUserStore_CRUD ./store
go test -run TestScheduledItemStore_CRUD ./store
go test -run TestTodoItemStore_CRUD ./store
```

## Test Coverage

The tests cover the following operations for each store:

- Create: Adding new items to the database
- Read: Retrieving items by ID and getting all items
- Update: Modifying existing items
- Delete: Removing items from the database

Each test verifies that the operations work correctly and that the data is stored and retrieved as expected.

## Future Improvements

For a more isolated testing environment, consider using one of these approaches:

1. **Docker Containers**: Use the `ory/dockertest` package to spin up a PostgreSQL container for testing.

```go
import (
    "github.com/ory/dockertest/v3"
    "github.com/ory/dockertest/v3/docker"
)
```

2. **Embedded PostgreSQL**: Use a library like Zonky's embedded PostgreSQL.

```go
import (
    "github.com/zonkyio/embedded-postgres"
)
```

3. **SQL Mock**: For pure unit testing without a database, use a SQL mock library.

```go
import (
    "github.com/DATA-DOG/go-sqlmock"
)
```