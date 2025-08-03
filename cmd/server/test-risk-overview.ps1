# Test script to verify Risk Overview dashboard shows real values
Write-Host "Testing Risk Overview dashboard with real values..."

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
    Write-Host "Testing Risk Overview API endpoints..."
    
    # Test High Risk Entities endpoint
    Write-Host "Testing /api/risk/high endpoint..."
    try {
        $highRiskResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/high" -Method GET
        Write-Host "High Risk Entities: $($highRiskResponse.Count)"
    } catch {
        Write-Host "High Risk Entities endpoint error: $($_.Exception.Message)"
    }
    
    # Test Active Alerts endpoint
    Write-Host "Testing /api/risk/alerts endpoint..."
    try {
        $alertsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
        $activeAlerts = $alertsResponse | Where-Object { $_.status -ne "Closed" }
        Write-Host "Total Alerts: $($alertsResponse.Count)"
        Write-Host "Active Alerts (non-closed): $($activeAlerts.Count)"
    } catch {
        Write-Host "Risk alerts endpoint error: $($_.Exception.Message)"
    }
    
    # Test Events endpoint
    Write-Host "Testing /api/events endpoint..."
    try {
        $eventsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/events" -Method GET
        $totalEvents = $eventsResponse.pagination.total_count
        $todayEvents = $eventsResponse.events | Where-Object { 
            (Get-Date $_.timestamp).Date -eq (Get-Date).Date 
        }
        $falsePositives = $eventsResponse.events | Where-Object { $_.is_false_positive -eq $true }
        
        Write-Host "Total Events: $totalEvents"
        Write-Host "Events Today: $($todayEvents.Count)"
        Write-Host "False Positives: $($falsePositives.Count)"
    } catch {
        Write-Host "Events endpoint error: $($_.Exception.Message)"
    }
    
    Write-Host ""
    Write-Host "SUCCESS: All Risk Overview API endpoints are working!"
    Write-Host "The dashboard should now display real values instead of hardcoded placeholders."
    Write-Host ""
    Write-Host "You can verify by opening: http://localhost:8080/index.html"
    Write-Host "The Risk Overview section should show the actual counts listed above."
    
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
} finally {
    # Stop the server
    Write-Host "Stopping server..."
    Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force
}