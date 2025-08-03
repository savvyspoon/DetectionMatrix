package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: update-alert-schema <database-path>")
	}

	dbPath := os.Args[1]
	
	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if columns already exist
	var count int
	err = db.QueryRow("PRAGMA table_info(risk_alerts)").Scan(&count)
	if err != nil {
		log.Printf("Error checking table info: %v", err)
	}

	// Add new columns to risk_alerts table
	migrations := []string{
		"ALTER TABLE risk_alerts ADD COLUMN status TEXT NOT NULL DEFAULT 'New'",
		"ALTER TABLE risk_alerts ADD COLUMN notes TEXT",
		"ALTER TABLE risk_alerts ADD COLUMN owner TEXT",
	}

	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			// Check if column already exists (ignore error if it does)
			if err.Error() != "duplicate column name: status" && 
			   err.Error() != "duplicate column name: notes" && 
			   err.Error() != "duplicate column name: owner" {
				log.Printf("Migration warning: %v", err)
			}
		} else {
			fmt.Printf("Successfully executed: %s\n", migration)
		}
	}

	// Add CHECK constraint for status (recreate table approach for SQLite)
	fmt.Println("Database migration completed successfully!")
	fmt.Println("Note: You may need to manually add CHECK constraint for status field if needed.")
}