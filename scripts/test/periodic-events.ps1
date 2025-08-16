# Periodic Event Generator for DetectionMatrix
# Generates test events at regular intervals to test risk scoring system

param(
    [string]$ServerUrl = "http://localhost:8080",
    [int]$IntervalSeconds = 30,
    [int]$EventsPerInterval = 5,
    [int]$MaxIterations = 100
)

Write-Host "Starting Periodic Event Generator" -ForegroundColor Green
Write-Host "Server: $ServerUrl"
Write-Host "Interval: $IntervalSeconds seconds"
Write-Host "Events per interval: $EventsPerInterval"
Write-Host "Max iterations: $MaxIterations"
Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
Write-Host ""

# Risk objects to cycle through
$riskObjects = @(
    @{type="user"; identifier="admin@company.com"},
    @{type="user"; identifier="john.doe@company.com"},
    @{type="user"; identifier="jane.smith@company.com"},
    @{type="host"; identifier="WORKSTATION-01"},
    @{type="host"; identifier="SERVER-DB-01"},
    @{type="host"; identifier="LAPTOP-05"},
    @{type="ip"; identifier="192.168.1.100"},
    @{type="ip"; identifier="10.0.0.50"},
    @{type="ip"; identifier="172.16.0.25"}
)

# Detection IDs to use (adjust based on your detections)
$detectionIds = @(1, 2, 3, 4, 5)

# Severity levels and their weights
$severities = @(
    @{level="low"; weight=10},
    @{level="medium"; weight=5},
    @{level="high"; weight=3},
    @{level="critical"; weight=1}
)

function Get-RandomWeighted {
    param($items)
    $totalWeight = ($items | Measure-Object -Property weight -Sum).Sum
    $random = Get-Random -Minimum 0 -Maximum $totalWeight
    $currentWeight = 0
    
    foreach ($item in $items) {
        $currentWeight += $item.weight
        if ($random -lt $currentWeight) {
            return $item.level
        }
    }
    return $items[-1].level
}

function Send-Event {
    param(
        [string]$detectionId,
        [string]$riskObjectType,
        [string]$riskObjectIdentifier,
        [string]$severity
    )
    
    $timestamp = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ss.fffZ")
    
    $body = @{
        detection_id = [int]$detectionId
        risk_object_type = $riskObjectType
        risk_object_identifier = $riskObjectIdentifier
        severity = $severity
        timestamp = $timestamp
        metadata = @{
            source = "periodic-event-generator"
            iteration = $iteration
            test_run = $true
        }
    } | ConvertTo-Json -Depth 10
    
    try {
        $response = Invoke-RestMethod -Uri "$ServerUrl/api/events" -Method Post -Body $body -ContentType "application/json"
        Write-Host "✓" -NoNewline -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "✗" -NoNewline -ForegroundColor Red
        return $false
    }
}

# Main loop
$iteration = 0
$totalEvents = 0
$successfulEvents = 0

while ($iteration -lt $MaxIterations) {
    $iteration++
    Write-Host "`n[$iteration/$MaxIterations] Sending $EventsPerInterval events: " -NoNewline
    
    for ($i = 0; $i -lt $EventsPerInterval; $i++) {
        # Randomly select parameters
        $riskObject = $riskObjects | Get-Random
        $detectionId = $detectionIds | Get-Random
        $severity = Get-RandomWeighted -items $severities
        
        # Send the event
        $success = Send-Event -detectionId $detectionId `
                             -riskObjectType $riskObject.type `
                             -riskObjectIdentifier $riskObject.identifier `
                             -severity $severity
        
        $totalEvents++
        if ($success) {
            $successfulEvents++
        }
        
        # Small delay between events
        Start-Sleep -Milliseconds 500
    }
    
    # Display statistics
    $successRate = [math]::Round(($successfulEvents / $totalEvents) * 100, 2)
    Write-Host " | Success: $successfulEvents/$totalEvents ($successRate%)" -ForegroundColor Cyan
    
    # Check for risk alerts
    try {
        $alerts = Invoke-RestMethod -Uri "$ServerUrl/api/risk/alerts" -Method Get
        if ($alerts.Count -gt 0) {
            Write-Host "  ⚠ Active risk alerts: $($alerts.Count)" -ForegroundColor Yellow
            foreach ($alert in $alerts | Select-Object -First 3) {
                Write-Host "    - $($alert.risk_object_type): $($alert.risk_object_identifier) (Score: $($alert.risk_score))" -ForegroundColor Yellow
            }
        }
    }
    catch {
        # Ignore alert check errors
    }
    
    # Wait for next interval
    if ($iteration -lt $MaxIterations) {
        Write-Host "  Waiting $IntervalSeconds seconds..." -ForegroundColor Gray
        Start-Sleep -Seconds $IntervalSeconds
    }
}

Write-Host "`n" -NoNewline
Write-Host "="*50 -ForegroundColor Cyan
Write-Host "Event Generation Complete!" -ForegroundColor Green
Write-Host "Total Events Sent: $totalEvents" -ForegroundColor White
Write-Host "Successful Events: $successfulEvents" -ForegroundColor White
Write-Host "Success Rate: $successRate%" -ForegroundColor White
Write-Host "="*50 -ForegroundColor Cyan