# Test script to verify add detection functionality
Write-Host "Testing add detection functionality..."

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
    # Test creating a new detection
    Write-Host "Testing detection creation..."
    $newDetection = @{
        name = "Test Detection - Field Removal"
        description = "Test detection to verify field removal works"
        status = "idea"
        severity = "medium"
        risk_points = 50
        playbook_link = ""
        owner = "Test User"
        risk_object = "Host"
        testing_description = "Test description"
    }
    
    $jsonBody = $newDetection | ConvertTo-Json
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/detections" -Method POST -Body $jsonBody -ContentType "application/json"
    
    Write-Host "Detection created successfully with ID: $($response.id)"
    Write-Host "Name: $($response.name)"
    Write-Host "Status: $($response.status)"
    Write-Host "SUCCESS: Add detection functionality is working correctly!"
    
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
} finally {
    # Stop the server
    Write-Host "Stopping server..."
    Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force
}