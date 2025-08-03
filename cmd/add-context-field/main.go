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

	fmt.Println("Adding context field to events table...")

	// Add context column to events table
	_, err = db.Exec("ALTER TABLE events ADD COLUMN context TEXT;")
	if err != nil {
		// Check if column already exists (ignore this error)
		if err.Error() == "duplicate column name: context" {
			fmt.Println("Context column already exists, skipping...")
		} else {
			log.Printf("Error adding context column: %v", err)
		}
	} else {
		fmt.Println("Context column added successfully!")
	}

	fmt.Println("Database migration completed!")
}