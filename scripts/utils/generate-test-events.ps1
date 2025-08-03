# Script to generate test events for Contributing Events functionality
Write-Host "Generating test events for Contributing Events..."

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
    # Get existing risk objects and detections
    Write-Host "Getting existing risk objects..."
    $riskObjectsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/objects" -Method GET
    Write-Host "Found $($riskObjectsResponse.Count) risk objects"
    
    Write-Host "Getting existing detections..."
    $detectionsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/detections" -Method GET
    Write-Host "Found $($detectionsResponse.Count) detections"
    
    if ($riskObjectsResponse.Count -gt 0 -and $detectionsResponse.Count -gt 0) {
        # Generate test events for the first few risk objects
        $eventsToCreate = @()
        
        for ($i = 0; $i -lt [Math]::Min(5, $riskObjectsResponse.Count); $i++) {
            $riskObject = $riskObjectsResponse[$i]
            $detection = $detectionsResponse[$i % $detectionsResponse.Count]
            
            # Create multiple events per risk object to build up risk score
            for ($j = 0; $j -lt 3; $j++) {
                $event = @{
                    detection_id = $detection.id
                    entity_id = $riskObject.id
                    timestamp = (Get-Date).AddHours(-($j * 2)).ToString("yyyy-MM-ddTHH:mm:ssZ")
                    raw_data = "Test event data for detection: $($detection.name)"
                    context = @{
                        source_ip = "192.168.1.$($i + 10)"
                        user_agent = "Mozilla/5.0 Test Browser"
                        process_name = "test_process_$j.exe"
                        command_line = "test_command_$j --flag value"
                    } | ConvertTo-Json -Compress
                    risk_points = $detection.risk_points
                    is_false_positive = $false
                }
                
                $eventsToCreate += $event
            }
        }
        
        Write-Host "Creating $($eventsToCreate.Count) test events..."
        
        # Create events one by one
        $createdCount = 0
        foreach ($event in $eventsToCreate) {
            try {
                $jsonBody = $event | ConvertTo-Json -Depth 3
                $response = Invoke-RestMethod -Uri "http://localhost:8080/api/events" -Method POST -Body $jsonBody -ContentType "application/json"
                $createdCount++
                Write-Host "Created event $createdCount with ID: $($response.id)"
            } catch {
                Write-Host "Error creating event: $($_.Exception.Message)"
            }
        }
        
        Write-Host "Successfully created $createdCount events"
        
        # Wait a moment for risk processing
        Start-Sleep -Seconds 2
        
        # Check if any new alerts were generated
        Write-Host "Checking for new risk alerts..."
        $alertsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
        Write-Host "Total alerts now: $($alertsResponse.Count)"
        
        if ($alertsResponse.Count -gt 0) {
            $latestAlert = $alertsResponse[0]
            Write-Host "Testing Contributing Events for alert ID: $($latestAlert.id)"
            
            $eventsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts/$($latestAlert.id)/events" -Method GET
            Write-Host "Contributing events for alert $($latestAlert.id): $($eventsResponse.Count)"
            
            if ($eventsResponse.Count -gt 0) {
                Write-Host "SUCCESS: Contributing Events are now populated!"
                Write-Host "Sample contributing event:"
                $sampleEvent = $eventsResponse[0]
                Write-Host "- Event ID: $($sampleEvent.id)"
                Write-Host "- Detection ID: $($sampleEvent.detection_id)"
                Write-Host "- Risk Points: $($sampleEvent.risk_points)"
                Write-Host "- Timestamp: $($sampleEvent.timestamp)"
            } else {
                Write-Host "No contributing events found for the alert."
            }
        }
        
    } else {
        Write-Host "ERROR: Need existing risk objects and detections to create events"
    }
    
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
} finally {
    # Stop the server
    Write-Host "Stopping server..."
    Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force
}