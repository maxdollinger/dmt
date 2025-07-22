#!/bin/bash

# Script to run migrations manually using the official migrate image
echo "Running database migrations..."

# Check if .env file exists and source it
if [ -f .env ]; then
    source .env
fi

# Set default values if not provided
DATABASE_URL=${DATABASE_URL:-"postgres://user:password@localhost:5432/dmt?sslmode=disable"}

# Run migrations using the official migrate image
docker run --rm \
    -v "$(pwd)/internals/migrations:/migrations" \
    --network dmt_default \
    migrate/migrate \
    -path=/migrations/ \
    -database="$DATABASE_URL" \
    up

echo "Migrations completed!" 