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

	// SQL commands to add missing columns
	commands := []string{
		"ALTER TABLE detections ADD COLUMN owner TEXT;",
		"ALTER TABLE detections ADD COLUMN risk_object TEXT CHECK (risk_object IN ('IP', 'Host', 'User'));",
		"ALTER TABLE detections ADD COLUMN testing_description TEXT;",
		"ALTER TABLE detections ADD COLUMN event_count_last_30_days INTEGER NOT NULL DEFAULT 0;",
		"ALTER TABLE detections ADD COLUMN false_positives_last_30_days INTEGER NOT NULL DEFAULT 0;",
	}

	// Execute each command
	for i, cmd := range commands {
		fmt.Printf("Executing command %d: %s\n", i+1, cmd)
		_, err := db.Exec(cmd)
		if err != nil {
			// Check if column already exists (ignore this error)
			if err.Error() == "duplicate column name: owner" ||
			   err.Error() == "duplicate column name: risk_object" ||
			   err.Error() == "duplicate column name: testing_description" ||
			   err.Error() == "duplicate column name: event_count_last_30_days" ||
			   err.Error() == "duplicate column name: false_positives_last_30_days" {
				fmt.Printf("Column already exists, skipping: %v\n", err)
				continue
			}
			log.Printf("Error executing command %d: %v\n", i+1, err)
		} else {
			fmt.Printf("Command %d executed successfully\n", i+1)
		}
	}

	fmt.Println("Database fix completed!")
}