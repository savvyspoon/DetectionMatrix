# PowerShell script to generate events that will trigger risk alerts

$baseUrl = "http://localhost:8080"

# Define test events with high risk points to trigger alerts
$events = @(
    @{
        detection_id = 1
        risk_object = @{
            entity_type = "user"
            entity_value = "alice.admin"
        }
        raw_data = "Suspicious admin activity detected"
        risk_points = 30
    },
    @{
        detection_id = 2
        risk_object = @{
            entity_type = "user"
            entity_value = "alice.admin"
        }
        raw_data = "Multiple privilege escalations"
        risk_points = 25
    },
    @{
        detection_id = 3
        risk_object = @{
            entity_type = "host"
            entity_value = "critical-server"
        }
        raw_data = "Malware detected on critical server"
        risk_points = 40
    },
    @{
        detection_id = 1
        risk_object = @{
            entity_type = "host"
            entity_value = "critical-server"
        }
        raw_data = "Unauthorized access attempt"
        risk_points = 20
    },
    @{
        detection_id = 4
        risk_object = @{
            entity_type = "ip"
            entity_value = "203.0.113.100"
        }
        raw_data = "Botnet communication detected"
        risk_points = 35
    },
    @{
        detection_id = 2
        risk_object = @{
            entity_type = "ip"
            entity_value = "203.0.113.100"
        }
        raw_data = "Data exfiltration attempt"
        risk_points = 30
    }
)

Write-Host "Sending events to generate risk alerts..."

foreach ($event in $events) {
    $json = $event | ConvertTo-Json -Depth 3
    Write-Host "Sending event: $($event.raw_data)"
    
    try {
        $response = Invoke-WebRequest -Uri "$baseUrl/api/events" -Method POST -Body $json -ContentType "application/json"
        Write-Host "  Status: $($response.StatusCode)"
    }
    catch {
        Write-Host "  Error: $($_.Exception.Message)"
    }
    
    Start-Sleep -Milliseconds 500
}

Write-Host "`nChecking for generated alerts..."
Start-Sleep -Seconds 2

try {
    $alertsResponse = Invoke-WebRequest -Uri "$baseUrl/api/risk/alerts" -Method GET
    $alerts = $alertsResponse.Content | ConvertFrom-Json
    
    if ($alerts -and $alerts.Count -gt 0) {
        Write-Host "SUCCESS: $($alerts.Count) risk alerts generated!"
        foreach ($alert in $alerts) {
            Write-Host "  Alert ID: $($alert.id), Entity ID: $($alert.entity_id), Score: $($alert.total_score)"
        }
    } else {
        Write-Host "No alerts generated yet. Checking risk object scores..."
        
        $objectsResponse = Invoke-WebRequest -Uri "$baseUrl/api/risk/objects" -Method GET
        $objects = $objectsResponse.Content | ConvertFrom-Json
        
        foreach ($obj in $objects) {
            Write-Host "  $($obj.entity_type)/$($obj.entity_value): Score $($obj.current_score)"
        }
    }
}
catch {
    Write-Host "Error checking alerts: $($_.Exception.Message)"
}