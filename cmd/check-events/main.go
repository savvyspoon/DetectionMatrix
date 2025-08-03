package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "data/riskmatrix.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	fmt.Println("Checking events for detection 8...")

	// Check all events for detection 8
	query := `SELECT id, detection_id, timestamp, is_false_positive, risk_points 
	          FROM events 
	          WHERE detection_id = 8 
	          ORDER BY timestamp DESC`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Error querying events:", err)
	}
	defer rows.Close()

	fmt.Println("ID | Detection ID | Timestamp | False Positive | Risk Points")
	fmt.Println("---|-------------|-----------|----------------|------------")

	count := 0
	fpCount := 0
	for rows.Next() {
		var id, detectionID, riskPoints int
		var timestamp string
		var isFalsePositive bool

		err := rows.Scan(&id, &detectionID, &timestamp, &isFalsePositive, &riskPoints)
		if err != nil {
			log.Fatal("Error scanning row:", err)
		}

		fpStr := "No"
		if isFalsePositive {
			fpStr = "Yes"
			fpCount++
		}

		fmt.Printf("%d | %d | %s | %s | %d\n", id, detectionID, timestamp, fpStr, riskPoints)
		count++
	}

	fmt.Printf("\nTotal events for detection 8: %d\n", count)
	fmt.Printf("False positives: %d\n", fpCount)

	// Test the 30-day query
	fmt.Println("\nTesting 30-day query...")
	query30 := `SELECT COUNT(*) FROM events 
	            WHERE detection_id = 8 
	            AND timestamp >= datetime('now', '-30 days')`

	var count30 int
	err = db.QueryRow(query30).Scan(&count30)
	if err != nil {
		log.Fatal("Error with 30-day query:", err)
	}

	fmt.Printf("Events in last 30 days: %d\n", count30)

	// Test false positive 30-day query
	queryFP30 := `SELECT COUNT(*) FROM events 
	              WHERE detection_id = 8 
	              AND is_false_positive = 1 
	              AND timestamp >= datetime('now', '-30 days')`

	var fpCount30 int
	err = db.QueryRow(queryFP30).Scan(&fpCount30)
	if err != nil {
		log.Fatal("Error with FP 30-day query:", err)
	}

	fmt.Printf("False positives in last 30 days: %d\n", fpCount30)
}