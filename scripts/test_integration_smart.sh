#!/bin/bash

echo "Smart Integration Test Runner"
echo "============================"
echo

echo "Checking available options for integration testing..."
echo

# Check if Docker is available
if command -v docker &> /dev/null && docker --version &> /dev/null; then
    echo "[✓] Docker is available"
    DOCKER_AVAILABLE=true
else
    echo "[✗] Docker is not available"
    DOCKER_AVAILABLE=false
fi

# Check if local PostgreSQL is available on default port
echo "Testing local PostgreSQL connection..."
if command -v netstat &> /dev/null && netstat -an 2>/dev/null | grep -q ":5432"; then
    echo "[✓] PostgreSQL appears to be running on port 5432"
    LOCAL_PG_AVAILABLE=true
elif command -v ss &> /dev/null && ss -an 2>/dev/null | grep -q ":5432"; then
    echo "[✓] PostgreSQL appears to be running on port 5432"
    LOCAL_PG_AVAILABLE=true
else
    echo "[✗] No PostgreSQL detected on port 5432"
    LOCAL_PG_AVAILABLE=false
fi

echo
echo "Choosing best testing approach..."
echo

# Prioritize testcontainers (most isolated), then Docker Compose, then local DB
if [ "$DOCKER_AVAILABLE" = true ]; then
    echo "[SELECTED] Using testcontainers approach (most isolated)"
    echo
    ./scripts/test_integration_testcontainers.sh
    
    if [ $? -eq 0 ]; then
        echo
        echo "[SUCCESS] Testcontainer tests completed successfully!"
        exit 0
    else
        echo
        echo "[FALLBACK] Testcontainer tests failed, trying Docker Compose..."
        # Fall through to docker_compose
    fi
fi

# Docker Compose fallback
if [ "$DOCKER_AVAILABLE" = true ]; then
    echo "[SELECTED] Using Docker Compose approach"
    echo
    ./scripts/test_integration_docker.sh
    exit $?
fi

# Local DB fallback
if [ "$LOCAL_PG_AVAILABLE" = true ]; then
    echo "[SELECTED] Using local PostgreSQL database"
    echo "Please ensure your local PostgreSQL has a 'test_db' database"
    echo
    ./scripts/test_integration.sh
    exit $?
fi

# No options available
echo "[ERROR] No testing options available!"
echo
echo "Please install Docker OR set up a local PostgreSQL server."
echo
echo "For Docker: https://docs.docker.com/get-docker/"
echo "For PostgreSQL: https://www.postgresql.org/download/"
exit 1