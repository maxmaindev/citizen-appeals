#!/bin/bash

# Script to clear seeded test users
# This will delete only the test users created by seed_users.go:
# - admin@example.com
# - disp@disp.example
# - exec@example.com
# - usr@example.com

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-citizen_appeals}

echo "⚠️  WARNING: This will DELETE test users created by seed_users.go!"
echo "Users to be deleted:"
echo "  - admin@example.com"
echo "  - disp@disp.example"
echo "  - exec@example.com"
echo "  - usr@example.com"
echo ""
echo "Press Ctrl+C to cancel, or Enter to continue..."
read

# Check if running in Docker or locally
if command -v docker &> /dev/null && docker ps | grep -q citizen_appeals_db; then
    echo "Using Docker container..."
    
    echo "Deleting test users..."
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "
        DELETE FROM users 
        WHERE email IN ('admin@example.com', 'disp@disp.example', 'exec@example.com', 'usr@example.com');
    " > /dev/null 2>&1
    
    echo "Resetting user sequence..."
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "
        SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
    " > /dev/null 2>&1
else
    echo "Using local PostgreSQL..."
    
    export PGPASSWORD=$DB_PASSWORD
    
    echo "Deleting test users..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        DELETE FROM users 
        WHERE email IN ('admin@example.com', 'disp@disp.example', 'exec@example.com', 'usr@example.com');
    " > /dev/null 2>&1
    
    echo "Resetting user sequence..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
    " > /dev/null 2>&1
fi

echo ""
echo "✅ Test users deleted!"

