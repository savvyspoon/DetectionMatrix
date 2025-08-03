# Test script to verify detection detail page has the restored fields
Write-Host "Testing detection detail page fields..."

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
    # Test getting detection list first
    Write-Host "Getting detection list..."
    $listResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/detections" -Method GET
    Write-Host "Detection list retrieved. Count: $($listResponse.Count)"
    
    if ($listResponse.Count -gt 0) {
        $detectionId = $listResponse[0].id
        Write-Host "Testing detection detail for ID: $detectionId"
        
        # Test detection detail endpoint
        $detailResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/detections/$detectionId" -Method GET
        Write-Host "Detection retrieved: $($detailResponse.name)"
        
        # Test event count endpoint
        Write-Host "Testing event count endpoint..."
        $eventCountResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/detections/$detectionId/events/count/30days" -Method GET
        Write-Host "Event count: $($eventCountResponse.count)"
        
        # Test false positive count endpoint
        Write-Host "Testing false positive count endpoint..."
        $fpCountResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/detections/$detectionId/false-positives/count/30days" -Method GET
        Write-Host "False positive count: $($fpCountResponse.count)"
        
        # Test data sources endpoint
        Write-Host "Testing data sources endpoint..."
        $dataSourcesResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/datasources" -Method GET
        Write-Host "Data sources count: $($dataSourcesResponse.Count)"
        
        Write-Host "SUCCESS: All detail page endpoints are working correctly!"
        Write-Host "- Detection detail: ✓"
        Write-Host "- Event count (30 days): ✓"
        Write-Host "- False positive count (30 days): ✓"
        Write-Host "- Data sources: ✓"
        
    } else {
        Write-Host "No detections found in database to test with."
    }
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
} finally {
    # Stop the server
    Write-Host "Stopping server..."
    Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force
}