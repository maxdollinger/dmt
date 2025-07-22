#!/bin/bash

# Script to create new migration files using the official migrate image
if [ -z "$1" ]; then
    echo "Usage: $0 <migration_name>"
    echo "Example: $0 add_users_table"
    exit 1
fi

MIGRATION_NAME=$1

echo "Creating migration: $MIGRATION_NAME"

# Create migration files using the official migrate image
docker run --rm \
    -v "$(pwd)/internals/migrations:/migrations" \
    migrate/migrate \
    create -ext sql -dir /migrations/ -seq "$MIGRATION_NAME"

echo "Migration files created in internals/migrations/ directory" 