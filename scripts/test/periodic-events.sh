#!/bin/bash

# Periodic Event Generator for DetectionMatrix
# Generates test events at regular intervals to test risk scoring system

SERVER_URL="${1:-http://localhost:8080}"
INTERVAL_SECONDS="${2:-30}"
EVENTS_PER_INTERVAL="${3:-5}"
MAX_ITERATIONS="${4:-100}"

echo -e "\033[32mStarting Periodic Event Generator\033[0m"
echo "Server: $SERVER_URL"
echo "Interval: $INTERVAL_SECONDS seconds"
echo "Events per interval: $EVENTS_PER_INTERVAL"
echo "Max iterations: $MAX_ITERATIONS"
echo -e "\033[33mPress Ctrl+C to stop\033[0m"
echo ""

# Risk objects to cycle through
RISK_OBJECTS=(
    "user:admin@company.com"
    "user:john.doe@company.com"
    "user:jane.smith@company.com"
    "host:WORKSTATION-01"
    "host:SERVER-DB-01"
    "host:LAPTOP-05"
    "ip:192.168.1.100"
    "ip:10.0.0.50"
    "ip:172.16.0.25"
)

# Detection IDs to use
DETECTION_IDS=(1 2 3 4 5)

# Severity levels with weights for random selection
SEVERITIES=("low:10" "medium:5" "high:3" "critical:1")

get_random_weighted_severity() {
    local total_weight=0
    for item in "${SEVERITIES[@]}"; do
        weight="${item#*:}"
        total_weight=$((total_weight + weight))
    done
    
    local random=$((RANDOM % total_weight))
    local current_weight=0
    
    for item in "${SEVERITIES[@]}"; do
        severity="${item%%:*}"
        weight="${item#*:}"
        current_weight=$((current_weight + weight))
        if [ $random -lt $current_weight ]; then
            echo "$severity"
            return
        fi
    done
    echo "low"
}

send_event() {
    local detection_id=$1
    local risk_object_type=$2
    local risk_object_identifier=$3
    local severity=$4
    local iteration=$5
    
    # Generate timestamp - macOS date doesn't support milliseconds
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS - use Python for proper timestamp
        local timestamp=$(python3 -c "import datetime; print(datetime.datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%S.%f')[:-3] + 'Z')" 2>/dev/null)
    else
        # Linux - use date with nanoseconds
        local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    fi
    
    # Calculate risk points based on severity
    local risk_points=5
    case "$severity" in
        "low") risk_points=5 ;;
        "medium") risk_points=10 ;;
        "high") risk_points=15 ;;
        "critical") risk_points=20 ;;
    esac
    
    local json_body=$(cat <<EOF
{
    "detection_id": $detection_id,
    "risk_object": {
        "entity_type": "$risk_object_type",
        "entity_value": "$risk_object_identifier"
    },
    "risk_points": $risk_points,
    "timestamp": "$timestamp",
    "raw_data": "{\"source\": \"periodic-event-generator\", \"iteration\": $iteration, \"severity\": \"$severity\"}",
    "context": "{\"test_run\": true, \"generator_version\": \"1.0\"}",
    "is_false_positive": false
}
EOF
)
    
    # Debug: show the request being sent
    if [ "${DEBUG:-0}" == "1" ]; then
        echo ""
        echo "Sending request:"
        echo "$json_body" | jq .
    fi
    
    # Send the request and capture response
    response=$(curl -s -X POST "$SERVER_URL/api/events" \
        -H "Content-Type: application/json" \
        -d "$json_body" -w "\n%{http_code}")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" == "201" ] || [ "$http_code" == "200" ]; then
        echo -n -e "\033[32m✓\033[0m"
        return 0
    else
        echo -n -e "\033[31m✗($http_code)\033[0m"
        if [ "${DEBUG:-0}" == "1" ]; then
            echo " Error: $body"
        fi
        return 1
    fi
}

# Main loop
iteration=0
total_events=0
successful_events=0

while [ $iteration -lt $MAX_ITERATIONS ]; do
    iteration=$((iteration + 1))
    echo -n -e "\n[$iteration/$MAX_ITERATIONS] Sending $EVENTS_PER_INTERVAL events: "
    
    for ((i=0; i<$EVENTS_PER_INTERVAL; i++)); do
        # Randomly select parameters
        risk_object="${RISK_OBJECTS[$RANDOM % ${#RISK_OBJECTS[@]}]}"
        risk_object_type="${risk_object%%:*}"
        risk_object_identifier="${risk_object#*:}"
        detection_id="${DETECTION_IDS[$RANDOM % ${#DETECTION_IDS[@]}]}"
        severity=$(get_random_weighted_severity)
        
        # Send the event
        if send_event "$detection_id" "$risk_object_type" "$risk_object_identifier" "$severity" "$iteration"; then
            successful_events=$((successful_events + 1))
        fi
        
        total_events=$((total_events + 1))
        
        # Small delay between events
        sleep 0.5
    done
    
    # Display statistics
    if [ $total_events -gt 0 ]; then
        success_rate=$(awk "BEGIN {printf \"%.2f\", ($successful_events / $total_events) * 100}")
        echo -e " | \033[36mSuccess: $successful_events/$total_events ($success_rate%)\033[0m"
    fi
    
    # Check for risk alerts
    alerts_response=$(curl -s "$SERVER_URL/api/risk/alerts" 2>/dev/null)
    if [ $? -eq 0 ] && [ -n "$alerts_response" ]; then
        alert_count=$(echo "$alerts_response" | jq '. | length' 2>/dev/null || echo "0")
        if [ "$alert_count" -gt "0" ]; then
            echo -e "  \033[33m⚠ Active risk alerts: $alert_count\033[0m"
            echo "$alerts_response" | jq -r '.[:3][] | "    - \(.risk_object_type): \(.risk_object_identifier) (Score: \(.risk_score))"' 2>/dev/null | while read line; do
                echo -e "  \033[33m$line\033[0m"
            done
        fi
    fi
    
    # Wait for next interval
    if [ $iteration -lt $MAX_ITERATIONS ]; then
        echo -e "  \033[90mWaiting $INTERVAL_SECONDS seconds...\033[0m"
        sleep $INTERVAL_SECONDS
    fi
done

echo ""
echo -e "\033[36m==================================================\033[0m"
echo -e "\033[32mEvent Generation Complete!\033[0m"
echo "Total Events Sent: $total_events"
echo "Successful Events: $successful_events"
if [ $total_events -gt 0 ]; then
    success_rate=$(awk "BEGIN {printf \"%.2f\", ($successful_events / $total_events) * 100}")
    echo "Success Rate: $success_rate%"
fi
echo -e "\033[36m==================================================\033[0m"