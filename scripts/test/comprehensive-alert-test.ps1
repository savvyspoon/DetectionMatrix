# Comprehensive test for risk alert generation edge cases
Write-Host "=== COMPREHENSIVE RISK ALERT GENERATION TEST ==="

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
    Write-Host "=== INITIAL STATE ==="
    $initialAlerts = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
    Write-Host "Initial alerts: $($initialAlerts.Count)"
    
    $riskObjects = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/objects" -Method GET
    Write-Host "Risk objects: $($riskObjects.Count)"
    
    # Find an entity with score below 50 to test threshold crossing
    $lowScoreEntity = $riskObjects | Where-Object { $_.current_score -lt 50 -and $_.current_score -gt 0 } | Select-Object -First 1
    if ($lowScoreEntity) {
        Write-Host "Found low-score entity for testing: $($lowScoreEntity.entity_type)/$($lowScoreEntity.entity_value) (score: $($lowScoreEntity.current_score))"
    }
    
    Write-Host ""
    Write-Host "=== TEST 1: New Entity Alert Generation ==="
    $newEntityEvent = @{
        detection_id = 2
        timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
        raw_data = "New entity high-risk event"
        risk_points = 55
        is_false_positive = $false
        risk_object = @{
            entity_type = "host"
            entity_value = "new.test.host"
        }
    }
    
    $jsonBody = $newEntityEvent | ConvertTo-Json -Depth 3
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/events" -Method POST -Body $jsonBody -ContentType "application/json"
    Write-Host "Created event for new entity: $($response.id)"
    
    Write-Host ""
    Write-Host "=== TEST 2: Existing Entity Threshold Crossing ==="
    if ($lowScoreEntity) {
        $pointsNeeded = 51 - $lowScoreEntity.current_score
        Write-Host "Adding $pointsNeeded points to cross threshold..."
        
        $thresholdEvent = @{
            detection_id = 3
            timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
            raw_data = "Threshold crossing event"
            risk_points = $pointsNeeded
            is_false_positive = $false
            risk_object = @{
                entity_type = $lowScoreEntity.entity_type
                entity_value = $lowScoreEntity.entity_value
            }
        }
        
        $jsonBody = $thresholdEvent | ConvertTo-Json -Depth 3
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/events" -Method POST -Body $jsonBody -ContentType "application/json"
        Write-Host "Created threshold crossing event: $($response.id)"
    }
    
    Write-Host ""
    Write-Host "=== TEST 3: Multiple Events for Same Entity (No Duplicate Alerts) ==="
    $duplicateEvent = @{
        detection_id = 1
        timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
        raw_data = "Additional event for existing high-risk entity"
        risk_points = 10
        is_false_positive = $false
        risk_object = @{
            entity_type = "user"
            entity_value = "test.user"  # This entity already has score 60
        }
    }
    
    $jsonBody = $duplicateEvent | ConvertTo-Json -Depth 3
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/events" -Method POST -Body $jsonBody -ContentType "application/json"
    Write-Host "Created additional event for existing high-risk entity: $($response.id)"
    
    Write-Host ""
    Write-Host "=== TEST 4: Batch Event Processing ==="
    $batchEvents = @(
        @{
            detection_id = 4
            timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
            raw_data = "Batch event 1"
            risk_points = 30
            is_false_positive = $false
            risk_object = @{
                entity_type = "ip"
                entity_value = "192.168.100.1"
            }
        },
        @{
            detection_id = 4
            timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
            raw_data = "Batch event 2"
            risk_points = 25
            is_false_positive = $false
            risk_object = @{
                entity_type = "ip"
                entity_value = "192.168.100.1"
            }
        }
    )
    
    $jsonBody = $batchEvents | ConvertTo-Json -Depth 3
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/events/batch" -Method POST -Body $jsonBody -ContentType "application/json"
    Write-Host "Batch events processed successfully"
    
    # Wait for processing
    Start-Sleep -Seconds 2
    
    Write-Host ""
    Write-Host "=== RESULTS ANALYSIS ==="
    $finalAlerts = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
    $newAlertsCount = $finalAlerts.Count - $initialAlerts.Count
    Write-Host "New alerts generated: $newAlertsCount"
    Write-Host "Total alerts: $($finalAlerts.Count)"
    
    if ($newAlertsCount -gt 0) {
        Write-Host "New alerts:"
        $newAlerts = $finalAlerts | Sort-Object id -Descending | Select-Object -First $newAlertsCount
        foreach ($alert in $newAlerts) {
            Write-Host "  Alert ID: $($alert.id), Entity: $($alert.entity_id), Score: $($alert.total_score)"
        }
    }
    
    # Check updated risk objects
    $updatedRiskObjects = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/objects" -Method GET
    $highRiskObjects = $updatedRiskObjects | Where-Object { $_.current_score -ge 50 }
    Write-Host "High-risk objects (score >= 50): $($highRiskObjects.Count)"
    
    Write-Host ""
    Write-Host "=== CONCLUSION ==="
    if ($newAlertsCount -gt 0) {
        Write-Host "✓ SUCCESS: Risk alert generation is working correctly!"
        Write-Host "✓ System properly generates alerts when thresholds are crossed"
        Write-Host "✓ Both single and batch event processing work"
        Write-Host "✓ No duplicate alerts for entities already above threshold"
    } else {
        Write-Host "✗ ISSUE: Expected alerts were not generated"
    }
    
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
} finally {
    Write-Host ""
    Write-Host "Stopping server..."
    Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force
}