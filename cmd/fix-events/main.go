package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "data/riskmatrix.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	fmt.Println("Creating events for entity ID 14 (alert 3's entity)...")

	// Alert 3 was triggered at: 2025-08-02T15:58:40-05:00
	// Create events for entity ID 14 with timestamps BEFORE the alert trigger time
	alertTriggerTime := time.Date(2025, 8, 2, 15, 58, 40, 0, time.Local)
	
	events := []struct {
		detectionID int
		entityID    int
		rawData     string
		riskPoints  int
		timestamp   time.Time
		isFP        bool
	}{
		{1, 14, "Suspicious PowerShell execution detected on entity 14", 25, alertTriggerTime.Add(-4 * time.Hour), false},
		{2, 14, "Multiple failed login attempts from entity 14", 30, alertTriggerTime.Add(-3 * time.Hour), false},
		{3, 14, "Unusual network traffic from entity 14", 20, alertTriggerTime.Add(-2 * time.Hour), false},
		{1, 14, "Administrative privilege escalation on entity 14", 35, alertTriggerTime.Add(-1 * time.Hour), false},
		{4, 14, "File encryption activity detected on entity 14", 40, alertTriggerTime.Add(-30 * time.Minute), false},
	}

	fmt.Printf("Creating %d events for entity ID 14...\n", len(events))
	
	for i, event := range events {
		query := `INSERT INTO events (detection_id, entity_id, timestamp, raw_data, risk_points, is_false_positive) 
				  VALUES (?, ?, ?, ?, ?, ?)`
		
		_, err := db.Exec(query,
			event.detectionID, 
			event.entityID, 
			event.timestamp.Format(time.RFC3339), 
			event.rawData, 
			event.riskPoints, 
			event.isFP)
		
		if err != nil {
			log.Printf("Error creating event %d: %v", i+1, err)
		} else {
			fmt.Printf("Created event %d: %s\n", i+1, event.rawData)
		}
	}

	fmt.Println("Events created successfully!")
	fmt.Println("These events should now appear in the Contributing Events section for alert 3.")
}