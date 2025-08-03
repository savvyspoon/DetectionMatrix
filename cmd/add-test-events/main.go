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

	fmt.Println("Adding test risk objects and events...")

	// Create test risk objects first
	riskObjects := []struct {
		entityType  string
		entityValue string
		score       int
	}{
		{"user", "john.doe", 15},
		{"user", "jane.smith", 25},
		{"host", "workstation-01", 30},
		{"host", "server-db-01", 45},
		{"ip", "192.168.1.100", 20},
		{"ip", "10.0.0.50", 35},
	}

	fmt.Println("Creating risk objects...")
	for _, ro := range riskObjects {
		_, err := db.Exec(`INSERT INTO risk_objects (entity_type, entity_value, current_score, last_seen) 
						  VALUES (?, ?, ?, ?)`,
			ro.entityType, ro.entityValue, ro.score, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Printf("Error creating risk object %s/%s: %v", ro.entityType, ro.entityValue, err)
		} else {
			fmt.Printf("Created risk object: %s/%s\n", ro.entityType, ro.entityValue)
		}
	}

	// Create test events
	events := []struct {
		detectionID int
		entityID    int
		rawData     string
		riskPoints  int
		timestamp   time.Time
		isFP        bool
	}{
		{1, 1, "Suspicious PowerShell execution detected on john.doe workstation", 10, time.Now().Add(-2 * time.Hour), false},
		{2, 2, "Multiple failed login attempts from jane.smith", 15, time.Now().Add(-4 * time.Hour), false},
		{3, 3, "Unusual network traffic from workstation-01", 20, time.Now().Add(-6 * time.Hour), false},
		{1, 4, "Administrative privilege escalation on server-db-01", 25, time.Now().Add(-8 * time.Hour), false},
		{4, 5, "Port scan detected from 192.168.1.100", 12, time.Now().Add(-10 * time.Hour), true},
		{2, 6, "Brute force attack from 10.0.0.50", 18, time.Now().Add(-12 * time.Hour), false},
		{3, 1, "File encryption activity on john.doe machine", 30, time.Now().Add(-14 * time.Hour), false},
		{5, 2, "Suspicious DNS queries from jane.smith", 8, time.Now().Add(-16 * time.Hour), true},
		{1, 3, "Registry modification on workstation-01", 15, time.Now().Add(-18 * time.Hour), false},
		{4, 4, "Database access anomaly on server-db-01", 22, time.Now().Add(-20 * time.Hour), false},
		{2, 5, "Network reconnaissance from 192.168.1.100", 14, time.Now().Add(-22 * time.Hour), false},
		{3, 6, "Malware communication to 10.0.0.50", 28, time.Now().Add(-24 * time.Hour), false},
	}

	fmt.Println("Creating events...")
	for i, event := range events {
		_, err := db.Exec(`INSERT INTO events (detection_id, entity_id, timestamp, raw_data, risk_points, is_false_positive) 
						  VALUES (?, ?, ?, ?, ?, ?)`,
			event.detectionID, event.entityID, event.timestamp.Format(time.RFC3339), 
			event.rawData, event.riskPoints, event.isFP)
		if err != nil {
			log.Printf("Error creating event %d: %v", i+1, err)
		} else {
			description := event.rawData
			if len(description) > 50 {
				description = description[:50] + "..."
			}
			fmt.Printf("Created event %d: %s\n", i+1, description)
		}
	}

	// Update risk object scores based on events
	fmt.Println("Updating risk object scores...")
	_, err = db.Exec(`UPDATE risk_objects 
					  SET current_score = current_score + (
						  SELECT COALESCE(SUM(risk_points), 0) 
						  FROM events 
						  WHERE events.entity_id = risk_objects.id 
						  AND events.is_false_positive = 0
					  )`)
	if err != nil {
		log.Printf("Error updating risk scores: %v", err)
	} else {
		fmt.Println("Updated risk object scores based on events")
	}

	fmt.Println("Test data creation completed!")
}