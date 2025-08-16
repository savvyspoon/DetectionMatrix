#!/bin/bash

# Test script for detection classes feature
# Tests the database schema changes and relationships

DB_PATH="${1:-data/riskmatrix.db}"

echo "Testing Detection Classes Feature"
echo "================================="
echo ""

# Test 1: Verify detection_classes table exists
echo "Test 1: Checking detection_classes table..."
if sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM detection_classes;" > /dev/null 2>&1; then
    echo "✓ detection_classes table exists"
else
    echo "✗ detection_classes table not found"
    exit 1
fi

# Test 2: Verify default classes
echo ""
echo "Test 2: Checking default classes..."
COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM detection_classes WHERE is_system = 1;")
if [ "$COUNT" = "4" ]; then
    echo "✓ Found 4 system classes"
    sqlite3 "$DB_PATH" "SELECT name FROM detection_classes WHERE is_system = 1 ORDER BY display_order;" | while read CLASS; do
        echo "  - $CLASS"
    done
else
    echo "✗ Expected 4 system classes, found $COUNT"
fi

# Test 3: Verify class_id column in detections
echo ""
echo "Test 3: Checking class_id column in detections table..."
if sqlite3 "$DB_PATH" "SELECT class_id FROM detections LIMIT 1;" > /dev/null 2>&1; then
    echo "✓ class_id column exists in detections table"
else
    echo "✗ class_id column not found in detections table"
    exit 1
fi

# Test 4: Test adding a custom class
echo ""
echo "Test 4: Adding custom class..."
sqlite3 "$DB_PATH" "INSERT INTO detection_classes (name, description, color, icon, is_system, display_order) 
                    VALUES ('Custom', 'Test custom class', '#FF5722', 'alert', 0, 5);"
if [ $? -eq 0 ]; then
    echo "✓ Successfully added custom class"
    
    # Verify it was added
    CUSTOM_ID=$(sqlite3 "$DB_PATH" "SELECT id FROM detection_classes WHERE name = 'Custom';")
    echo "  Custom class ID: $CUSTOM_ID"
else
    echo "✗ Failed to add custom class"
fi

# Test 5: Test assigning class to detection
echo ""
echo "Test 5: Assigning class to detection..."
FIRST_DETECTION=$(sqlite3 "$DB_PATH" "SELECT id FROM detections LIMIT 1;")
if [ -n "$FIRST_DETECTION" ]; then
    AUTH_CLASS_ID=$(sqlite3 "$DB_PATH" "SELECT id FROM detection_classes WHERE name = 'Auth';")
    sqlite3 "$DB_PATH" "UPDATE detections SET class_id = $AUTH_CLASS_ID WHERE id = $FIRST_DETECTION;"
    
    if [ $? -eq 0 ]; then
        echo "✓ Successfully assigned Auth class to detection ID $FIRST_DETECTION"
        
        # Verify the assignment
        ASSIGNED_CLASS=$(sqlite3 "$DB_PATH" "SELECT dc.name FROM detections d 
                                              JOIN detection_classes dc ON d.class_id = dc.id 
                                              WHERE d.id = $FIRST_DETECTION;")
        echo "  Verified: Detection has class '$ASSIGNED_CLASS'"
    else
        echo "✗ Failed to assign class to detection"
    fi
else
    echo "⚠ No detections found to test with"
fi

# Test 6: Test that system classes cannot be deleted (at application level)
echo ""
echo "Test 6: Verifying system class protection..."
echo "  Note: Protection is enforced at application level, not database level"
echo "  Database allows deletion but application should prevent it"

# Test 7: Clean up test data
echo ""
echo "Test 7: Cleaning up test data..."
sqlite3 "$DB_PATH" "DELETE FROM detection_classes WHERE name = 'Custom';"
if [ $? -eq 0 ]; then
    echo "✓ Test data cleaned up"
else
    echo "⚠ Could not clean up test data"
fi

echo ""
echo "================================="
echo "Detection Classes Tests Complete"
echo ""

# Show summary
echo "Summary of detection classes:"
sqlite3 "$DB_PATH" "SELECT printf('%-10s %-50s %-10s %-10s', name, description, color, icon) 
                    FROM detection_classes 
                    ORDER BY display_order;" | column -t