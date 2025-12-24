#!/bin/bash

# Script to reset database (drop all tables and reapply migrations)

DB_USER=${DB_USER:-postgres}
DB_NAME=${DB_NAME:-citizen_appeals}

echo "⚠️  WARNING: This will DROP ALL TABLES and data!"
echo "Press Ctrl+C to cancel, or Enter to continue..."
read

echo "Dropping all tables..."

# Drop all tables in correct order (to avoid foreign key constraints)
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS appeal_history CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS notifications CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS comments CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS photos CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS appeals CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS services CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS categories CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TABLE IF EXISTS users CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TYPE IF EXISTS appeal_status CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TYPE IF EXISTS notification_type CASCADE;" > /dev/null 2>&1
docker exec citizen_appeals_db psql -U $DB_USER -d $DB_NAME -c "DROP TYPE IF EXISTS user_role CASCADE;" > /dev/null 2>&1

echo "✓ All tables dropped"
echo ""
echo "Now applying migrations..."

# Apply migrations
cd "$(dirname "$0")/.."
bash scripts/apply-migrations.sh

echo ""
echo "✅ Database reset complete!"

