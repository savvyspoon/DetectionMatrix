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

	fmt.Println("Adding query column to detections table...")

	// Add query column to detections table
	_, err = db.Exec(`ALTER TABLE detections ADD COLUMN query TEXT`)
	if err != nil {
		// Check if column already exists
		if err.Error() == "duplicate column name: query" {
			fmt.Println("Query column already exists")
		} else {
			log.Printf("Error adding query column: %v", err)
			return
		}
	} else {
		fmt.Println("Successfully added query column to detections table")
	}

	// Verify the column was added
	fmt.Println("Verifying query column...")
	_, err = db.Query("SELECT query FROM detections LIMIT 1")
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
	} else {
		fmt.Println("Query column verified successfully")
	}

	fmt.Println("Schema update completed!")
}