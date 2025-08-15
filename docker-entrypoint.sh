#!/bin/sh

# Docker entrypoint script for RiskMatrix
# This script initializes the database and imports MITRE data if needed

set -e

DB_PATH="${DB_PATH:-/app/data/riskmatrix.db}"
CSV_PATH="${CSV_PATH:-/app/data/mitre.csv}"

echo "=== RiskMatrix Container Initialization ==="
echo "Database path: $DB_PATH"
echo "MITRE CSV path: $CSV_PATH"

# Ensure data directory exists
mkdir -p "$(dirname "$DB_PATH")"

# Check if database exists and if MITRE data needs to be imported
IMPORT_NEEDED=false

if [ ! -f "$DB_PATH" ]; then
    echo "Database not found. Will create and import MITRE data."
    IMPORT_NEEDED=true
else
    echo "Database exists. Checking if MITRE data is already imported..."
    
    # Check if MITRE data exists by counting techniques
    MITRE_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM mitre_techniques;" 2>/dev/null || echo "0")
    
    if [ "$MITRE_COUNT" -eq "0" ]; then
        echo "No MITRE techniques found in database. Will import MITRE data."
        IMPORT_NEEDED=true
    else
        echo "Found $MITRE_COUNT MITRE techniques in database. Skipping import."
    fi
fi

# Import MITRE data if needed
if [ "$IMPORT_NEEDED" = "true" ]; then
    if [ -f "$CSV_PATH" ]; then
        echo "=== Importing MITRE ATT&CK Data ==="
        echo "This may take a few moments..."
        
        ./mitre-importer -db "$DB_PATH" -csv "$CSV_PATH"
        
        if [ $? -eq 0 ]; then
            echo "✅ MITRE data import completed successfully"
        else
            echo "❌ MITRE data import failed"
            exit 1
        fi
    else
        echo "⚠️  MITRE CSV file not found at $CSV_PATH"
        echo "   Container will start without MITRE data"
        echo "   You can import it later using: docker exec <container> ./mitre-importer"
    fi
fi

# Verify database is accessible
echo "=== Database Verification ==="
if sqlite3 "$DB_PATH" "SELECT 1;" >/dev/null 2>&1; then
    echo "✅ Database is accessible"
    
    # Show some basic stats
    MITRE_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM mitre_techniques;" 2>/dev/null || echo "0")
    DETECTION_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM detections;" 2>/dev/null || echo "0")
    
    echo "📊 Database Statistics:"
    echo "   MITRE Techniques: $MITRE_COUNT"
    echo "   Detections: $DETECTION_COUNT"
else
    echo "❌ Database verification failed"
    exit 1
fi

echo "=== Starting RiskMatrix Server ==="
echo "Server will be available at http://localhost:8080"

# Execute the main application with all passed arguments
exec "$@"