#!/bin/bash

echo "Running integration tests..."
echo

cd test/integration

echo "Setting up environment variables (modify as needed):"
export TEST_DB_HOST=${TEST_DB_HOST:-localhost}
export TEST_DB_PORT=${TEST_DB_PORT:-5432}
export TEST_DB_USER=${TEST_DB_USER:-postgres}
export TEST_DB_PASS=${TEST_DB_PASS:-your-password}
export TEST_DB_NAME=${TEST_DB_NAME:-test_db}

echo "TEST_DB_HOST=$TEST_DB_HOST"
echo "TEST_DB_PORT=$TEST_DB_PORT"
echo "TEST_DB_USER=$TEST_DB_USER"
echo "TEST_DB_NAME=$TEST_DB_NAME"

echo
echo "Running all integration tests..."
go test -v

echo
echo "Integration tests completed."
read -p "Press any key to continue..."