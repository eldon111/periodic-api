#!/bin/bash

echo "Starting Docker-based integration tests..."
echo

echo "Step 1: Starting PostgreSQL test container..."
cd test
docker-compose -f docker-compose.test.yml up -d test-postgres

echo "Waiting for PostgreSQL to be ready..."
sleep 10

echo "Step 2: Running integration tests..."
cd ../test/integration

echo "Setting Docker-based environment variables:"
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5433
export TEST_DB_USER=testuser
export TEST_DB_PASS=testpassword
export TEST_DB_NAME=test_db

echo "TEST_DB_HOST=$TEST_DB_HOST"
echo "TEST_DB_PORT=$TEST_DB_PORT"
echo "TEST_DB_USER=$TEST_DB_USER"
echo "TEST_DB_NAME=$TEST_DB_NAME"

echo
echo "Running integration tests against Docker PostgreSQL..."
go test -v

echo
echo "Step 3: Cleaning up..."
cd ../test
docker-compose -f docker-compose.test.yml down -v

echo
echo "Docker-based integration tests completed."
read -p "Press any key to continue..."