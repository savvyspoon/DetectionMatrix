# Test script to verify detection retrieval functionality
Write-Host "Testing detection retrieval functionality..."

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
    Write-Host "Testing detection list endpoint..."
    $listResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/detections" -Method GET
    Write-Host "Detection list retrieved successfully. Count: $($listResponse.Count)"
    
    if ($listResponse.Count -gt 0) {
        # Test getting a specific detection
        $detectionId = $listResponse[0].id
        Write-Host "Testing detection detail endpoint for ID: $detectionId"
        $detailResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/detections/$detectionId" -Method GET
        Write-Host "Detection detail retrieved successfully: $($detailResponse.name)"
        Write-Host "SUCCESS: Detection retrieval is working correctly!"
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