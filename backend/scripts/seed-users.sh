#!/bin/bash

# Script to seed users with different roles

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

echo "Running seed_users.go..."
go run scripts/seed_users.go

