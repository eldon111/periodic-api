#!/bin/bash

echo "Running integration tests with testcontainers..."
echo "This will automatically spin up and tear down a PostgreSQL container"
echo

# Check if Docker is available
if ! command -v docker &> /dev/null || ! docker --version &> /dev/null; then
    echo "[ERROR] Docker is not available"
    echo "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo "[ERROR] Docker daemon is not running"
    echo "Please start Docker and try again"
    exit 1
fi

echo "[✓] Docker is available and running"
echo

cd test/integration

echo "Setting up testcontainer environment..."
export USE_TESTCONTAINERS=true

echo "Running integration tests with automatic PostgreSQL container..."
go test -v

exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo
    echo "[SUCCESS] Testcontainer integration tests completed successfully!"
else
    echo
    echo "[ERROR] Testcontainer integration tests failed"
fi

exit $exit_code