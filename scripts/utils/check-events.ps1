Write-Host "Checking events and alerts in database..."

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

Write-Host "Getting all events..."
$eventsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/events" -Method GET
Write-Host "Total events in database: $($eventsResponse.Count)"

if ($eventsResponse.Count -gt 0) {
    Write-Host "Sample event details:"
    $sampleEvent = $eventsResponse[0]
    Write-Host "- Event ID: $($sampleEvent.id)"
    Write-Host "- Entity ID: $($sampleEvent.entity_id)"
    Write-Host "- Detection ID: $($sampleEvent.detection_id)"
    Write-Host "- Timestamp: $($sampleEvent.timestamp)"
    Write-Host "- Risk Points: $($sampleEvent.risk_points)"
    Write-Host "- Is False Positive: $($sampleEvent.is_false_positive)"
}

Write-Host ""
Write-Host "Getting all risk alerts..."
$alertsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
Write-Host "Total alerts in database: $($alertsResponse.Count)"

if ($alertsResponse.Count -gt 0) {
    Write-Host "Sample alert details:"
    $sampleAlert = $alertsResponse[0]
    Write-Host "- Alert ID: $($sampleAlert.id)"
    Write-Host "- Entity ID: $($sampleAlert.entity_id)"
    Write-Host "- Triggered At: $($sampleAlert.triggered_at)"
    Write-Host "- Total Score: $($sampleAlert.total_score)"
}

Write-Host ""
Write-Host "Getting all risk objects..."
$objectsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/objects" -Method GET
Write-Host "Total risk objects in database: $($objectsResponse.Count)"

if ($objectsResponse.Count -gt 0) {
    Write-Host "Sample risk object details:"
    $sampleObject = $objectsResponse[0]
    Write-Host "- Object ID: $($sampleObject.id)"
    Write-Host "- Entity Type: $($sampleObject.entity_type)"
    Write-Host "- Entity Value: $($sampleObject.entity_value)"
    Write-Host "- Current Score: $($sampleObject.current_score)"
}

Write-Host ""
Write-Host "Stopping server..."
Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force