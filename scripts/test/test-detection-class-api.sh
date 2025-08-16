#!/bin/bash

# Test script for Detection Class API endpoints
# Tests all CRUD operations and relationships

SERVER_URL="${1:-http://localhost:8080}"

echo "Testing Detection Class API Endpoints"
echo "====================================="
echo "Server: $SERVER_URL"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: List all detection classes
echo "Test 1: List all detection classes"
echo "-----------------------------------"
response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/detection-classes")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ GET /api/detection-classes returned 200${NC}"
    echo "Classes found:"
    echo "$body" | jq -r '.[] | "  - \(.name) (ID: \(.id), System: \(.is_system))"'
else
    echo -e "${RED}✗ GET /api/detection-classes failed with status $http_code${NC}"
fi
echo ""

# Test 2: Get a specific class
echo "Test 2: Get specific detection class (ID: 1)"
echo "--------------------------------------------"
response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/detection-classes/1")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ GET /api/detection-classes/1 returned 200${NC}"
    echo "Class details:"
    echo "$body" | jq '.'
else
    echo -e "${RED}✗ GET /api/detection-classes/1 failed with status $http_code${NC}"
fi
echo ""

# Test 3: Create a new custom class
echo "Test 3: Create new custom detection class"
echo "-----------------------------------------"
new_class=$(cat <<EOF
{
    "name": "TestClass",
    "description": "Test custom detection class",
    "color": "#FF6B6B",
    "icon": "test",
    "display_order": 10
}
EOF
)

response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$new_class" \
    -w "\n%{http_code}" \
    "$SERVER_URL/api/detection-classes")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
    echo -e "${GREEN}✓ POST /api/detection-classes returned 201${NC}"
    custom_class_id=$(echo "$body" | jq -r '.id')
    echo "Created class with ID: $custom_class_id"
else
    echo -e "${RED}✗ POST /api/detection-classes failed with status $http_code${NC}"
    echo "Error: $body"
fi
echo ""

# Test 4: Update the custom class
if [ -n "$custom_class_id" ]; then
    echo "Test 4: Update custom detection class"
    echo "-------------------------------------"
    update_data=$(cat <<EOF
{
    "name": "UpdatedTestClass",
    "description": "Updated test description",
    "color": "#00B4D8",
    "icon": "updated",
    "display_order": 15
}
EOF
)
    
    response=$(curl -s -X PUT \
        -H "Content-Type: application/json" \
        -d "$update_data" \
        -w "\n%{http_code}" \
        "$SERVER_URL/api/detection-classes/$custom_class_id")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "204" ]; then
        echo -e "${GREEN}✓ PUT /api/detection-classes/$custom_class_id returned 204${NC}"
        
        # Verify the update
        updated=$(curl -s "$SERVER_URL/api/detection-classes/$custom_class_id" | jq -r '.name')
        echo "Verified updated name: $updated"
    else
        echo -e "${RED}✗ PUT /api/detection-classes/$custom_class_id failed with status $http_code${NC}"
    fi
    echo ""
fi

# Test 5: Try to update a system class (should fail)
echo "Test 5: Try to update system class (should fail)"
echo "------------------------------------------------"
system_update=$(cat <<EOF
{
    "name": "ModifiedAuth",
    "description": "Trying to modify system class"
}
EOF
)

response=$(curl -s -X PUT \
    -H "Content-Type: application/json" \
    -d "$system_update" \
    -w "\n%{http_code}" \
    "$SERVER_URL/api/detection-classes/1")
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "403" ]; then
    echo -e "${GREEN}✓ PUT /api/detection-classes/1 correctly returned 403 (Forbidden)${NC}"
else
    echo -e "${RED}✗ Expected 403 but got $http_code${NC}"
fi
echo ""

# Test 6: Assign class to a detection
echo "Test 6: Assign class to detection"
echo "---------------------------------"
if [ -n "$custom_class_id" ]; then
    # Get first detection
    first_detection_id=$(curl -s "$SERVER_URL/api/detections" | jq -r '.[0].id')
    
    if [ -n "$first_detection_id" ] && [ "$first_detection_id" != "null" ]; then
        # Update detection with class_id
        detection_update=$(cat <<EOF
{
    "class_id": $custom_class_id
}
EOF
)
        
        # Note: This would require updating the detection through the detection API
        echo "Would assign class $custom_class_id to detection $first_detection_id"
        echo "(Detection update endpoint would handle this)"
    else
        echo "No detections found to test with"
    fi
else
    echo "Skipping - no custom class created"
fi
echo ""

# Test 7: List detections by class
echo "Test 7: List detections by class"
echo "--------------------------------"
response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/detection-classes/1/detections")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ GET /api/detection-classes/1/detections returned 200${NC}"
    count=$(echo "$body" | jq '. | length')
    echo "Found $count detections with Auth class"
else
    echo -e "${RED}✗ GET /api/detection-classes/1/detections failed with status $http_code${NC}"
fi
echo ""

# Test 8: Filter detections by class_id
echo "Test 8: Filter detections by class_id parameter"
echo "-----------------------------------------------"
response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/detections?class_id=1")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ GET /api/detections?class_id=1 returned 200${NC}"
    count=$(echo "$body" | jq '. | length')
    echo "Found $count detections with class_id=1"
else
    echo -e "${RED}✗ GET /api/detections?class_id=1 failed with status $http_code${NC}"
fi
echo ""

# Test 9: Delete custom class
if [ -n "$custom_class_id" ]; then
    echo "Test 9: Delete custom detection class"
    echo "-------------------------------------"
    response=$(curl -s -X DELETE -w "\n%{http_code}" "$SERVER_URL/api/detection-classes/$custom_class_id")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "204" ]; then
        echo -e "${GREEN}✓ DELETE /api/detection-classes/$custom_class_id returned 204${NC}"
        
        # Verify deletion
        verify_response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/detection-classes/$custom_class_id")
        verify_code=$(echo "$verify_response" | tail -n1)
        if [ "$verify_code" = "404" ]; then
            echo "Verified: Class successfully deleted"
        fi
    else
        echo -e "${RED}✗ DELETE /api/detection-classes/$custom_class_id failed with status $http_code${NC}"
    fi
    echo ""
fi

# Test 10: Try to delete a system class (should fail)
echo "Test 10: Try to delete system class (should fail)"
echo "-------------------------------------------------"
response=$(curl -s -X DELETE -w "\n%{http_code}" "$SERVER_URL/api/detection-classes/1")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "403" ]; then
    echo -e "${GREEN}✓ DELETE /api/detection-classes/1 correctly returned 403 (Forbidden)${NC}"
    echo "Error message: $body"
else
    echo -e "${RED}✗ Expected 403 but got $http_code${NC}"
fi
echo ""

echo "====================================="
echo "Detection Class API Tests Complete"
echo "====================================="