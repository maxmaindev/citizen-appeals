#!/bin/bash

# Script to create admin user
# Usage: ./create-admin.sh [email] [password]

EMAIL=${1:-"admin@example.com"}
PASSWORD=${2:-"Admin123!"}

echo "Creating admin user..."
echo "Email: $EMAIL"
echo "Password: $PASSWORD"

cd "$(dirname "$0")/.."

go run scripts/create-admin.go -email "$EMAIL" -password "$PASSWORD"

