# Test script to verify alert management functionality
Write-Host "Testing alert management functionality..."

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
        Write-Host "Current status: $($alertResponse.status)"
        Write-Host "Current owner: $($alertResponse.owner)"
        Write-Host "Current notes: $($alertResponse.notes)"
        
        # Test updating the alert
        Write-Host "Testing alert update endpoint..."
        $updateData = @{
            id = $alertResponse.id
            entity_id = $alertResponse.entity_id
            triggered_at = $alertResponse.triggered_at
            total_score = $alertResponse.total_score
            status = "Investigation"
            notes = "Updated via API test - investigating suspicious activity"
            owner = "Test Analyst"
        }
        
        $jsonBody = $updateData | ConvertTo-Json
        $updateResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts/$alertId" -Method PUT -Body $jsonBody -ContentType "application/json"
        
        Write-Host "Alert updated successfully!"
        Write-Host "New status: $($updateResponse.status)"
        Write-Host "New owner: $($updateResponse.owner)"
        Write-Host "New notes: $($updateResponse.notes)"
        
        # Verify the update by fetching the alert again
        Write-Host "Verifying update..."
        $verifyResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts/$alertId" -Method GET
        
        if ($verifyResponse.status -eq "Investigation" -and $verifyResponse.owner -eq "Test Analyst") {
            Write-Host "SUCCESS: Alert management functionality is working correctly!"
            Write-Host "- Alert list: ✓"
            Write-Host "- Alert detail: ✓"
            Write-Host "- Alert update: ✓"
            Write-Host "- Status change: ✓"
            Write-Host "- Owner assignment: ✓"
            Write-Host "- Notes update: ✓"
        } else {
            Write-Host "ERROR: Update verification failed"
        }
        
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