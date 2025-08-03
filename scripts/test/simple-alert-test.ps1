# Simple test to generate one risk alert

$baseUrl = "http://localhost:8080"

# Single event with enough points to cross threshold (50+)
$event = @{
    detection_id = 1
    risk_object = @{
        entity_type = "user"
        entity_value = "test.user"
    }
    raw_data = "High-risk security event detected"
    risk_points = 60
}

Write-Host "Sending single high-risk event to generate alert..."
Write-Host "Entity: $($event.risk_object.entity_value)"
Write-Host "Risk Points: $($event.risk_points) (threshold is 50)"

$json = $event | ConvertTo-Json -Depth 3
Write-Host "`nJSON payload:"
Write-Host $json

try {
    $response = Invoke-WebRequest -Uri "$baseUrl/api/events" -Method POST -Body $json -ContentType "application/json"
    Write-Host "`nEvent sent successfully!"
    Write-Host "Status: $($response.StatusCode)"
    Write-Host "Response: $($response.Content)"
}
catch {
    Write-Host "`nError sending event:"
    Write-Host "  $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorContent = $reader.ReadToEnd()
        Write-Host "  Error details: $errorContent"
    }
}

Write-Host "`nChecking for alerts..."
Start-Sleep -Seconds 2

try {
    $alertsResponse = Invoke-WebRequest -Uri "$baseUrl/api/risk/alerts" -Method GET
    $alerts = $alertsResponse.Content | ConvertFrom-Json
    
    if ($alerts -and $alerts.Count -gt 0) {
        Write-Host "SUCCESS: $($alerts.Count) risk alerts found!"
        foreach ($alert in $alerts) {
            Write-Host "  Alert ID: $($alert.id)"
            Write-Host "  Entity ID: $($alert.entity_id)"
            Write-Host "  Score: $($alert.total_score)"
            Write-Host "  Triggered: $($alert.triggered_at)"
            Write-Host ""
        }
    } else {
        Write-Host "No alerts generated."
    }
}
catch {
    Write-Host "Error checking alerts: $($_.Exception.Message)"
}

Write-Host "`nChecking risk objects..."
try {
    $objectsResponse = Invoke-WebRequest -Uri "$baseUrl/api/risk/objects" -Method GET
    $objects = $objectsResponse.Content | ConvertFrom-Json
    
    Write-Host "Risk objects:"
    foreach ($obj in $objects) {
        if ($obj.entity_value -eq "test.user") {
            Write-Host "  >>> $($obj.entity_type)/$($obj.entity_value): Score $($obj.current_score) <<<"
        } else {
            Write-Host "  $($obj.entity_type)/$($obj.entity_value): Score $($obj.current_score)"
        }
    }
}
catch {
    Write-Host "Error checking risk objects: $($_.Exception.Message)"
}