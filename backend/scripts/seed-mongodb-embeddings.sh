#!/bin/bash

# Script to seed MongoDB with service embeddings data
# This script uses the same data as in 005_expand_service_keywords.sql

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

echo "Seeding MongoDB with service embeddings data..."
echo ""

# Check if MongoDB driver is installed
if ! go list -m go.mongodb.org/mongo-driver/mongo &> /dev/null; then
    echo "Installing MongoDB driver..."
    go get go.mongodb.org/mongo-driver/mongo go.mongodb.org/mongo-driver/bson go.mongodb.org/mongo-driver/mongo/options
    if [ $? -ne 0 ]; then
        echo "✗ Failed to install MongoDB driver"
        exit 1
    fi
    echo "✓ MongoDB driver installed"
    echo ""
fi

# Run the seeding script
go run scripts/seed-mongodb-embeddings.go

exit $?

