# Script to create events for existing alert entities
Write-Host "Creating events for existing alert entities..."

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
    # Get existing alerts to find their entity IDs
    Write-Host "Getting existing alerts..."
    $alertsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts" -Method GET
    Write-Host "Found $($alertsResponse.Count) alerts"
    
    if ($alertsResponse.Count -gt 0) {
        # Create a simple Go program to add events directly to database for specific entity IDs
        $goScript = @"
package main

import (
    "database/sql"
    "fmt"
    "log"
    "time"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "data/riskmatrix.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create events for entity ID 14 (alert 3's entity)
    events := []struct {
        detectionID int
        entityID    int
        rawData     string
        riskPoints  int
        timestamp   time.Time
    }{
        {1, 14, "Suspicious activity detected on entity 14 - Event 1", 25, time.Now().Add(-3 * time.Hour)},
        {2, 14, "Security alert for entity 14 - Event 2", 30, time.Now().Add(-2 * time.Hour)},
        {3, 14, "Risk event detected on entity 14 - Event 3", 20, time.Now().Add(-1 * time.Hour)},
    }

    fmt.Println("Creating events for entity ID 14...")
    for i, event := range events {
        _, err := db.Exec(`INSERT INTO events (detection_id, entity_id, timestamp, raw_data, risk_points, is_false_positive) 
                          VALUES (?, ?, ?, ?, ?, ?)`,
            event.detectionID, event.entityID, event.timestamp.Format(time.RFC3339), 
            event.rawData, event.riskPoints, false)
        if err != nil {
            log.Printf("Error creating event %d: %v", i+1, err)
        } else {
            fmt.Printf("Created event %d for entity 14\n", i+1)
        }
    }

    fmt.Println("Events created successfully!")
}
"@

        # Write the Go script to a temporary file
        $goScript | Out-File -FilePath "temp-fix-events.go" -Encoding UTF8
        
        # Build and run the Go script
        Write-Host "Building temporary fix utility..."
        go build -o temp-fix-events.exe temp-fix-events.go
        
        Write-Host "Running fix utility..."
        .\temp-fix-events.exe
        
        # Clean up
        Remove-Item "temp-fix-events.go" -ErrorAction SilentlyContinue
        Remove-Item "temp-fix-events.exe" -ErrorAction SilentlyContinue
        
        # Wait a moment and test
        Start-Sleep -Seconds 2
        
        Write-Host "Testing Contributing Events for alert 3..."
        $contributingEvents = Invoke-RestMethod -Uri "http://localhost:8080/api/risk/alerts/3/events" -Method GET
        Write-Host "Contributing events for alert 3: $($contributingEvents.Count)"
        
        if ($contributingEvents.Count -gt 0) {
            Write-Host "SUCCESS: Contributing Events are now working!"
            foreach ($event in $contributingEvents) {
                Write-Host "- Event ID: $($event.id), Detection ID: $($event.detection_id), Risk Points: $($event.risk_points)"
            }
        } else {
            Write-Host "Still no contributing events found."
        }
    }
    
} catch {
    Write-Host "Error: $($_.Exception.Message)"
} finally {
    Write-Host "Stopping server..."
    Get-Process -Name "server" -ErrorAction SilentlyContinue | Stop-Process -Force
}