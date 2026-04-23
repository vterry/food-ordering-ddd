#!/bin/bash
set -e

# Cleanup any previous runs
echo "Cleaning up previous runs..."
docker compose -f docker-compose.test.yml down -v --remove-orphans

# Build and start services
echo "Starting services..."
docker compose -f docker-compose.test.yml up -d --build

# Wait for services to be ready
echo "Waiting for services to be ready (via healthchecks)..."
# O sleep 20 foi removido porque o test-runner no docker-compose.test.yml 
# agora depende da condição 'service_healthy' de todos os serviços.

# Run tests
echo "Running E2E tests..."
docker compose -f docker-compose.test.yml run --rm test-runner

# Cleanup
echo "Cleaning up..."
docker compose -f docker-compose.test.yml down -v
