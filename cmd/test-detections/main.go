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

	// Test basic query
	fmt.Println("Testing basic detection query...")
	query := `SELECT id, name, description, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days, created_at, updated_at FROM detections ORDER BY name`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Error querying detections:", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var id int64
		var name, description, status, severity, playbook_link, owner, risk_object, testing_description, created_at, updated_at string
		var risk_points, event_count, false_positives int
		
		err := rows.Scan(&id, &name, &description, &status, &severity, &risk_points, &playbook_link, &owner, &risk_object, &testing_description, &event_count, &false_positives, &created_at, &updated_at)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		
		fmt.Printf("Detection %d: %s (Status: %s)\n", id, name, status)
	}

	fmt.Printf("Total detections found: %d\n", count)
	
	if count == 0 {
		fmt.Println("No detections found in database. This explains why the API returns empty results.")
	}
}