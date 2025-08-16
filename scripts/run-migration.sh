#!/bin/bash

# Migration Runner Script for DetectionMatrix
# Usage: ./scripts/run-migration.sh [migration_file] [database_path]

set -e

# Default values
MIGRATION_FILE="${1:-migrations/001_add_detection_classes.sql}"
DB_PATH="${2:-data/riskmatrix.db}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}DetectionMatrix Migration Runner${NC}"
echo "================================"
echo "Migration: $MIGRATION_FILE"
echo "Database: $DB_PATH"
echo ""

# Check if migration file exists
if [ ! -f "$MIGRATION_FILE" ]; then
    echo -e "${RED}Error: Migration file not found: $MIGRATION_FILE${NC}"
    exit 1
fi

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo -e "${YELLOW}Warning: Database not found at $DB_PATH${NC}"
    echo "Creating new database..."
fi

# Create backup of database
if [ -f "$DB_PATH" ]; then
    BACKUP_FILE="${DB_PATH}.backup.$(date +%Y%m%d_%H%M%S)"
    echo "Creating backup: $BACKUP_FILE"
    cp "$DB_PATH" "$BACKUP_FILE"
    echo -e "${GREEN}Backup created successfully${NC}"
fi

# Run migration
echo ""
echo "Running migration..."
if sqlite3 "$DB_PATH" < "$MIGRATION_FILE"; then
    echo -e "${GREEN}Migration completed successfully!${NC}"
    
    # Verify the migration
    echo ""
    echo "Verifying migration..."
    echo "Detection classes in database:"
    sqlite3 "$DB_PATH" "SELECT name, description, is_system FROM detection_classes ORDER BY display_order;"
    
    echo ""
    echo "Table structure:"
    sqlite3 "$DB_PATH" ".schema detection_classes"
else
    echo -e "${RED}Migration failed!${NC}"
    
    # Offer to restore backup
    if [ -f "$BACKUP_FILE" ]; then
        echo ""
        read -p "Would you like to restore the backup? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            mv "$BACKUP_FILE" "$DB_PATH"
            echo -e "${GREEN}Backup restored${NC}"
        fi
    fi
    exit 1
fi

echo ""
echo -e "${GREEN}Migration complete!${NC}"