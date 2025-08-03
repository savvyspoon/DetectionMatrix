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

	fmt.Println("Checking detection-MITRE mappings...")

	// Check production detections with MITRE mappings
	query := `
		SELECT d.id, d.name, d.status, dmm.mitre_id
		FROM detections d
		LEFT JOIN detection_mitre_map dmm ON d.id = dmm.detection_id
		WHERE d.status = 'production'
		ORDER BY d.id
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Error querying mappings:", err)
	}
	defer rows.Close()

	fmt.Println("Production detections and their MITRE mappings:")
	fmt.Println("ID | Name | Status | MITRE ID")
	fmt.Println("---|------|--------|----------")

	for rows.Next() {
		var id int
		var name, status string
		var mitreID sql.NullString

		err := rows.Scan(&id, &name, &status, &mitreID)
		if err != nil {
			log.Fatal("Error scanning row:", err)
		}

		mitreStr := "None"
		if mitreID.Valid {
			mitreStr = mitreID.String
		}

		fmt.Printf("%d | %s | %s | %s\n", id, name, status, mitreStr)
	}

	// Count unique techniques covered by production detections
	countQuery := `
		SELECT COUNT(DISTINCT dmm.mitre_id) as covered_techniques
		FROM detection_mitre_map dmm
		JOIN detections d ON dmm.detection_id = d.id
		WHERE d.status = 'production'
	`

	var coveredTechniques int
	err = db.QueryRow(countQuery).Scan(&coveredTechniques)
	if err != nil {
		log.Fatal("Error counting covered techniques:", err)
	}

	fmt.Printf("\nUnique techniques covered by production detections: %d\n", coveredTechniques)
}