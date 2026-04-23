#!/bin/bash
set -e

# Cleanup any previous runs
echo "Cleaning up previous runs..."
docker compose -f docker-compose.test.yml down -v --remove-orphans

# Build and start services
echo "Starting services..."
docker compose -f docker-compose.test.yml up -d --build

# Wait for services to be ready
echo "Waiting for services to be ready..."
# Wait more for RabbitMQ and MySQL to fully initialize and apps to connect/migrate
sleep 20

# Run tests
echo "Running E2E tests..."
docker compose -f docker-compose.test.yml run --rm test-runner

# Cleanup
echo "Cleaning up..."
docker compose -f docker-compose.test.yml down -v
