Write-Host "Testing contributing events functionality..."

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

Write-Host "Testing risk alerts list endpoint..."
$alertsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
Write-Host "Risk alerts retrieved successfully. Count: $($alertsResponse.Count)"

if ($alertsResponse.Count -gt 0) {
    $alertId = $alertsResponse[0].id
    Write-Host "Testing contributing events endpoint for alert ID: $alertId"
    $eventsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts/$alertId/events" -Method GET
    Write-Host "Contributing events retrieved successfully. Count: $($eventsResponse.Count)"
    
    if ($eventsResponse.Count -gt 0) {
        Write-Host "Sample event details:"
        $sampleEvent = $eventsResponse[0]
        Write-Host "- Event ID: $($sampleEvent.id)"
        Write-Host "- Detection ID: $($sampleEvent.detection_id)"
        Write-Host "- Risk Points: $($sampleEvent.risk_points)"
    } else {
        Write-Host "No contributing events found for this alert."
    }
} else {
    Write-Host "No alerts found in database."
}

Write-Host "Stopping server..."
Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force