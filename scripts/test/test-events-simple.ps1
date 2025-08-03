# Simple test to check events and contributing events
Write-Host "Testing events and contributing events..."

Write-Host "Building server..."
$env:CGO_ENABLED=1
go build -o server.exe ./cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build server"
    exit 1
}

Write-Host "Starting server..."
Start-Process -FilePath ".\server.exe" -WindowStyle Hidden
Start-Sleep -Seconds 3

try {
    # Test events endpoint directly
    Write-Host "Testing events endpoint..."
    try {
        $eventsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/events" -Method GET
        Write-Host "Events API response: $($eventsResponse.events.Count) events (total: $($eventsResponse.pagination.total_count))"
        
        if ($eventsResponse.events.Count -gt 0) {
            Write-Host "Sample event:"
            $event = $eventsResponse.events[0]
            Write-Host "- ID: $($event.id)"
            Write-Host "- Entity ID: $($event.entity_id)"
            Write-Host "- Detection ID: $($event.detection_id)"
            Write-Host "- Risk Points: $($event.risk_points)"
            Write-Host "- Timestamp: $($event.timestamp)"
        }
    } catch {
        Write-Host "Events API error: $($_.Exception.Message)"
    }
    
    # Test alerts
    Write-Host ""
    Write-Host "Testing alerts..."
    $alertsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
    Write-Host "Alerts: $($alertsResponse.Count)"
    
    if ($alertsResponse.Count -gt 0) {
        $alert = $alertsResponse[0]
        Write-Host "Testing contributing events for alert $($alert.id)..."
        
        try {
            $contributingEvents = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts/$($alert.id)/events" -Method GET
            Write-Host "Contributing events: $($contributingEvents.Count)"
            
            if ($contributingEvents.Count -gt 0) {
                Write-Host "SUCCESS: Contributing events found!"
                $event = $contributingEvents[0]
                Write-Host "- Event ID: $($event.id)"
                Write-Host "- Detection ID: $($event.detection_id)"
                Write-Host "- Risk Points: $($event.risk_points)"
            } else {
                Write-Host "No contributing events found for alert $($alert.id)"
                Write-Host "Alert entity ID: $($alert.entity_id)"
                Write-Host "Alert triggered at: $($alert.triggered_at)"
            }
        } catch {
            Write-Host "Contributing events API error: $($_.Exception.Message)"
        }
    }
    
} catch {
    Write-Host "General error: $($_.Exception.Message)"
} finally {
    Write-Host "Stopping server..."
    Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force
}