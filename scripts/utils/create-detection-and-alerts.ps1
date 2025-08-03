# Script to create a detection and then generate events to trigger alerts

$baseUrl = "http://localhost:8080"

# First, create a detection
$detection = @{
    name = "Test Security Detection"
    description = "Test detection for generating risk alerts"
    status = "production"
    severity = "high"
    risk_points = 0
    owner = "test"
    risk_object = "User"
}

Write-Host "Creating test detection..."
$detectionJson = $detection | ConvertTo-Json -Depth 3

try {
    $detectionResponse = Invoke-WebRequest -Uri "$baseUrl/api/detections" -Method POST -Body $detectionJson -ContentType "application/json"
    $createdDetection = $detectionResponse.Content | ConvertFrom-Json
    $detectionId = $createdDetection.id
    Write-Host "Detection created successfully with ID: $detectionId"
}
catch {
    Write-Host "Error creating detection: $($_.Exception.Message)"
    exit 1
}

# Now create events using the valid detection ID
$events = @(
    @{
        detection_id = $detectionId
        risk_object = @{
            entity_type = "user"
            entity_value = "alert.user1"
        }
        raw_data = "Suspicious activity detected - first event"
        risk_points = 30
    },
    @{
        detection_id = $detectionId
        risk_object = @{
            entity_type = "user"
            entity_value = "alert.user1"
        }
        raw_data = "Suspicious activity detected - second event"
        risk_points = 25
    },
    @{
        detection_id = $detectionId
        risk_object = @{
            entity_type = "host"
            entity_value = "alert.host1"
        }
        raw_data = "Malware detected on host"
        risk_points = 60
    },
    @{
        detection_id = $detectionId
        risk_object = @{
            entity_type = "ip"
            entity_value = "192.0.2.100"
        }
        raw_data = "Attack from external IP"
        risk_points = 55
    }
)

Write-Host "`nSending events to generate alerts..."
Write-Host "Threshold is 50 points - some entities should cross it."

foreach ($event in $events) {
    $json = $event | ConvertTo-Json -Depth 3
    Write-Host "Sending event for $($event.risk_object.entity_value): $($event.raw_data) (+$($event.risk_points) points)"
    
    try {
        $response = Invoke-WebRequest -Uri "$baseUrl/api/events" -Method POST -Body $json -ContentType "application/json"
        Write-Host "  Status: $($response.StatusCode)"
    }
    catch {
        Write-Host "  Error: $($_.Exception.Message)"
    }
    
    Start-Sleep -Milliseconds 500
}

Write-Host "`nChecking for generated alerts..."
Start-Sleep -Seconds 2

try {
    $alertsResponse = Invoke-WebRequest -Uri "$baseUrl/api/risk/alerts" -Method GET
    $alerts = $alertsResponse.Content | ConvertFrom-Json
    
    if ($alerts -and $alerts.Count -gt 0) {
        Write-Host "SUCCESS: $($alerts.Count) risk alerts generated!"
        foreach ($alert in $alerts) {
            Write-Host "  Alert ID: $($alert.id)"
            Write-Host "  Entity ID: $($alert.entity_id)"
            Write-Host "  Score: $($alert.total_score)"
            Write-Host "  Triggered: $($alert.triggered_at)"
            Write-Host ""
        }
    } else {
        Write-Host "No alerts generated."
    }
}
catch {
    Write-Host "Error checking alerts: $($_.Exception.Message)"
}

Write-Host "`nChecking risk objects..."
try {
    $objectsResponse = Invoke-WebRequest -Uri "$baseUrl/api/risk/objects" -Method GET
    $objects = $objectsResponse.Content | ConvertFrom-Json
    
    Write-Host "New risk objects:"
    foreach ($obj in $objects) {
        if ($obj.entity_value -like "alert.*" -or $obj.entity_value -eq "192.0.2.100") {
            Write-Host "  >>> $($obj.entity_type)/$($obj.entity_value): Score $($obj.current_score) <<<"
        }
    }
}
catch {
    Write-Host "Error checking risk objects: $($_.Exception.Message)"
}