# Test script to verify risk alert generation when new events are added
Write-Host "Testing risk alert generation functionality..."

Write-Host "Building server..."
$env:CGO_ENABLED=1
go build -o server.exe ./cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build server"
    exit 1
}

# Start the server in background
Write-Host "Starting server..."
Start-Process -FilePath ".\server.exe" -WindowStyle Hidden

# Wait for server to start
Start-Sleep -Seconds 3

try {
    # First, check current state
    Write-Host "=== INITIAL STATE ==="
    
    # Get current alerts count
    $initialAlertsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
    Write-Host "Initial alerts count: $($initialAlertsResponse.Count)"
    
    # Get current risk objects
    $riskObjectsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/objects" -Method GET
    Write-Host "Current risk objects count: $($riskObjectsResponse.Count)"
    
    if ($riskObjectsResponse.Count -gt 0) {
        $testEntity = $riskObjectsResponse[0]
        Write-Host "Test entity: $($testEntity.entity_type) '$($testEntity.entity_value)' (current score: $($testEntity.current_score))"
    }
    
    Write-Host ""
    Write-Host "=== TESTING EVENT CREATION ==="
    
    # Test 1: Try to create an event with missing RiskObject field (current API behavior)
    Write-Host "Test 1: Creating event without RiskObject field..."
    $eventWithoutRiskObject = @{
        detection_id = 1
        entity_id = 1
        timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
        raw_data = "Test event without RiskObject field"
        risk_points = 60  # Should exceed threshold of 50
        is_false_positive = $false
    }
    
    try {
        $jsonBody = $eventWithoutRiskObject | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/events" -Method POST -Body $jsonBody -ContentType "application/json"
        Write-Host "✓ Event created successfully with ID: $($response.id)"
    } catch {
        Write-Host "✗ Event creation failed: $($_.Exception.Message)"
    }
    
    # Test 2: Try to create an event with RiskObject field populated
    Write-Host "Test 2: Creating event with RiskObject field..."
    $eventWithRiskObject = @{
        detection_id = 1
        entity_id = 1
        timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
        raw_data = "Test event with RiskObject field"
        risk_points = 60  # Should exceed threshold of 50
        is_false_positive = $false
        risk_object = @{
            entity_type = "user"
            entity_value = "test.user"
        }
    }
    
    try {
        $jsonBody = $eventWithRiskObject | ConvertTo-Json -Depth 3
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/events" -Method POST -Body $jsonBody -ContentType "application/json"
        Write-Host "✓ Event created successfully with ID: $($response.id)"
    } catch {
        Write-Host "✗ Event creation failed: $($_.Exception.Message)"
    }
    
    # Wait a moment for processing
    Start-Sleep -Seconds 2
    
    Write-Host ""
    Write-Host "=== CHECKING RESULTS ==="
    
    # Check if any new alerts were generated
    $finalAlertsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
    Write-Host "Final alerts count: $($finalAlertsResponse.Count)"
    
    $newAlertsCount = $finalAlertsResponse.Count - $initialAlertsResponse.Count
    Write-Host "New alerts generated: $newAlertsCount"
    
    if ($newAlertsCount -gt 0) {
        Write-Host "✓ SUCCESS: Risk alerts were generated!"
        $latestAlert = $finalAlertsResponse[0]
        Write-Host "Latest alert: ID $($latestAlert.id), Score: $($latestAlert.total_score), Entity: $($latestAlert.entity_id)"
    } else {
        Write-Host "✗ ISSUE: No new risk alerts were generated despite adding high-risk events"
        Write-Host "This suggests the alert generation mechanism is not working properly"
    }
    
    # Check updated risk objects
    $updatedRiskObjectsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/objects" -Method GET
    Write-Host "Updated risk objects count: $($updatedRiskObjectsResponse.Count)"
    
    # Check if any risk scores were updated
    $highScoreObjects = $updatedRiskObjectsResponse | Where-Object { $_.current_score -ge 50 }
    Write-Host "Risk objects with score >= 50: $($highScoreObjects.Count)"
    
    if ($highScoreObjects.Count -gt 0) {
        Write-Host "High-risk objects found:"
        foreach ($obj in $highScoreObjects) {
            Write-Host "  - $($obj.entity_type) '$($obj.entity_value)': $($obj.current_score)"
        }
    }
    
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
} finally {
    # Stop the server
    Write-Host ""
    Write-Host "Stopping server..."
    Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force
}