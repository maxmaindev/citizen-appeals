#!/bin/bash

# Apply all migrations to PostgreSQL database

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-citizen_appeals}

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# MIGRATIONS_DIR is relative to backend directory (one level up from scripts)
MIGRATIONS_DIR="$SCRIPT_DIR/../migrations"

echo "Applying migrations to $DB_NAME..."

# Check if running in Docker or locally
if command -v docker &> /dev/null && docker ps | grep -q citizen_appeals_db; then
    echo "Using Docker container..."

    # Get migrations, but exclude seed_admin_user (use create-admin.go script instead)
    for migration in $(ls $MIGRATIONS_DIR/*.sql | grep -v "seed_admin_user" | sort); do
        echo "Applying $(basename $migration)..."
        
        # Extract only the "Up" section from migration file
        # Read file line by line, extract content between "-- +migrate Up" and "-- +migrate Down"
        awk '
            /^--\s*\+migrate\s+Up/ { in_up=1; next }
            /^--\s*\+migrate\s+Down/ { exit }
            in_up { print }
        ' "$migration" | docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME

        if [ $? -eq 0 ]; then
            echo "✓ $(basename $migration) applied successfully"
        else
            echo "✗ Failed to apply $(basename $migration)"
            exit 1
        fi
    done
else
    echo "Using local PostgreSQL..."

    export PGPASSWORD=$DB_PASSWORD

    # Get migrations, but exclude seed_admin_user (use create-admin.go script instead)
    for migration in $(ls $MIGRATIONS_DIR/*.sql | grep -v "seed_admin_user" | sort); do
        echo "Applying $(basename $migration)..."
        
        # Extract only the "Up" section from migration file
        # Read file line by line, extract content between "-- +migrate Up" and "-- +migrate Down"
        awk '
            /^--\s*\+migrate\s+Up/ { in_up=1; next }
            /^--\s*\+migrate\s+Down/ { exit }
            in_up { print }
        ' "$migration" | psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME

        if [ $? -eq 0 ]; then
            echo "✓ $(basename $migration) applied successfully"
        else
            echo "✗ Failed to apply $(basename $migration)"
            exit 1
        fi
    done
fi

echo ""
echo "All migrations applied successfully!"
