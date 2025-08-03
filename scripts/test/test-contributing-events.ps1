# Test script to verify contributing events functionality
Write-Host "Testing contributing events functionality..."

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
    # Test getting risk alerts list
    Write-Host "Testing risk alerts list endpoint..."
    $alertsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
    Write-Host "Risk alerts retrieved successfully. Count: $($alertsResponse.Count)"
    
    if ($alertsResponse.Count -gt 0) {
        # Test getting a specific alert
        $alertId = $alertsResponse[0].id
        Write-Host "Testing alert detail endpoint for ID: $alertId"
        $alertResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts/$alertId" -Method GET
        Write-Host "Alert detail retrieved successfully: Alert #$($alertResponse.id)"
        
        # Test getting contributing events for the alert
        Write-Host "Testing contributing events endpoint for alert ID: $alertId"
        $eventsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts/$alertId/events" -Method GET
        Write-Host "Contributing events retrieved successfully. Count: $($eventsResponse.Count)"
        
        if ($eventsResponse.Count -gt 0) {
            Write-Host "Sample event details:"
            $sampleEvent = $eventsResponse[0]
            Write-Host "- Event ID: $($sampleEvent.id)"
            Write-Host "- Detection ID: $($sampleEvent.detection_id)"
            Write-Host "- Risk Points: $($sampleEvent.risk_points)"
            Write-Host "- Timestamp: $($sampleEvent.timestamp)"
            Write-Host "- Raw Data: $($sampleEvent.raw_data)"
            Write-Host "- Context: $($sampleEvent.context)"
        } else {
            Write-Host "No contributing events found for this alert."
            Write-Host "This might be why the Contributing Events section appears empty."
        }
        
        Write-Host "SUCCESS: API endpoints are working correctly!"
        Write-Host "- Alert list: ✓"
        Write-Host "- Alert detail: ✓"
        Write-Host "- Contributing events: ✓"
        
    } else {
        Write-Host "No alerts found in database to test with."
        Write-Host "You may need to generate some risk alerts first."
    }
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
} finally {
    # Stop the server
    Write-Host "Stopping server..."
    Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force
}