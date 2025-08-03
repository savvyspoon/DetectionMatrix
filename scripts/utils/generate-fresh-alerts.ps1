# PowerShell script to generate events for NEW entities to trigger risk alerts

$baseUrl = "http://localhost:8080"

# Define test events for NEW entities that will cross the 50-point threshold
$events = @(
    # First entity: bob.hacker - will accumulate 60 points total
    @{
        detection_id = 1
        risk_object = @{
            entity_type = "user"
            entity_value = "bob.hacker"
        }
        raw_data = "Suspicious login from unusual location"
        risk_points = 20
    },
    @{
        detection_id = 2
        risk_object = @{
            entity_type = "user"
            entity_value = "bob.hacker"
        }
        raw_data = "Multiple failed authentication attempts"
        risk_points = 15
    },
    @{
        detection_id = 3
        risk_object = @{
            entity_type = "user"
            entity_value = "bob.hacker"
        }
        raw_data = "Privilege escalation attempt detected"
        risk_points = 25
    },
    
    # Second entity: malware-host - will accumulate 70 points total
    @{
        detection_id = 1
        risk_object = @{
            entity_type = "host"
            entity_value = "malware-host"
        }
        raw_data = "Malicious process execution detected"
        risk_points = 30
    },
    @{
        detection_id = 2
        risk_object = @{
            entity_type = "host"
            entity_value = "malware-host"
        }
        raw_data = "Suspicious network connections"
        risk_points = 25
    },
    @{
        detection_id = 3
        risk_object = @{
            entity_type = "host"
            entity_value = "malware-host"
        }
        raw_data = "File encryption activity detected"
        risk_points = 15
    },
    
    # Third entity: attacker-ip - will accumulate 80 points total
    @{
        detection_id = 1
        risk_object = @{
            entity_type = "ip"
            entity_value = "198.51.100.50"
        }
        raw_data = "Port scanning activity detected"
        risk_points = 25
    },
    @{
        detection_id = 2
        risk_object = @{
            entity_type = "ip"
            entity_value = "198.51.100.50"
        }
        raw_data = "Brute force attack detected"
        risk_points = 30
    },
    @{
        detection_id = 3
        risk_object = @{
            entity_type = "ip"
            entity_value = "198.51.100.50"
        }
        raw_data = "Data exfiltration attempt"
        risk_points = 25
    }
)

Write-Host "Sending events for NEW entities to generate risk alerts..."
Write-Host "Threshold is 50 points - these entities will cross it and trigger alerts."

foreach ($event in $events) {
    $json = $event | ConvertTo-Json -Depth 3
    Write-Host "Sending event for $($event.risk_object.entity_value): $($event.raw_data) (+$($event.risk_points) points)"
    
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
            Write-Host "  Alert ID: $($alert.id), Entity ID: $($alert.entity_id), Score: $($alert.total_score), Triggered: $($alert.triggered_at)"
        }
    } else {
        Write-Host "No alerts generated. Checking risk object scores..."
        
        $objectsResponse = Invoke-WebRequest -Uri "$baseUrl/api/risk/objects" -Method GET
        $objects = $objectsResponse.Content | ConvertFrom-Json
        
        Write-Host "Current risk object scores:"
        foreach ($obj in $objects) {
            Write-Host "  $($obj.entity_type)/$($obj.entity_value): Score $($obj.current_score)"
        }
    }
}
catch {
    Write-Host "Error checking alerts: $($_.Exception.Message)"
}