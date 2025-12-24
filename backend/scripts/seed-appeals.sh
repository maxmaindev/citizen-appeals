#!/bin/bash

# Script to seed appeals
# Usage: ./seed-appeals.sh [count]
# Example: ./seed-appeals.sh 50

COUNT=${1:-50}

if ! [[ "$COUNT" =~ ^[0-9]+$ ]] || [ "$COUNT" -le 0 ]; then
    echo "Error: Count must be a positive number"
    echo "Usage: ./seed-appeals.sh [count]"
    echo "Example: ./seed-appeals.sh 50"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

echo "Seeding $COUNT appeals..."
go run scripts/seed_appeals.go "$COUNT"

