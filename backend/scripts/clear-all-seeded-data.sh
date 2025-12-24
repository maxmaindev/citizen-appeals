#!/bin/bash

# Script to clear ALL seeded data (appeals + test users)
# This deletes appeals, related data, and test users in one go

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-citizen_appeals}

echo "⚠️  WARNING: This will DELETE ALL seeded data!"
echo "This includes:"
echo "  - All appeals"
echo "  - All comments"
echo "  - All photos"
echo "  - All notifications"
echo "  - All appeal history"
echo "  - Test users (admin@example.com, disp@disp.example, exec@example.com, usr@example.com)"
echo ""
echo "It will NOT delete:"
echo "  - Other users (if any)"
echo "  - Categories"
echo "  - Services"
echo ""
echo "Press Ctrl+C to cancel, or Enter to continue..."
read

# Check if running in Docker or locally
if command -v docker &> /dev/null && docker ps | grep -q citizen_appeals_db; then
    echo "Using Docker container..."
    
    echo "Deleting appeal history..."
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DELETE FROM appeal_history;" > /dev/null 2>&1
    
    echo "Deleting notifications..."
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DELETE FROM notifications;" > /dev/null 2>&1
    
    echo "Deleting comments..."
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DELETE FROM comments;" > /dev/null 2>&1
    
    echo "Deleting photos..."
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DELETE FROM photos;" > /dev/null 2>&1
    
    echo "Deleting appeals..."
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DELETE FROM appeals;" > /dev/null 2>&1
    
    echo "Deleting test users..."
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "
        DELETE FROM users 
        WHERE email IN ('admin@example.com', 'disp@disp.example', 'exec@example.com', 'usr@example.com');
    " > /dev/null 2>&1
    
    echo "Resetting sequences..."
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE appeals_id_seq RESTART WITH 1;" > /dev/null 2>&1
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE comments_id_seq RESTART WITH 1;" > /dev/null 2>&1
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE photos_id_seq RESTART WITH 1;" > /dev/null 2>&1
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE notifications_id_seq RESTART WITH 1;" > /dev/null 2>&1
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE appeal_history_id_seq RESTART WITH 1;" > /dev/null 2>&1
    docker exec -i citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "
        SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 1));
    " > /dev/null 2>&1
else
    echo "Using local PostgreSQL..."
    
    export PGPASSWORD=$DB_PASSWORD
    
    echo "Deleting appeal history..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM appeal_history;" > /dev/null 2>&1
    
    echo "Deleting notifications..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM notifications;" > /dev/null 2>&1
    
    echo "Deleting comments..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM comments;" > /dev/null 2>&1
    
    echo "Deleting photos..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM photos;" > /dev/null 2>&1
    
    echo "Deleting appeals..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM appeals;" > /dev/null 2>&1
    
    echo "Deleting test users..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        DELETE FROM users 
        WHERE email IN ('admin@example.com', 'disp@disp.example', 'exec@example.com', 'usr@example.com');
    " > /dev/null 2>&1
    
    echo "Resetting sequences..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE appeals_id_seq RESTART WITH 1;" > /dev/null 2>&1
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE comments_id_seq RESTART WITH 1;" > /dev/null 2>&1
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE photos_id_seq RESTART WITH 1;" > /dev/null 2>&1
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE notifications_id_seq RESTART WITH 1;" > /dev/null 2>&1
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "ALTER SEQUENCE appeal_history_id_seq RESTART WITH 1;" > /dev/null 2>&1
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 1));
    " > /dev/null 2>&1
fi

echo ""
echo "✅ All seeded data cleared!"
echo ""
echo "Note: Uploaded photo files in backend/uploads/ are NOT deleted by this script."
echo "You may want to manually clean them if needed."

